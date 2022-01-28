package main

import (
	"fmt"
	"image"
	"image/color"
	"time"
	"math/rand"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// Renderer handles running our programs
type Renderer struct {
	// Pipeable programs
	Programs []Program
	Window *glfw.Window

	RefreshRate float64
	Width, Height int

	vao uint32
	vbo uint32
}

// NewRenderer Create new renderer
func NewRenderer(window *glfw.Window) *Renderer {
	programs := make([]Program, 0)
	return &Renderer{
		Programs: programs,
		Window: window,
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
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
}

func (r *Renderer) Start() {
	tick := time.Tick(time.Duration(1000/r.RefreshRate) * time.Millisecond)

	// TODO: REFACTOR
	// Configure the vertex and fragment shaders
	program, err := newProgram(vertexShader, golShader)
	if err != nil {
		fmt.Println(err)
	}

	// Configure the vertex and fragment shaders
	copyProgram, err := newProgram(vertexShader, copyShader)
	if err != nil {
		fmt.Println(err)
	}

	// bindings
	// shared with copyProgram
	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// shared with copyProgram
	vertAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertAttrib)
	gl.VertexAttribPointerWithOffset(vertAttrib, 2, gl.FLOAT, false, 4*4, 0)

	// shared with copyProgram
	texCoordAttrib := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoordAttrib)
	gl.VertexAttribPointerWithOffset(texCoordAttrib, 2, gl.FLOAT, false, 4*4, 2*4)


	// create textures
	img1 := *image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
	img2 := *image.NewRGBA(image.Rect(0, 0, windowWidth, windowHeight))
	for x := 0; x < img1.Rect.Max.X; x++ {
		for y := 0; y < img1.Rect.Max.Y; y++ {
			if rand.Float32() < 0.5 {
				img1.Set(x, y, color.White)
				img2.Set(x, y, color.Black)
			} else {
				img1.Set(x, y, color.Black)
				img2.Set(x, y, color.White)
			}
		}
	}

	prevTexture, err := newTexture(&img1)
	if err != nil {
		panic(err)
	}

	// create back texture
	nextTexture, err := newTexture(&img2)
	if err != nil {
		panic(err)
	}

	stateAttrib := gl.GetUniformLocation(program, gl.Str("state\x00"))
	scaleAttrib := gl.GetUniformLocation(program, gl.Str("scale\x00"))

	copyStateAttrib := gl.GetUniformLocation(copyProgram, gl.Str("state\x00"))
	copyScaleAttrib := gl.GetUniformLocation(copyProgram, gl.Str("scale\x00"))
	copyTimeAttrib := gl.GetUniformLocation(copyProgram, gl.Str("time\x00"))

	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, nextTexture, 0)


	// use copy program
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(r.vao)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, nextTexture)

	for !r.Window.ShouldClose() {
		select {
		// frame limiter
		case <-tick:
			t := glfw.GetTime()

			// use gol program
			gl.BindFramebuffer(gl.FRAMEBUFFER, fbo)
			gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, nextTexture, 0)

			gl.BindVertexArray(r.vao)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, prevTexture)

			gl.UseProgram(program)
			gl.ProgramUniform1i(program, stateAttrib, 0)
			gl.ProgramUniform2f(program, scaleAttrib, float32(r.Width), float32(r.Height))
			gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)

			// swap texture
			prevTexture, nextTexture = nextTexture, prevTexture

			// use copy program
			gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
			gl.BindVertexArray(r.vao)
			gl.ActiveTexture(gl.TEXTURE0)
			gl.BindTexture(gl.TEXTURE_2D, prevTexture)

			gl.UseProgram(copyProgram)
			gl.ProgramUniform1i(copyProgram, copyStateAttrib, 0)
			gl.ProgramUniform2f(copyProgram, copyScaleAttrib, float32(r.Width), float32(r.Height))
			gl.ProgramUniform1f(copyProgram, copyTimeAttrib, float32(t))
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


