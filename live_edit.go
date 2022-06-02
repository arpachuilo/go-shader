package main

import (
	"image"
	"io/ioutil"
	"log"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/rjeczalik/notify"
)

type LiveEditProgram struct {
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

func NewLiveEditProgram(filename string) Program {
	return &LiveEditProgram{
		filename: filename,
		history:  make([]string, 0),
	}
}

func (self *LiveEditProgram) watch() {
	self.watcher = make(chan notify.EventInfo, 1)
	self.pending = make(chan string, 1)
	err := notify.Watch(self.filename, self.watcher, notify.Create, notify.Write)
	if err != nil {
		panic(err)
	}
	defer notify.Stop(self.watcher)

	// initial file load
	data, err := ioutil.ReadFile(self.filename)
	if err == nil {
		self.pending <- string(data)
	}

	for {
		select {
		case ei := <-self.watcher:
			event := ei.Event()
			switch event {
			case notify.Rename:
				// p.filename = ei.Path()
			case notify.Write:
				log.Println("modified file:", ei.Path())

				// read file in
				data, err := ioutil.ReadFile(self.filename)
				if err != nil {
					log.Println("error:", err)
					continue
				}

				// attempt to create new shader
				self.pending <- string(data)
			}
		}
	}

}

func (self *LiveEditProgram) Load(window *glfw.Window, vao, vbo uint32) {
	self.Window = window
	self.vao = vao
	self.vbo = vbo
	width, height := window.GetSize()

	// create textures
	img := *image.NewRGBA(image.Rect(0, 0, width, height))
	self.texture = LoadTexture(&img)

	self.shader = MustCompileShader(vertexShader, fragShader)

	go self.watch()
}

func (self *LiveEditProgram) compile(code string) {
	newShader, err := CompileShader(vertexShader, code)
	if err != nil {
		log.Println("error", err)
		return
	}

	self.shader = newShader
	self.history = append(self.history, code)
}

func (self *LiveEditProgram) run(t float64) {
	width, height := self.Window.GetSize()
	mx, my := self.Window.GetCursorPos()

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.vao)
	self.texture.Activate(gl.TEXTURE0)

	self.shader.Use().
		Uniform1i("state", 0).
		Uniform1f("time", float32(t)).
		Uniform2f("mouse", float32(mx), float32(height)-float32(my)).
		Uniform2f("scale", float32(width), float32(height))
	gl.DrawArrays(gl.TRIANGLE_FAN, 0, 6)
}

func (self *LiveEditProgram) Render(t float64) {
	select {
	case code := <-self.pending:
		self.compile(code)
	default:
		self.run(t)
	}
}

func (self *LiveEditProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
}

func (self *LiveEditProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.texture.Resize(width, height)
}
