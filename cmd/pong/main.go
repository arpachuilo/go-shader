package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl64"

	. "gogl"
	. "gogl/arrayutil"
	. "gogl/assets"
	. "gogl/mathutil"

	_ "embed"
)

var BuildDate = ""

func HotRender(kill <-chan bool, window *glfw.Window) {
	fmt.Println(BuildDate)
	program := NewPongProgram()

	NewRenderer(window, program).Run(kill)
}

var RecolorCmd = "recolor"

//go:embed pong.glsl
var PongShader string

type Pong struct {
	Heading  mgl64.Vec2
	Position mgl64.Vec2

	Speed float64
	Size  float64

	Width  int
	Height int
}

// TODO: Look into storing this inside texture for use with sampler
func NewPong() *Pong {
	return &Pong{
		Position: mgl64.Vec2{rand.Float64(), rand.Float64()},
		Heading:  mgl64.Vec2{-1.0 + rand.Float64()*2, -1.0 + rand.Float64()*2}.Normalize(),
		Speed:    rand.Float64()*10 + 1.0,
		Size:     rand.Float64()*12 + 3,
	}
}

func (self *Pong) Resize(width, height int) {
	self.Position[0] = rand.Float64() * float64(width)
	self.Position[1] = rand.Float64() * float64(height)
	self.Width = width
	self.Height = height
}

func (self *Pong) Turn(degrees float64) mgl64.Vec2 {
	self.Heading = TurnVec2(self.Heading, degrees)
	return self.Heading
}

func (self *Pong) Advance() {
	next := self.Position.Add(self.Heading.Mul(self.Speed))

	if next[0] <= self.Size {
		self.Heading = ReflectVec2(self.Heading, Vector3Right).Normalize()
		next = self.Position.Add(self.Heading.Mul(self.Speed))
	} else if next[0] >= float64(self.Width)-self.Size {
		self.Heading = ReflectVec2(self.Heading, Vector3Left).Normalize()
		next = self.Position.Add(self.Heading.Mul(self.Speed))
	} else if next[1] <= self.Size {
		self.Heading = ReflectVec2(self.Heading, Vector3Up).Normalize()
		next = self.Position.Add(self.Heading.Mul(self.Speed))
	} else if next[1] >= float64(self.Height)-self.Size {
		self.Heading = ReflectVec2(self.Heading, Vector3Down).Normalize()
		next = self.Position.Add(self.Heading.Mul(self.Speed))
	}

	self.Position = next
}

type PongProgram struct {
	Window *glfw.Window

	pong []*Pong

	// state
	frame  int32
	paused bool
	alpha  float64
	cmds   CmdChannels

	// textures
	tex *Texture

	// compute shaders
	pongShader Shader

	// output shaders
	outputShaders CyclicArray[Shader]
	gradientIndex CyclicArray[int32]

	// buffers
	fbo uint32
	bo  BufferObject
}

func NewPongProgram() Program {
	cmds := NewCmdChannels()
	cmds.Register(RecolorCmd)

	pongs := make([]*Pong, 0)
	for i := 0; i < 100; i++ {
		pongs = append(pongs, NewPong())
	}

	return &PongProgram{
		pong: pongs,

		frame:  0,
		paused: false,
		alpha:  0.0,

		cmds:          cmds,
		gradientIndex: *NewCyclicArray([]int32{0, 1, 2, 3}),
	}
}

func (self *PongProgram) Load(window *glfw.Window) {
	self.Window = window
	self.bo = NewVBuffer(QuadVertices, 2, 4)
	width, height := window.GetFramebufferSize()

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
	self.pongShader = MustCompileShader(VertexShader, PongShader, self.bo)

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

	for _, p := range self.pong {
		p.Resize(width, height)
	}
}

func (self *PongProgram) recolor() {
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

func (self *PongProgram) run(t float64) {
	width, height := self.Window.GetFramebufferSize()
	mx, my := self.Window.GetCursorPos()

	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.tex.Handle, 0)

	gl.BindVertexArray(self.bo.VAO())
	self.tex.Activate(gl.TEXTURE0)

	for _, p := range self.pong {
		p.Advance()
	}

	sizes := make([]float32, 0)
	pongs := make([]float32, 0)
	for _, p := range self.pong {
		sizes = append(sizes, float32(p.Size))
		pongs = append(pongs, float32(p.Position[0]))
		pongs = append(pongs, float32(p.Position[1]))
	}

	self.pongShader.Use().
		Uniform2fv("pPos", pongs).
		Uniform1fv("pSize", sizes).
		Uniform1i("len", int32(len(self.pong))).
		Uniform1i("iChannel1", 0).
		Uniform1f("iTime", float32(t)).
		Uniform2f("iResolution", float32(width), float32(height)).
		Uniform2f("iMouse", float32(mx), float32(height)-float32(my))
	self.bo.Draw()

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.bo.VAO())
	self.tex.Activate(gl.TEXTURE0)

	self.outputShaders.Current().Use().
		Uniform1i("index", *self.gradientIndex.Current()).
		Uniform1f("alpha", float32(self.alpha)).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	self.bo.Draw()
}

func (self *PongProgram) Render(t float64) {
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

func (self *PongProgram) ScrollCallback(w *glfw.Window, xoff float64, yoff float64) {
	if yoff > 0 {
		self.alpha = math.Max(0, self.alpha-0.005)
	} else if yoff < 0 {
		self.alpha = math.Min(1, self.alpha+0.005)
	}
}

func (self *PongProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.tex.Resize(width, height)

	for _, p := range self.pong {
		p.Resize(width, height)
	}
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
