package main

import (
	"image"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	. "gogl"
	. "gogl/arrayutil"
	. "gogl/assets"

	_ "embed"
)

func init() {
	HotProgram = NewMandelbrotProgram()
}

func HotProgramFn(kill <-chan bool, window *glfw.Window) {
	HotRender(kill, window)
}

//go:embed mandelbrot.glsl
var MandelbrotShader string

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
	outputShaders CyclicArray[Shader]

	mouseDelta *MouseDelta

	// buffers
	fbo uint32
	bo  BufferObject
}

func NewMandelbrotProgram() Program {
	return &MandelbrotProgram{
		paused:     false,
		iterations: 1000,
		x:          -0.51,
		y:          0.0,
		zoom:       1.0,

		zoomFactor: 0.01,
	}
}

func (self *MandelbrotProgram) LoadR(r *Renderer) {
	self.Load(r.Window)
}

func (self *MandelbrotProgram) Load(window *glfw.Window) {
	self.Window = window
	self.mouseDelta = NewMouseDelta(self.Window, .0001)
	self.bo = NewV4Buffer(QuadVertices, 2, 4)
	width, height := window.GetFramebufferSize()

	img := *image.NewRGBA(image.Rect(0, 0, width, height))

	// create compute textures
	self.fractalTexture = LoadTexture(&img)

	// create compute shaders
	self.fractalShader = MustCompileShader(VertexShader, MandelbrotShader, self.bo)

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
	self.Window.SetCursorPosCallback(self.CursorPosCallback)
}

func (self *MandelbrotProgram) Render(t float64) {
	if self.paused {
		return
	}

	width, height := self.Window.GetFramebufferSize()

	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.fractalTexture.Handle, 0)

	gl.BindVertexArray(self.bo.VAO())
	self.fractalTexture.Activate(gl.TEXTURE0)

	self.fractalShader.Use().
		Uniform1i("maxIterations", self.iterations).
		Uniform2f("focus", float32(self.x), float32(self.y)).
		Uniform1f("zoom", float32(self.zoom)).
		Uniform2f("scale", float32(width), float32(height))
	self.bo.Draw()

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.bo.VAO())
	self.fractalTexture.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", 0).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	self.bo.Draw()
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
	self.y += dy / self.zoom
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
