package main

import (
	"fmt"
	"image"
	"image/color"
	"math/rand"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Renderer handles running our programs
type Renderer struct {
	// Pipeable programs
	Programs []Program
	Window   *glfw.Window

	RefreshRate   float64
	Width, Height int

	vao uint32
	vbo uint32
}

// NewRenderer Create new renderer
func NewRenderer(window *glfw.Window) *Renderer {
	programs := make([]Program, 0)
	return &Renderer{
		Programs: programs,
		Window:   window,
	}
}

func (r *Renderer) Setup() {
	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	// get current resolution
	r.Width, r.Height = r.Window.GetSize()

	// get refresh rate
	r.RefreshRate = float64(glfw.GetPrimaryMonitor().GetVideoMode().RefreshRate)

	// print some info
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
	fmt.Println("Refresh rate", r.RefreshRate)

	// register callbacks
	r.Window.SetKeyCallback(r.KeyCallback)
	r.Window.SetSizeCallback(r.ResizeCallback)

	// configure the vertex data
	gl.GenVertexArrays(1, &r.vao)
	gl.BindVertexArray(r.vao)

	gl.GenBuffers(1, &r.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, r.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*4, gl.Ptr(quadVertices), gl.STATIC_DRAW)

	// Configure global settings
	// gl.ClearColor(0.0, 0.0, 0.0, 1.0)
}

var frames = 0
var lastTime time.Time

func (r *Renderer) Start() {
	tick := time.Tick(time.Duration(1000/r.RefreshRate) * time.Millisecond)

	// create textures
	img1 := *image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
	img2 := *image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
	img3 := *image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
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

	prevTexture := LoadTexture(&img1)
	nextTexture := LoadTexture(&img2)
	gainTexture := LoadTexture(&img3)

	survive := []int32{-1, -1, 2, 3, -1, -1, -1, -1, -1}
	birth := []int32{-1, -1, -1, 3, -1, -1, -1, -1, -1}

	lifeShader := MustCompileShader(vertexShader, golShader)
	gainShader := MustCompileShader(vertexShader, gainShader)
	copyShader := MustCompileShader(vertexShader, copyShader)

	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)

	for !r.Window.ShouldClose() {
		select {
		// frame limiter
		case <-tick:
			t := glfw.GetTime()
			frames++
			currentTime := time.Now()
			delta := currentTime.Sub(lastTime)
			if delta > time.Second {
				fps := frames / int(delta.Seconds())
				r.Window.SetTitle(fmt.Sprintf("FPS: %v", fps))

				lastTime = currentTime
				frames = 0
			}

			// use gol program
			gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, nextTexture.Texture, 0)

			gl.BindVertexArray(r.vao)
			prevTexture.Activate(gl.TEXTURE0)

			lifeShader.Use().
				Uniform1iv("s", survive).
				Uniform1iv("b", birth).
				Uniform1i("state", 0).
				Uniform2f("scale", float32(r.Width), float32(r.Height))
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

			// swap texture
			prevTexture, nextTexture = nextTexture, prevTexture

			// use decay program
			gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, gainTexture.Texture, 0)

			gl.BindVertexArray(r.vao)
			prevTexture.Activate(gl.TEXTURE0)
			gainTexture.Activate(gl.TEXTURE1)

			gainShader.Use().
				Uniform1i("state", 0).
				Uniform1i("self", 1).
				Uniform2f("scale", float32(r.Width), float32(r.Height))
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

			// use copy program
			gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
			gl.BindVertexArray(r.vao)
			gainTexture.Activate(gl.TEXTURE0)

			copyShader.Use().
				Uniform1i("state", 0).
				Uniform1f("time", float32(t)).
				Uniform2f("scale", float32(r.Width), float32(r.Height))
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

			// Maintenance
			r.Window.SwapBuffers()
			glfw.PollEvents()
		}
	}

	glfw.Terminate()
}

func (r *Renderer) ResizeCallback(w *glfw.Window, width int, height int) {
	r.Width, r.Height = width, height
}

func (r *Renderer) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	// call key handlers for each program
	for _, p := range r.Programs {
		p.KeyCallback(key, scancode, action, mods)
	}
}
