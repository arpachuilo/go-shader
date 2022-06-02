package main

import (
	"image"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type MandelbrotProgram struct {
	Window *glfw.Window

	// state
	iterations int32
	zoom       float64
	x, y       float64

	zoomFactor float64

	// textures
	fractalTexture *Texture

	// compute shaders
	fractalShader Shader

	// output shaders
	outputShaders CyclicArray[Shader]

	mouseDelta *MouseDelta

	// buffers
	fbo, vao, vbo uint32
}

func NewMandelbrotProgram() Program {
	return &MandelbrotProgram{
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
	width, height := window.GetSize()

	img := *image.NewRGBA(image.Rect(0, 0, width, height))

	// create compute textures
	self.fractalTexture = LoadTexture(&img)

	// create compute shaders
	self.fractalShader = MustCompileShader(vertexShader, mandelbrotShader)

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
	})

	// create framebuffers
	gl.GenFramebuffers(1, &self.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)

	self.Window.SetScrollCallback(self.ScrollCallback)
	self.Window.SetCursorPosCallback(self.CursorPosCallback)
}

func (self *MandelbrotProgram) Render(t float64) {
	width, height := self.Window.GetSize()
	sy, sx := self.Window.GetContentScale()
	gl.Viewport(
		0, 0,
		int32(float32(width)*sx), int32(float32(height)*sy),
	)

	// use gol program
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.fractalTexture.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.fractalTexture.Activate(gl.TEXTURE0)

	self.fractalShader.Use().
		Uniform1i("maxIterations", self.iterations).
		Uniform2f("focus", float32(self.x), float32(self.y)).
		Uniform1f("zoom", float32(self.zoom)).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)
	self.fractalTexture.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", 0).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
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
