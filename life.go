package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"go-trip/assets"
)

var RecolorCmd = "recolor"

type LifeMode int

const (
	LifeStd LifeMode = iota
	LifeCyclic
)

type LifeProgram struct {
	Window *glfw.Window

	// rule set
	survive []int32
	birth   []int32

	// state
	frame      int32
	paused     bool
	mode       LifeMode
	cursorSize float64
	cmds       CmdChannels

	// textures
	prevTexture        *Texture
	nextTexture        *Texture
	growthDecayTexture *Texture

	// compute shaders
	lifeShader        Shader
	cyclicShader      Shader
	growthDecayShader Shader

	// output shaders
	outputShaders cyclicArray[Shader]
	gradientIndex cyclicArray[int32]

	// buffers
	fbo, vao, vbo uint32
}

func NewLifeProgram() Program {
	survive := []int32{-1, -1, 2, 3, -1, -1, -1, -1, -1}
	birth := []int32{-1, -1, -1, 3, -1, -1, -1, -1, -1}

	cmds := NewCmdChannels()
	cmds.Register(RecolorCmd)

	return &LifeProgram{
		survive: survive,
		birth:   birth,

		frame:      0,
		paused:     false,
		mode:       LifeStd,
		cursorSize: 0.025,

		cmds:          cmds,
		gradientIndex: *newCyclicArray([]int32{0, 1, 2, 3}),
	}
}

func (self *LifeProgram) Load(window *glfw.Window, vao, vbo uint32) {
	self.Window = window
	self.vao = vao
	self.vbo = vbo
	width, height := window.GetSize()

	// create textures
	img1 := *image.NewRGBA(image.Rect(0, 0, width, height))
	img2 := *image.NewRGBA(image.Rect(0, 0, width, height))
	img3 := *image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < img1.Rect.Max.X; x++ {
		for y := 0; y < img1.Rect.Max.Y; y++ {
			r := uint8(rand.Intn(255))
			g := uint8(rand.Intn(255))
			b := uint8(rand.Intn(255))
			a := uint8(rand.Intn(255))

			c := color.RGBA{r, g, b, a}
			img1.Set(x, y, c)
			img2.Set(x, y, c)
			img3.Set(x, y, color.White)
		}
	}

	// create compute textures
	self.prevTexture = LoadTexture(&img1)
	self.nextTexture = LoadTexture(&img2)
	self.growthDecayTexture = LoadTexture(&img3)

	// create compute shaders
	self.cyclicShader = MustCompileShader(assets.VertexShader, assets.CyclicShader)
	self.lifeShader = MustCompileShader(assets.VertexShader, assets.GOLShader)
	self.growthDecayShader = MustCompileShader(assets.VertexShader, assets.GainShader)

	// create output shaders
	self.outputShaders = *newCyclicArray([]Shader{
		MustCompileShader(assets.VertexShader, assets.ViridisShader),
		MustCompileShader(assets.VertexShader, assets.InfernoShader),
		MustCompileShader(assets.VertexShader, assets.MagmaShader),
		MustCompileShader(assets.VertexShader, assets.PlasmaShader),
		MustCompileShader(assets.VertexShader, assets.CividisShader),
		MustCompileShader(assets.VertexShader, assets.TurboShader),
		MustCompileShader(assets.VertexShader, assets.SinebowShader),
		MustCompileShader(assets.VertexShader, assets.RGBShader),
		MustCompileShader(assets.VertexShader, assets.RGBAShader),
	})

	// create framebuffers
	gl.GenFramebuffers(1, &self.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)

	self.Window.SetScrollCallback(self.ScrollCallback)
}

func (self *LifeProgram) recolor() {
	width, height := self.Window.GetSize()

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)

	switch self.mode {
	case LifeStd:
		self.growthDecayTexture.Activate(gl.TEXTURE0)
	case LifeCyclic:
		self.prevTexture.Activate(gl.TEXTURE0)
	}

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

}

func (self *LifeProgram) life(t float64) {
	width, height := self.Window.GetSize()
	mx, my := self.Window.GetCursorPos()

	// use gol program
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.nextTexture.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.prevTexture.Activate(gl.TEXTURE0)

	self.lifeShader.Use().
		Uniform1iv("s", self.survive).
		Uniform1iv("b", self.birth).
		Uniform1i("state", 0).
		Uniform1f("cursorSize", float32(self.cursorSize)).
		Uniform1f("time", float32(t)).
		Uniform2f("scale", float32(width), float32(height)).
		Uniform2f("mouse", float32(mx), float32(height)-float32(my))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// swap texture
	self.prevTexture, self.nextTexture = self.nextTexture, self.prevTexture

	// use decay program
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.growthDecayTexture.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.prevTexture.Activate(gl.TEXTURE0)
	self.growthDecayTexture.Activate(gl.TEXTURE1)

	self.growthDecayShader.Use().
		Uniform1i("state", 0).
		Uniform1i("self", 1).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)
	self.growthDecayTexture.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (self *LifeProgram) cyclic(t float64) {
	width, height := self.Window.GetSize()
	mx, my := self.Window.GetCursorPos()

	// use cyclic life program
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.nextTexture.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.prevTexture.Activate(gl.TEXTURE0)

	self.cyclicShader.Use().
		Uniform1f("stages", 16.0).
		Uniform1i("state", 0).
		Uniform1f("cursorSize", float32(self.cursorSize)).
		Uniform1f("time", float32(t)).
		Uniform2f("scale", float32(width), float32(height)).
		Uniform2f("mouse", float32(mx), float32(height)-float32(my))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// swap texture
	self.prevTexture, self.nextTexture = self.nextTexture, self.prevTexture

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)
	self.prevTexture.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (self *LifeProgram) Render(t float64) {
	select {
	case <-self.cmds[RecolorCmd]:
		self.recolor()
	default:
		if self.paused {
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
			return
		}

		self.frame = self.frame + 1
		switch self.mode {
		case LifeStd:
			self.life(t)
		case LifeCyclic:
			self.cyclic(t)
		}
	}
}

func (self *LifeProgram) ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	if yoff > 0 {
		self.cursorSize = math.Max(0, self.cursorSize-0.005)
	} else if yoff < 0 {
		self.cursorSize = math.Min(1, self.cursorSize+0.005)
	}
}

func (self *LifeProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.prevTexture.Resize(width, height)
	self.nextTexture.Resize(width, height)
	self.growthDecayTexture.Resize(width, height)
}

func (self *LifeProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Release {
		if key == glfw.Key1 {
			self.mode = LifeStd
		}

		if key == glfw.Key2 {
			self.mode = LifeCyclic
		}

		if key == glfw.KeyJ {
			self.outputShaders.Previous()
			self.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeyK {
			self.outputShaders.Next()
			self.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeyH {
			self.gradientIndex.Previous()
			self.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeyL {
			self.gradientIndex.Next()
			self.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeySpace {
			self.paused = !self.paused
		}
	}
}
