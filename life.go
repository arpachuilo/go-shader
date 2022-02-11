package main

import (
	"image"
	"image/color"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var RecolorCmd = "recolor"

type LifeMode int

const (
	LifeStd LifeMode = iota
	LifeCyclic
)

type LifeProgram struct {
	Window *glfw.Window

	// rule set
	survive []int32
	birth   []int32

	// state
	paused bool
	mode   LifeMode
	cmds   CmdChannels

	// textures
	prevTexture        *Texture
	nextTexture        *Texture
	growthDecayTexture *Texture

	// compute shaders
	lifeShader        Shader
	cyclicShader      Shader
	growthDecayShader Shader

	// output shaders
	outputShaders ShaderCycler
	gradientIndex CyclableInt

	// buffers
	fbo, vao, vbo uint32
}

func NewLifeProgram() Program {
	survive := []int32{-1, -1, 2, 3, -1, -1, -1, -1, -1}
	birth := []int32{-1, -1, -1, 3, -1, -1, -1, -1, -1}

	cmds := NewCmdChannels()
	cmds.Register(RecolorCmd)

	return &LifeProgram{
		survive: survive,
		birth:   birth,

		paused: false,
		mode:   LifeStd,

		cmds:          cmds,
		gradientIndex: *NewCyclableInt(0, 4),
	}
}

func (p *LifeProgram) Load(window *glfw.Window, vao, vbo uint32) {
	p.Window = window
	p.vao = vao
	p.vbo = vbo
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
			a := uint8(rand.Intn(255))

			c := color.RGBA{r, g, b, a}
			img1.Set(x, y, c)
			img2.Set(x, y, c)
			img3.Set(x, y, color.White)
		}
	}

	// create compute textures
	p.prevTexture = LoadTexture(&img1)
	p.nextTexture = LoadTexture(&img2)
	p.growthDecayTexture = LoadTexture(&img3)

	// create compute shaders
	p.cyclicShader = MustCompileShader(vertexShader, cyclicShader)
	p.lifeShader = MustCompileShader(vertexShader, golShader)
	p.growthDecayShader = MustCompileShader(vertexShader, gainShader)

	// create output shaders
	p.outputShaders = *NewShaderCyclerFromArray([]Shader{
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

	// create framebuffers
	gl.GenFramebuffers(1, &p.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.fbo)
}

func (p *LifeProgram) recolor() {
	width, height := p.Window.GetSize()

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(p.vao)

	switch p.mode {
	case LifeStd:
		p.growthDecayTexture.Activate(gl.TEXTURE0)
	case LifeCyclic:
		p.prevTexture.Activate(gl.TEXTURE0)
	}

	p.outputShaders.Current().Use().
		Uniform1i("index", int32(p.gradientIndex.Index())).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

}

func (p *LifeProgram) life(t float64) {
	width, height := p.Window.GetSize()
	mx, my := p.Window.GetCursorPos()

	// use gol program
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, p.nextTexture.Handle, 0)

	gl.BindVertexArray(p.vao)
	p.prevTexture.Activate(gl.TEXTURE0)

	p.lifeShader.Use().
		Uniform1iv("s", p.survive).
		Uniform1iv("b", p.birth).
		Uniform1i("state", 0).
		Uniform1f("time", float32(t)).
		Uniform2f("scale", float32(width), float32(height)).
		Uniform2f("mouse", float32(mx), float32(height)-float32(my))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// swap texture
	p.prevTexture, p.nextTexture = p.nextTexture, p.prevTexture

	// use decay program
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, p.growthDecayTexture.Handle, 0)

	gl.BindVertexArray(p.vao)
	p.prevTexture.Activate(gl.TEXTURE0)
	p.growthDecayTexture.Activate(gl.TEXTURE1)

	p.growthDecayShader.Use().
		Uniform1i("state", 0).
		Uniform1i("self", 1).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(p.vao)
	p.growthDecayTexture.Activate(gl.TEXTURE0)

	p.outputShaders.Current().Use().
		Uniform1i("index", int32(p.gradientIndex.Index())).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

}

func (p *LifeProgram) cyclic(t float64) {
	width, height := p.Window.GetSize()
	mx, my := p.Window.GetCursorPos()

	// use cyclic life program
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, p.nextTexture.Handle, 0)

	gl.BindVertexArray(p.vao)
	p.prevTexture.Activate(gl.TEXTURE0)

	p.cyclicShader.Use().
		Uniform1f("stages", 16.0).
		Uniform1i("state", 0).
		Uniform1f("time", float32(t)).
		Uniform2f("scale", float32(width), float32(height)).
		Uniform2f("mouse", float32(mx), float32(height)-float32(my))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// swap texture
	p.prevTexture, p.nextTexture = p.nextTexture, p.prevTexture

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(p.vao)
	p.prevTexture.Activate(gl.TEXTURE0)

	p.outputShaders.Current().Use().
		Uniform1i("index", int32(p.gradientIndex.Index())).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (p *LifeProgram) Render(t float64) {
	select {
	case <-p.cmds[RecolorCmd]:
		p.recolor()
	default:
		if p.paused {
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
			return
		}

		switch p.mode {
		case LifeStd:
			p.life(t)
		case LifeCyclic:
			p.cyclic(t)
		}
	}
}

func (p *LifeProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	p.prevTexture.Resize(width, height)
	p.nextTexture.Resize(width, height)
	p.growthDecayTexture.Resize(width, height)
}

func (p *LifeProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Release {
		if key == glfw.Key1 {
			p.mode = LifeStd
		}

		if key == glfw.Key2 {
			p.mode = LifeCyclic
		}

		if key == glfw.KeyJ {
			p.outputShaders.Prev()
			p.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeyK {
			p.outputShaders.Next()
			p.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeyH {
			p.gradientIndex.Prev()
			p.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeyL {
			p.gradientIndex.Next()
			p.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeySpace {
			p.paused = !p.paused
		}
	}
}
