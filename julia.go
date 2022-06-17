package main

import (
	"image"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"go-trip/assets"
)

type JuliaProgram struct {
	Window *glfw.Window

	// state
	iterations int32
	zoom       float64
	cx, cy     float64
	ox, oy     float64

	complexMode bool
	zoomFactor  float64

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

func NewJuliaProgram() Program {
	return &JuliaProgram{
		iterations: 255,
		cx:         -0.7,
		cy:         0.27015,
		ox:         0.0,
		oy:         0.0,
		zoom:       1.0,

		complexMode: true,
		zoomFactor:  0.01,

		mouseDelta: NewMouseDelta(.0001),
	}
}

func (self *JuliaProgram) Load(window *glfw.Window, vao, vbo uint32) {
	self.Window = window
	self.vao = vao
	self.vbo = vbo
	width, height := window.GetSize()

	img := *image.NewRGBA(image.Rect(0, 0, width, height))

	// create compute textures
	self.fractalTexture = LoadTexture(&img)

	// create compute shaders
	self.fractalShader = MustCompileShader(assets.VertexShader, assets.JuliaShader)

	// create output shaders
	self.outputShaders = *NewCyclicArray([]Shader{
		MustCompileShader(assets.VertexShader, assets.ViridisShader),
		MustCompileShader(assets.VertexShader, assets.InfernoShader),
		MustCompileShader(assets.VertexShader, assets.MagmaShader),
		MustCompileShader(assets.VertexShader, assets.PlasmaShader),
		MustCompileShader(assets.VertexShader, assets.CividisShader),
		MustCompileShader(assets.VertexShader, assets.TurboShader),
		MustCompileShader(assets.VertexShader, assets.SinebowShader),
		MustCompileShader(assets.VertexShader, assets.RGBShader),
	})

	// create framebuffers
	gl.GenFramebuffers(1, &self.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)

	self.Window.SetScrollCallback(self.ScrollCallback)
	self.Window.SetCursorPosCallback(self.CursorPosCallback)
}

func (self *JuliaProgram) Render(t float64) {
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
		Uniform2f("focus", float32(self.cx), float32(self.cy)).
		Uniform2f("offset", float32(self.ox), float32(self.oy)).
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

func (self *JuliaProgram) ZoomOut() {
	self.zoom -= self.zoomFactor
}

func (self *JuliaProgram) ZoomIn() {
	self.zoom += self.zoomFactor
}

func (self *JuliaProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.fractalTexture.Resize(width, height)
}

func (self *JuliaProgram) CursorPosCallback(w *glfw.Window, x, y float64) {
	// pan screen
	dx, dy := self.mouseDelta.Delta(x, y)

	if self.complexMode {
		self.cx += dx / self.zoom
		self.cy -= dy / self.zoom
	} else {
		self.ox += dx / self.zoom
		self.oy -= dy / self.zoom
	}
}

func (self *JuliaProgram) ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	if yoff > 0 {
		self.ZoomIn()
		self.zoom += self.zoomFactor
	} else if yoff < 0 {
		self.ZoomOut()
	}
}

func (self *JuliaProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
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

		if key == glfw.KeySpace {
			self.complexMode = !self.complexMode
		}

		if key == glfw.KeyR {
			self.ox = 0.0
			self.oy = 0.0
			self.zoom = 1.0
			self.iterations = 255
			self.cx = -0.7
			self.cy = 0.27015
		}
	}
}
