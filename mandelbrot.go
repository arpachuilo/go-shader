package main

import (
	"image"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type MandelbrotProgram struct {
	Window *glfw.Window

	// state
	paused     bool
	iterations int32
	zoom       float64
	x, y       float64

	zoomFactor float64

	// textures
	fractalTexture *Texture

	// compute shaders
	fractalShader Shader

	// output shaders
	outputShaders cyclicArray[Shader]

	mouseDelta *MouseDelta

	// buffers
	fbo, vao, vbo uint32
}

func NewMandelbrotProgram() Program {
	return &MandelbrotProgram{
		paused:     false,
		iterations: 1000,
		x:          -0.51,
		y:          0.0,
		zoom:       1.0,

		zoomFactor: 0.01,

		mouseDelta: NewMouseDelta(.0001),
	}
}

func (self *MandelbrotProgram) Load(window *glfw.Window, vao, vbo uint32) {
	self.Window = window
	self.vao = vao
	self.vbo = vbo
	width, height := window.GetFramebufferSize()

	img := *image.NewRGBA(image.Rect(0, 0, width, height))

	// create compute textures
	self.fractalTexture = LoadTexture(&img)

	// create compute shaders
	self.fractalShader = MustCompileShader(VertexShader, MandelbrotShader)

	// create output shaders
	self.outputShaders = *newCyclicArray([]Shader{
		MustCompileShader(VertexShader, ViridisShader),
		MustCompileShader(VertexShader, InfernoShader),
		MustCompileShader(VertexShader, MagmaShader),
		MustCompileShader(VertexShader, PlasmaShader),
		MustCompileShader(VertexShader, CividisShader),
		MustCompileShader(VertexShader, TurboShader),
		MustCompileShader(VertexShader, SinebowShader),
		MustCompileShader(VertexShader, RGBShader),
		MustCompileShader(VertexShader, RGBAShader),
	})

	// create framebuffers
	gl.GenFramebuffers(1, &self.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)

	self.Window.SetScrollCallback(self.ScrollCallback)
	self.Window.SetCursorPosCallback(self.CursorPosCallback)
}

func (self *MandelbrotProgram) Render(t float64) {
	if self.paused {
		return
	}

	width, height := self.Window.GetFramebufferSize()

	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.fractalTexture.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.fractalTexture.Activate(gl.TEXTURE0)

	self.fractalShader.Use().
		Uniform1i("maxIterations", self.iterations).
		Uniform2f("focus", float32(self.x), float32(self.y)).
		Uniform1f("zoom", float32(self.zoom)).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)
	self.fractalTexture.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", 0).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, 4)
}

func (self *MandelbrotProgram) ZoomOut() {
	self.zoom -= self.zoomFactor
}

func (self *MandelbrotProgram) ZoomIn() {
	self.zoom += self.zoomFactor
}

func (self *MandelbrotProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.fractalTexture.Resize(width, height)
}

func (self *MandelbrotProgram) CursorPosCallback(w *glfw.Window, x, y float64) {
	// pan screen
	dx, dy := self.mouseDelta.Delta(x, y)
	self.x += dx / self.zoom
	self.y -= dy / self.zoom
}

func (self *MandelbrotProgram) ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	if yoff > 0 {
		self.ZoomIn()
		self.zoom += self.zoomFactor
	} else if yoff < 0 {
		self.ZoomOut()
	}
}

func (self *MandelbrotProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeySpace && action == glfw.Release {
		self.paused = !self.paused
	}

	if key == glfw.KeyEqual {
		self.iterations = self.iterations + 10
	}

	if key == glfw.KeyMinus {
		self.iterations = self.iterations - 10
		if self.iterations <= 0 {
			self.iterations = 1
		}
	}

	if action == glfw.Release {
		if key == glfw.KeyJ {
			self.outputShaders.Previous()
		}

		if key == glfw.KeyK {
			self.outputShaders.Next()
		}
	}
}
