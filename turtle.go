package main

import (
	"image"
	"image/color"
	"math"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"

	"go-trip/assets"
)

type Vector2 struct {
	X float64
	Y float64
}

func (self Vector2) Add(other Vector2) Vector2 {
	return Vector2{
		X: self.X + other.X,
		Y: self.Y + other.Y,
	}
}

func (self Vector2) Mul(scalar float64) Vector2 {
	return Vector2{
		X: self.X * scalar,
		Y: self.Y * scalar,
	}
}

func (self Vector2) Turn(degrees float64) Vector2 {
	theta := Deg2Rad(degrees)
	cs := math.Cos(theta)
	sn := math.Sin(theta)
	vx := self.X*cs - self.Y*sn
	vy := self.X*sn + self.Y*cs
	return Vector2{
		X: vx,
		Y: vy,
	}
}

func Deg2Rad(degrees float64) float64 {
	return degrees * math.Pi / 180
}

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

func (self *Turtle) Dot() (Vector2, Vector2) {
	return self.Position, self.Position
}

type TurtleProgram struct {
	Window *glfw.Window

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
	fbo, vao, vbo uint32
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

func (self *TurtleProgram) Load(window *glfw.Window, vao, vbo uint32) {
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
	self.turtleShader = MustCompileShader(assets.VertexShader, assets.TurtleShader)

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

	// load turtle
	// center of screen
	distance := math.Min(float64(width), float64(height)) * 0.25
	x := (float64(width) / 2.0)
	y := (float64(height) / 2.0) + (distance / 4.0)
	self.turtle.Position = Vector2{x, y}
	self.turtle.Turn(100)
}

func (self *TurtleProgram) recolor() {
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

func (self *TurtleProgram) run(t float64) {
	width, height := self.Window.GetSize()
	mx, my := self.Window.GetCursorPos()

	// use gol program
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.tex.Handle, 0)

	gl.BindVertexArray(self.vao)
	self.tex.Activate(gl.TEXTURE0)

	// prev, next := self.turtle.Dot()
	prev, next := self.turtle.Advance()

	// travel 1/12 min(width, height) before turning
	// uses refresh rate of 60 (make this dynamic later)
	distance := math.Min(float64(width), float64(height)) * 0.15
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

func (self *TurtleProgram) Render(t float64) {
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

func (self *TurtleProgram) ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	if yoff > 0 {
		self.cursorSize = math.Max(0, self.cursorSize-0.005)
	} else if yoff < 0 {
		self.cursorSize = math.Min(1, self.cursorSize+0.005)
	}
}

func (self *TurtleProgram) ResizeCallback(w *glfw.Window, width int, height int) {
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
