package main

import (
	"image"
	"log"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type LiveEditProgram struct {
	Window        *glfw.Window
	width, height int

	paused bool
	frame  int

	watcher *fsnotify.Watcher

	// current texture
	texture *Texture

	// current shader
	vert   string
	frag   string
	shader Shader

	// name of file to watch for updates
	vertFilename string
	fragFilename string

	// history of successfully compiled shaders
	vertHistory []string
	fragHistory []string

	// buffers
	bo BufferObject
}

func NewLiveEditProgram(vert, frag string) Program {
	return &LiveEditProgram{
		vertFilename: vert,
		fragFilename: frag,
		vertHistory:  make([]string, 0),
		fragHistory:  make([]string, 0),
	}
}

func (self *LiveEditProgram) watch() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	// defer watcher.Close()

	log.Println("watching: ", self.fragFilename)
	watcher.Add(self.vertFilename)
	watcher.Add(self.fragFilename)

	// initial file load
	vert, err := os.ReadFile(self.vertFilename)
	frag, err := os.ReadFile(self.fragFilename)
	if err == nil {
		self.compile(string(vert), string(frag))
	}

	self.watcher = watcher
}

func (self *LiveEditProgram) Load(window *glfw.Window, bo BufferObject) {
	self.Window = window
	self.bo = bo
	self.width, self.height = window.GetFramebufferSize()

	// create textures
	img := *image.NewRGBA(image.Rect(0, 0, self.width, self.height))
	self.texture = LoadTexture(&img)

	// load with blank shaders
	self.shader = MustCompileShader(VertexShader, FragShader, self.bo)

	self.watch()
}

func (self *LiveEditProgram) compile(vert, frag string) {
	newShader, err := CompileShader(vert, frag, self.bo)
	if err != nil {
		log.Println("error", err)
		return
	}

	if self.vert != vert {
		self.vert = vert
		self.vertHistory = append(self.vertHistory, frag)
	}

	if self.frag != frag {
		self.frag = frag
		self.fragHistory = append(self.fragHistory, frag)
	}

	self.shader = newShader
	self.paused = false
}

func (self *LiveEditProgram) run(t float64) {
	mx, my := self.Window.GetCursorPos()

	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	gl.BindVertexArray(self.bo.VAO())
	gl.Clear(gl.COLOR_BUFFER_BIT)

	self.texture.Activate(gl.TEXTURE0)

	self.shader.Use().
		Uniform1i("u_frame", int32(self.frame)).
		Uniform1f("u_time", float32(t)).
		Uniform2f("u_mouse", float32(mx), float32(self.height)-float32(my)).
		Uniform2f("u_resolution", float32(self.width), float32(self.height))
	self.bo.Draw()
}

func (self *LiveEditProgram) Render(t float64) {
	select {
	case event, ok := <-self.watcher.Events:
		if !ok {
			break
		}

		// have to check "rename" for some reason?!
		if event.Op == fsnotify.Create || event.Op == fsnotify.Write || event.Op == fsnotify.Rename {
			log.Println("modified file:", event.Name)
			self.watcher.Add(event.Name)

			vert := self.vert
			frag := self.frag

			// read file in
			data, err := os.ReadFile(event.Name)
			if err != nil {
				log.Println("error:", err)
				break
			}

			// if strings.ContainsAny
			enb := filepath.Base(event.Name)
			vfb := filepath.Base(self.vertFilename)
			ffb := filepath.Base(self.fragFilename)
			switch enb {
			case vfb:
				vert = string(data)
			case ffb:
				frag = string(data)
			}

			// attempt to create new shader
			self.compile(string(vert), string(frag))
		}
	default:
		if self.paused {
			self.bo.Draw()
			return
		}

		self.frame = self.frame + 1
		self.run(t)
	}
}

func (self *LiveEditProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeySpace && action == glfw.Release {
		self.paused = !self.paused
	}
}

func (self *LiveEditProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.width, self.height = width, height
	self.texture.Resize(width, height)
}
