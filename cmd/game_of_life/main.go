package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	. "gogl"
	. "gogl/arrayutil"
	. "gogl/assets"

	_ "embed"
)

var BuildDate = ""

func HotRender(kill <-chan bool, window *glfw.Window) {
	fmt.Println(BuildDate)
	program := NewLifeProgram()

	NewRenderer(window, program).Run(kill)
}

//go:embed cyclic.glsl
var CyclicShader string

//go:embed life.glsl
var GOLShader string

//go:embed growth_decay.glsl
var GainShader string

var RecolorCmd = "recolor"

type LifeMode int

const (
	LifeStd LifeMode = iota
	LifeCyclic
)

type LifeProgram struct {
	Window        *glfw.Window
	width, height int

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
	outputShaders CyclicArray[Shader]
	gradientIndex CyclicArray[int32]

	// buffers
	fbo uint32
	bo  BufferObject
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
		gradientIndex: *NewCyclicArray([]int32{0, 1, 2, 3}),
	}
}

func (self *LifeProgram) Load(window *glfw.Window) {
	self.Window = window
	self.bo = NewVBuffer(QuadVertices, 2, 4)
	self.width, self.height = window.GetFramebufferSize()

	// create textures
	img1 := *image.NewRGBA(image.Rect(0, 0, self.width, self.height))
	img2 := *image.NewRGBA(image.Rect(0, 0, self.width, self.height))
	img3 := *image.NewRGBA(image.Rect(0, 0, self.width, self.height))
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
	self.cyclicShader = MustCompileShader(VertexShader, CyclicShader, self.bo)
	self.lifeShader = MustCompileShader(VertexShader, GOLShader, self.bo)
	self.growthDecayShader = MustCompileShader(VertexShader, GainShader, self.bo)

	// create output shaders
	self.outputShaders = *NewCyclicArray([]Shader{
		MustCompileShader(VertexShader, ViridisShader, self.bo),
		MustCompileShader(VertexShader, InfernoShader, self.bo),
		MustCompileShader(VertexShader, MagmaShader, self.bo),
		MustCompileShader(VertexShader, PlasmaShader, self.bo),
		MustCompileShader(VertexShader, CividisShader, self.bo),
		MustCompileShader(VertexShader, TurboShader, self.bo),
		MustCompileShader(VertexShader, SinebowShader, self.bo),
		MustCompileShader(VertexShader, RGBShader, self.bo),
		MustCompileShader(VertexShader, RGBAShader, self.bo),
	})

	// create framebuffers
	gl.GenFramebuffers(1, &self.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)

	self.Window.SetScrollCallback(self.ScrollCallback)
}

func (self *LifeProgram) recolor() {
	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.bo.VAO())

	switch self.mode {
	case LifeStd:
		self.growthDecayTexture.Activate(gl.TEXTURE0)
	case LifeCyclic:
		self.prevTexture.Activate(gl.TEXTURE0)
	}

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(self.width), float32(self.height))
	self.bo.Draw()

}

func (self *LifeProgram) life(t float64) {
	mx, my := self.Window.GetCursorPos()

	gl.Clear(gl.COLOR_BUFFER_BIT)
	// use gol program
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.nextTexture.Handle, 0)

	gl.BindVertexArray(self.bo.VAO())
	self.prevTexture.Activate(gl.TEXTURE0)

	self.lifeShader.Use().
		Uniform1iv("s", self.survive).
		Uniform1iv("b", self.birth).
		Uniform1i("state", 0).
		Uniform1f("cursorSize", float32(self.cursorSize)).
		Uniform1i("u_frame", int32(self.frame)).
		Uniform1f("u_time", float32(t)).
		Uniform2f("u_mouse", float32(mx), float32(self.height)-float32(my)).
		Uniform2f("u_resolution", float32(self.width), float32(self.height))
	self.bo.Draw()

	// swap texture
	self.prevTexture, self.nextTexture = self.nextTexture, self.prevTexture

	// use decay program
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.growthDecayTexture.Handle, 0)

	gl.BindVertexArray(self.bo.VAO())
	self.prevTexture.Activate(gl.TEXTURE0)
	self.growthDecayTexture.Activate(gl.TEXTURE1)

	self.growthDecayShader.Use().
		Uniform1i("state", 0).
		Uniform1i("self", 1).
		Uniform1i("u_frame", int32(self.frame)).
		Uniform1f("u_time", float32(t)).
		Uniform2f("u_mouse", float32(mx), float32(self.height)-float32(my)).
		Uniform2f("u_resolution", float32(self.width), float32(self.height))
	self.bo.Draw()

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.bo.VAO())
	self.growthDecayTexture.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(self.width), float32(self.height)).
		Uniform1i("u_frame", int32(self.frame)).
		Uniform1f("u_time", float32(t)).
		Uniform2f("u_mouse", float32(mx), float32(self.height)-float32(my)).
		Uniform2f("u_resolution", float32(self.width), float32(self.height))
	self.bo.Draw()
}

func (self *LifeProgram) cyclic(t float64) {
	mx, my := self.Window.GetCursorPos()

	// use cyclic life program
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.nextTexture.Handle, 0)

	gl.BindVertexArray(self.bo.VAO())
	self.prevTexture.Activate(gl.TEXTURE0)

	self.cyclicShader.Use().
		Uniform1f("stages", 16.0).
		Uniform1i("state", 0).
		Uniform1f("cursorSize", float32(self.cursorSize)).
		Uniform1i("u_frame", int32(self.frame)).
		Uniform1f("u_time", float32(t)).
		Uniform2f("u_mouse", float32(mx), float32(self.height)-float32(my)).
		Uniform2f("u_resolution", float32(self.width), float32(self.height))
	self.bo.Draw()

	// swap texture
	self.prevTexture, self.nextTexture = self.nextTexture, self.prevTexture

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.bo.VAO())
	self.prevTexture.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(self.width), float32(self.height))
	self.bo.Draw()
}

func (self *LifeProgram) Render(t float64) {
	select {
	case <-self.cmds[RecolorCmd]:
		self.recolor()
	default:
		if self.paused {
			self.bo.Draw()
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
	self.width, self.height = self.Window.GetFramebufferSize()

	self.prevTexture.Resize(self.width, self.height)
	self.nextTexture.Resize(self.width, self.height)
	self.growthDecayTexture.Resize(self.width, self.height)
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
