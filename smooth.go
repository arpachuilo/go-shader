package main

import (
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type SmoothLifeProgram struct {
	Window *glfw.Window

	// state
	frame      int32
	paused     bool
	cursorSize float64
	cmds       CmdChannels

	// textures
	textureA *Texture
	textureB *Texture
	textureC *Texture

	// shaders
	smoothShader Shader
	gaussX       Shader
	gaussY       Shader

	// output shaders
	// outputShader Shader
	outputShaders CyclicArray[Shader]
	gradientIndex CyclicArray[int32]

	// buffers
	fbo, vao, vbo uint32
}

func NewSmoothLifeProgram() Program {
	cmds := NewCmdChannels()
	cmds.Register(RecolorCmd)

	return &SmoothLifeProgram{
		frame:      0,
		paused:     false,
		cursorSize: 0.025,

		cmds:          cmds,
		gradientIndex: *NewCyclicArray([]int32{0, 1, 2, 3}),
	}
}

func (self *SmoothLifeProgram) Load(window *glfw.Window, vao, vbo uint32) {
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
			// a := uint8(rand.Intn(255))

			c := color.RGBA{r, g, b, 255.0}
			img1.Set(x, y, c)
			img2.Set(x, y, color.Black)
			img3.Set(x, y, color.Black)
		}
	}

	// create compute textures
	self.textureA = LoadTexture(&img1)
	self.textureB = LoadTexture(&img2)
	self.textureC = LoadTexture(&img3)

	// create compute shaders
	self.smoothShader = MustCompileShader(vertexShader, smoothShader)
	self.gaussX = MustCompileShader(vertexShader, gaussXShader)
	self.gaussY = MustCompileShader(vertexShader, gaussYShader)

	// create output shaders
	self.outputShaders = *NewCyclicArray([]Shader{
		MustCompileShader(vertexShader, viridisShader),
		MustCompileShader(vertexShader, infernoShader),
		MustCompileShader(vertexShader, magmaShader),
		MustCompileShader(vertexShader, plasmaShader),
		MustCompileShader(vertexShader, cividisShader),
		MustCompileShader(vertexShader, turboShader),
		MustCompileShader(vertexShader, sinebowShader),
		MustCompileShader(vertexShader, rgbShader),
		MustCompileShader(vertexShader, rgbaShader),
	})

	// self.outputShader = MustCompileShader(vertexShader, smoothOutputShader)

	// create framebuffers
	gl.GenFramebuffers(1, &self.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)

	self.Window.SetScrollCallback(self.ScrollCallback)
}

func (self *SmoothLifeProgram) recolor() {
	width, height := self.Window.GetSize()
	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)
	self.textureA.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (self *SmoothLifeProgram) smooth(t float64) {
	width, height := self.Window.GetSize()
	mx, my := self.Window.GetCursorPos()
	mb1 := self.Window.GetMouseButton(glfw.MouseButton1)
	mb2 := self.Window.GetMouseButton(glfw.MouseButton2)

	// use smooth life program
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.textureA.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.textureA.Activate(gl.TEXTURE0)
	self.textureC.Activate(gl.TEXTURE1)

	self.smoothShader.Use().
		Uniform1i("inputA", 0).
		Uniform1i("inputC", 1).
		Uniform1i("frame", self.frame).
		Uniform1f("cursorSize", float32(self.cursorSize)).
		Uniform1f("time", float32(t)).
		Uniform2f("scale", float32(width), float32(height)).
		Uniform4f("mouse", float32(mx), float32(height)-float32(my), float32(mb1), float32(mb2))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// use gauss x
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.textureB.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.textureA.Activate(gl.TEXTURE0)

	self.gaussX.Use().
		Uniform1i("inputA", 0).
		Uniform1i("frame", self.frame).
		Uniform1f("cursorSize", float32(self.cursorSize)).
		Uniform1f("time", float32(t)).
		Uniform2f("scale", float32(width), float32(height)).
		Uniform2f("mouse", float32(mx), float32(height)-float32(my))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// use gauss y
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.textureC.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.textureB.Activate(gl.TEXTURE0)

	self.gaussY.Use().
		Uniform1i("inputB", 0).
		Uniform1i("frame", self.frame).
		Uniform1f("cursorSize", float32(self.cursorSize)).
		Uniform1f("time", float32(t)).
		Uniform2f("scale", float32(width), float32(height)).
		Uniform2f("mouse", float32(mx), float32(height)-float32(my))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)
	self.textureA.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (self *SmoothLifeProgram) Render(t float64) {
	select {
	case <-self.cmds[RecolorCmd]:
		self.recolor()
	default:
		if self.paused {
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
			return
		}

		self.frame = self.frame + 1
		self.smooth(t)
	}
}

func (self *SmoothLifeProgram) ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	if yoff > 0 {
		self.cursorSize = math.Max(0, self.cursorSize-0.005)
	} else if yoff < 0 {
		self.cursorSize = math.Min(1, self.cursorSize+0.005)
	}
}

func (self *SmoothLifeProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.textureA.Resize(width, height)
	self.textureB.Resize(width, height)
	self.textureC.Resize(width, height)
}

func (self *SmoothLifeProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Release {
		if key == glfw.KeySpace {
			self.paused = !self.paused
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
	}
}
