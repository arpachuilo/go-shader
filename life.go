package main

import (
	_ "embed"

	"image"
	"image/color"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type LifeProgram struct {
	Window *glfw.Window

	// rule set
	survive []int32
	birth   []int32

	// textures
	prevTexture        Texture
	nextTexture        Texture
	growthDecayTexture Texture

	// shaders
	lifeShader        *Shader
	growthDecayShader *Shader
	colorizeShader    *Shader

	// framebuffers
	fbo, vao, vbo uint32
}

func NewLifeProgram() Program {
	survive := []int32{-1, -1, 2, 3, -1, -1, -1, -1, -1}
	birth := []int32{-1, -1, -1, 3, -1, -1, -1, -1, -1}

	return &LifeProgram{
		survive: survive,
		birth:   birth,
	}
}

func (lp *LifeProgram) Load(window *glfw.Window, vao, vbo uint32) {
	lp.Window = window
	lp.vao = vao
	lp.vbo = vbo
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

	// create textures
	lp.prevTexture = LoadTexture(&img1)
	lp.nextTexture = LoadTexture(&img2)
	lp.growthDecayTexture = LoadTexture(&img3)

	// create shaders
	lp.lifeShader = MustCompileShader(vertexShader, golShader)
	lp.growthDecayShader = MustCompileShader(vertexShader, gainShader)
	lp.colorizeShader = MustCompileShader(vertexShader, copyShader)

	// create framebuffers
	gl.GenFramebuffers(1, &lp.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, lp.fbo)
}

func (lp *LifeProgram) Render(t float64) {
	width, height := lp.Window.GetSize()

	// use gol program
	gl.BindFramebuffer(gl.FRAMEBUFFER, lp.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, uint32(lp.nextTexture), 0)

	gl.BindVertexArray(lp.vao)
	lp.prevTexture.Activate(gl.TEXTURE0)

	lp.lifeShader.Use().
		Uniform1iv("s", lp.survive).
		Uniform1iv("b", lp.birth).
		Uniform1i("state", 0).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// swap texture
	lp.prevTexture, lp.nextTexture = lp.nextTexture, lp.prevTexture

	// use decay program
	gl.BindFramebuffer(gl.FRAMEBUFFER, lp.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, uint32(lp.growthDecayTexture), 0)

	gl.BindVertexArray(lp.vao)
	lp.prevTexture.Activate(gl.TEXTURE0)
	lp.growthDecayTexture.Activate(gl.TEXTURE1)

	lp.growthDecayShader.Use().
		Uniform1i("state", 0).
		Uniform1i("self", 1).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(lp.vao)
	lp.growthDecayTexture.Activate(gl.TEXTURE0)

	lp.colorizeShader.Use().
		Uniform1i("state", 0).
		Uniform1f("time", float32(t)).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (lp *LifeProgram) ResizeCallback(w *glfw.Window, width int, height int) {

}

func (lp *LifeProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

}

//go:embed shaders/vertex.glsl
var vertexShader string

//go:embed shaders/frag.glsl
var fragmentShader string

//go:embed shaders/life/life.glsl
var golShader string

//go:embed shaders/life/growthDecay.glsl
var gainShader string

//go:embed shaders/colorize.glsl
var copyShader string
