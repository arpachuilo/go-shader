package main

import (
	"image"
	"io/ioutil"
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/rjeczalik/notify"
)

type LiveProgram struct {
	Window *glfw.Window

	// current texture
	texture *Texture

	// current shader
	shader Shader

	// name of file to watch for updates
	filename string
	pending  chan string
	watcher  chan notify.EventInfo

	// history of successfuly compiled shaders
	history []string

	// buffers
	vao, vbo uint32
}

func NewLiveProgram(filename string) Program {
	return &LiveProgram{
		filename: filename,
		history:  make([]string, 0),
	}
}

func (p *LiveProgram) watch() {
	p.watcher = make(chan notify.EventInfo, 1)
	p.pending = make(chan string, 1)
	err := notify.Watch(p.filename, p.watcher, notify.Create, notify.Write)
	if err != nil {
		panic(err)
	}
	defer notify.Stop(p.watcher)

	// initial file load
	data, err := ioutil.ReadFile(p.filename)
	if err == nil {
		p.pending <- string(data)
	}

	for {
		select {
		case ei := <-p.watcher:
			event := ei.Event()
			switch event {
			case notify.Rename:
				// p.filename = ei.Path()
			case notify.Write:
				log.Println("modified file:", ei.Path())

				// read file in
				data, err := ioutil.ReadFile(p.filename)
				if err != nil {
					log.Println("error:", err)
					continue
				}

				// attempt to create new shader
				p.pending <- string(data)
			}
		}
	}

}

func (p *LiveProgram) Load(window *glfw.Window, vao, vbo uint32) {
	p.Window = window
	p.vao = vao
	p.vbo = vbo
	width, height := window.GetSize()

	// create textures
	img := *image.NewRGBA(image.Rect(0, 0, width, height))
	p.texture = LoadTexture(&img)

	p.shader = MustCompileShader(vertexShader, fragShader)

	go p.watch()
}

func (p *LiveProgram) compile(code string) {
	newShader, err := CompileShader(vertexShader, code)
	if err != nil {
		log.Println("error", err)
		return
	}

	p.shader = newShader
	p.history = append(p.history, code)
}

func (p *LiveProgram) run(t float64) {
	width, height := p.Window.GetSize()
	mx, my := p.Window.GetCursorPos()

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(p.vao)
	p.texture.Activate(gl.TEXTURE0)

	p.shader.Use().
		Uniform1i("state", 0).
		Uniform1f("time", float32(t)).
		Uniform2f("mouse", float32(mx), float32(height)-float32(my)).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (p *LiveProgram) Render(t float64) {
	select {
	case code := <-p.pending:
		p.compile(code)
	default:
		p.run(t)
	}
}

func (p *LiveProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
}

func (p *LiveProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	p.texture.Resize(width, height)
}
