package main

import (
	"image"
	"image/color"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Turtle struct {
	Down     int32
	Heading  Vector2
	Position Vector2
	Speed    float64
}

func NewTurtle() *Turtle {
	return &Turtle{
		Down: 1,
		Position: Vector2{
			X: 0,
			Y: 0,
		},
		Heading: Vector2{
			X: 0,
			Y: -1,
		},
		Speed: 1,
	}
}

func (self *Turtle) PenToggle() {
	if self.Down == 1 {
		self.Down = 0
	} else {
		self.Down = 1
	}
}

func (self *Turtle) PenUp() {
	self.Down = 1
}

func (self *Turtle) PenDown() {
	self.Down = 0
}

func (self *Turtle) Turn(degrees float64) Vector2 {
	self.Heading = self.Heading.Turn(degrees)
	return self.Heading
}

func (self *Turtle) Advance() (Vector2, Vector2) {
	previous := self.Position
	self.Position = previous.Add(self.Heading.Mul(self.Speed))
	return previous, self.Position
}

type TurtleProgram struct {
	Window        *glfw.Window
	width, height int

	turtle *Turtle

	// state
	frame      int32
	paused     bool
	cursorSize float64
	cmds       CmdChannels

	// textures
	tex *Texture

	// compute shaders
	turtleShader Shader

	// output shaders
	outputShaders cyclicArray[Shader]
	gradientIndex cyclicArray[int32]

	// buffers
	fbo uint32
	bo  BufferObject
}

func NewTurtleProgram() Program {
	cmds := NewCmdChannels()
	cmds.Register(RecolorCmd)

	return &TurtleProgram{
		turtle: NewTurtle(),

		frame:      0,
		paused:     false,
		cursorSize: 0.025,

		cmds:          cmds,
		gradientIndex: *newCyclicArray([]int32{0, 1, 2, 3}),
	}
}

func (self *TurtleProgram) Load(window *glfw.Window, bo BufferObject) {
	self.Window = window
	self.bo = bo
	self.width, self.height = window.GetFramebufferSize()

	// create textures
	prev := *image.NewRGBA(image.Rect(0, 0, self.width, self.height))
	next := *image.NewRGBA(image.Rect(0, 0, self.width, self.height))
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
	self.turtleShader = MustCompileShader(VertexShader, TurtleShader, self.bo)

	// create output shaders
	self.outputShaders = *newCyclicArray([]Shader{
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

	// load turtle
	// center of screen
	distance := math.Min(float64(self.width), float64(self.height)) * 0.25
	x := (float64(self.width) / 2.0)
	y := (float64(self.height) / 2.0) + (distance / 4.0)
	self.turtle.Position = Vector2{x, y}
	self.turtle.Turn(100)
}

func (self *TurtleProgram) recolor() {
	width, height := self.Window.GetFramebufferSize()

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.bo.VAO())

	self.tex.Activate(gl.TEXTURE0)
	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	self.bo.Draw()
}

func (self *TurtleProgram) run(t float64) {
	mx, my := self.Window.GetCursorPos()

	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.tex.Handle, 0)

	gl.BindVertexArray(self.bo.VAO())
	self.tex.Activate(gl.TEXTURE0)

	// prev, next := self.turtle.Dot()
	prev, next := self.turtle.Advance()

	// travel 1/12 min(width, height) before turning
	// uses refresh rate of 60 (make this dynamic later)
	distance := math.Min(float64(self.width), float64(self.height)) * 0.15
	pivot1 := int32(distance * self.turtle.Speed)
	pivot2 := int32(distance*self.turtle.Speed) / 2.0
	if self.frame%int32(pivot1) == 0 {
		self.turtle.Turn(90)
	}

	if self.frame%int32(pivot2) == 0 {
		self.turtle.Turn(100)
	}

	w := math.Abs(math.Max(2.0, math.Sin(t)*6))
	if self.frame%10 == 0 {
		// self.turtle.PenToggle()
	}

	self.turtleShader.Use().
		Uniform1i("state", 0).
		Uniform1i("d", self.turtle.Down).
		Uniform1f("w", float32(w)).
		Uniform2f("a", float32(prev.X), float32(prev.Y)).
		Uniform2f("b", float32(next.X), float32(next.Y)).
		Uniform1f("time", float32(t)).
		Uniform2f("scale", float32(self.width), float32(self.height)).
		Uniform2f("mouse", float32(mx), float32(self.height)-float32(my))
	self.bo.Draw()

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.bo.VAO())
	self.tex.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(self.width), float32(self.height))
	self.bo.Draw()
}

func (self *TurtleProgram) Render(t float64) {
	select {
	case <-self.cmds[RecolorCmd]:
		self.recolor()
	default:
		if self.paused {
			self.bo.Draw()
			return
		}

		self.frame = self.frame + 1
		self.run(t)
	}
}

func (self *TurtleProgram) ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	if yoff > 0 {
		self.cursorSize = math.Max(0, self.cursorSize-0.005)
	} else if yoff < 0 {
		self.cursorSize = math.Min(1, self.cursorSize+0.005)
	}
}

func (self *TurtleProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.width, self.height = self.Window.GetFramebufferSize()
	self.tex.Resize(width, height)
}

func (self *TurtleProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
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
