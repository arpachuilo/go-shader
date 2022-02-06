package main

import (
	_ "embed"

	"image"
	"image/color"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var RecolorCmd = "recolor"

type LifeProgram struct {
	Window *glfw.Window

	// rule set
	survive []int32
	birth   []int32

	// state
	paused bool
	cmds   CmdChannels

	// textures
	prevTexture        Texture
	nextTexture        Texture
	growthDecayTexture Texture

	// compute shaders
	lifeShader        Shader
	growthDecayShader Shader

	// output shaders
	outputShaders             []Shader
	outputShaderIndex         int
	outputShaderGradientIndex int32

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
		paused:  false,
		cmds:    cmds,
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
			var r uint8 = 0
			if rand.Float32() > 0.7 {
				r = 255
			}

			var g uint8 = 0
			if rand.Float32() > 0.3 {
				g = 255
			}

			var b uint8 = 0
			if rand.Float32() > 0.5 {
				b = 255
			}

			var a uint8 = 255
			if rand.Float32() > 0.5 {
				a = 255
			}

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
	p.lifeShader = MustCompileShader(vertexShader, golShader)
	p.growthDecayShader = MustCompileShader(vertexShader, gainShader)

	// create output shaders
	p.outputShaderIndex = 0
	p.outputShaders = make([]Shader, 0)
	p.outputShaders = append(p.outputShaders, MustCompileShader(vertexShader, rgbShader))
	p.outputShaders = append(p.outputShaders, MustCompileShader(vertexShader, viridisShader))
	p.outputShaders = append(p.outputShaders, MustCompileShader(vertexShader, infernoShader))
	p.outputShaders = append(p.outputShaders, MustCompileShader(vertexShader, magmaShader))
	p.outputShaders = append(p.outputShaders, MustCompileShader(vertexShader, plasmaShader))
	p.outputShaders = append(p.outputShaders, MustCompileShader(vertexShader, turboShader))

	// create framebuffers
	gl.GenFramebuffers(1, &p.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, p.fbo)
}

func (p *LifeProgram) Render(t float64) {
	select {
	case <-p.cmds[RecolorCmd]:
		width, height := p.Window.GetSize()

		// use copy program
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.BindVertexArray(p.vao)
		p.growthDecayTexture.Activate(gl.TEXTURE0)

		p.outputShaders[p.outputShaderIndex].Use().
			Uniform1i("state", 0).
			Uniform2f("scale", float32(width), float32(height))
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
	default:
		if p.paused {
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
			return
		}

		width, height := p.Window.GetSize()

		// use gol program
		gl.BindFramebuffer(gl.FRAMEBUFFER, p.fbo)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, uint32(p.nextTexture), 0)

		gl.BindVertexArray(p.vao)
		p.prevTexture.Activate(gl.TEXTURE0)

		p.lifeShader.Use().
			Uniform1iv("s", p.survive).
			Uniform1iv("b", p.birth).
			Uniform1i("state", 0).
			Uniform2f("scale", float32(width), float32(height))
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

		// swap texture
		p.prevTexture, p.nextTexture = p.nextTexture, p.prevTexture

		// use decay program
		gl.BindFramebuffer(gl.FRAMEBUFFER, p.fbo)
		gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, uint32(p.growthDecayTexture), 0)

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

		p.outputShaders[p.outputShaderIndex].Use().
			Uniform1i("index", p.outputShaderGradientIndex).
			Uniform1i("state", 0).
			Uniform2f("scale", float32(width), float32(height))
		gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
	}
}

func (p *LifeProgram) ResizeCallback(w *glfw.Window, width int, height int) {

}

func (p *LifeProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action == glfw.Release {
		if key == glfw.KeyJ {
			p.outputShaderIndex--
			if p.outputShaderIndex < 0 {
				p.outputShaderIndex = len(p.outputShaders) - 1
			}

			p.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeyK {
			p.outputShaderIndex++
			if p.outputShaderIndex > len(p.outputShaders)-1 {
				p.outputShaderIndex = 0
			}

			p.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeyH {
			p.outputShaderGradientIndex--
			if p.outputShaderGradientIndex < 0 {
				p.outputShaderGradientIndex = 2
			}

			p.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeyL {
			p.outputShaderGradientIndex++
			if p.outputShaderGradientIndex > 2 {
				p.outputShaderGradientIndex = 0
			}

			p.cmds.Issue(RecolorCmd)
		}

		if key == glfw.KeySpace {
			p.paused = !p.paused
		}
	}
}

//go:embed shaders/vertex.glsl
var vertexShader string

//go:embed shaders/life/life.glsl
var golShader string

//go:embed shaders/life/growth_decay.glsl
var gainShader string

//go:embed shaders/rgb_sampler.glsl
var rgbShader string

//go:embed shaders/gradients/viridis.glsl
var viridisShader string

//go:embed shaders/gradients/magma.glsl
var magmaShader string

//go:embed shaders/gradients/inferno.glsl
var infernoShader string

//go:embed shaders/gradients/plasma.glsl
var plasmaShader string

//go:embed shaders/gradients/turbo.glsl
var turboShader string
