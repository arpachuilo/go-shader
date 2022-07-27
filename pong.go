package main

import (
	"image"
	"image/color"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Pong struct {
	Heading  Vector2
	Position Vector2

	Speed float64
	Size  float64

	Width  int
	Height int
}

func NewPong() *Pong {
	return &Pong{
		Position: Vector2{
			X: 0,
			Y: 0,
		},
		Heading: Vector2{
			X: 0.5,
			Y: 0.3,
		}.Normalize(),
		Speed: 5,
		Size:  42,
	}
}

func (self *Pong) Resize(width, height int) {
	self.Width = width
	self.Height = height
}

func (self *Pong) Turn(degrees float64) Vector2 {
	self.Heading = self.Heading.Turn(degrees)
	return self.Heading
}

func (self *Pong) Advance() Vector2 {
	next := self.Position.Add(self.Heading.Mul(self.Speed))

	if next.X <= 0.0 {
		self.Heading = self.Heading.Reflect(Vector3Right).Normalize()
		next = self.Position.Add(self.Heading.Mul(self.Speed))
	} else if next.X >= float64(self.Width) {
		self.Heading = self.Heading.Reflect(Vector3Left).Normalize()
		next = self.Position.Add(self.Heading.Mul(self.Speed))
	} else if next.Y <= 0.0 {
		self.Heading = self.Heading.Reflect(Vector3Up).Normalize()
		next = self.Position.Add(self.Heading.Mul(self.Speed))
	} else if next.Y >= float64(self.Height) {
		self.Heading = self.Heading.Reflect(Vector3Down).Normalize()
		next = self.Position.Add(self.Heading.Mul(self.Speed))
	}

	self.Position = next
	return self.Position
}

type PongProgram struct {
	Window *glfw.Window

	pong *Pong

	// state
	frame      int32
	paused     bool
	cursorSize float64
	cmds       CmdChannels

	// textures
	tex *Texture

	// compute shaders
	pongShader Shader

	// output shaders
	outputShaders cyclicArray[Shader]
	gradientIndex cyclicArray[int32]

	// buffers
	fbo, vao, vbo uint32
}

func NewPongProgram() Program {
	cmds := NewCmdChannels()
	cmds.Register(RecolorCmd)

	return &PongProgram{
		pong: NewPong(),

		frame:      0,
		paused:     false,
		cursorSize: 0.025,

		cmds:          cmds,
		gradientIndex: *newCyclicArray([]int32{0, 1, 2, 3}),
	}
}

func (self *PongProgram) Load(window *glfw.Window, vao, vbo uint32) {
	self.Window = window
	self.vao = vao
	self.vbo = vbo
	width, height := window.GetSize()

	// create textures
	prev := *image.NewRGBA(image.Rect(0, 0, width, height))
	next := *image.NewRGBA(image.Rect(0, 0, width, height))
	for x := 0; x < prev.Rect.Max.X; x++ {
		for y := 0; y < prev.Rect.Max.Y; y++ {
			r := uint8(0)
			g := uint8(0)
			b := uint8(0)
			a := uint8(1)

			c := color.RGBA{r, g, b, a}
			prev.Set(x, y, c)
			next.Set(x, y, c)
		}
	}

	// create compute textures
	self.tex = LoadTexture(&prev)

	// create compute shaders
	self.pongShader = MustCompileShader(VertexShader, PongShader)

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

	// load turtle
	// center of screen
	distance := math.Min(float64(width), float64(height)) * 0.25
	x := (float64(width) / 2.0)
	y := (float64(height) / 2.0) + (distance / 4.0)
	self.pong.Position = Vector2{x, y}
	self.pong.Resize(width, height)
}

func (self *PongProgram) recolor() {
	width, height := self.Window.GetSize()

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)

	self.tex.Activate(gl.TEXTURE0)
	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (self *PongProgram) run(t float64) {
	width, height := self.Window.GetSize()
	mx, my := self.Window.GetCursorPos()

	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.tex.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.tex.Activate(gl.TEXTURE0)

	next := self.pong.Advance()

	self.pongShader.Use().
		Uniform1i("state", 0).
		Uniform1f("time", float32(t)).
		Uniform1f("size", float32(self.pong.Size)).
		Uniform2f("b", float32(next.X), float32(next.Y)).
		Uniform2f("scale", float32(width), float32(height)).
		Uniform2f("mouse", float32(mx), float32(height)-float32(my))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)
	self.tex.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (self *PongProgram) Render(t float64) {
	select {
	case <-self.cmds[RecolorCmd]:
		self.recolor()
	default:
		if self.paused {
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
			return
		}

		self.frame = self.frame + 1
		self.run(t)
	}
}

func (self *PongProgram) ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	if yoff > 0 {
		self.cursorSize = math.Max(0, self.cursorSize-0.005)
	} else if yoff < 0 {
		self.cursorSize = math.Min(1, self.cursorSize+0.005)
	}
}

func (self *PongProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.tex.Resize(width, height)
	self.pong.Resize(width, height)
}

func (self *PongProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Release {
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
