package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/arpachuilo/go-registrable"
	"github.com/gen2brain/beeep"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var quadVertices = []float32{
	// positions   // texCoords
	-1.0, 1.0, 0.0, 1.0,
	-1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 0.0,

	-1.0, 1.0, 0.0, 1.0,
	1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, 1.0, 1.0,
}

var CaptureCmd = "capture"

// Renderer handles running our programs
type Renderer struct {
	Program        Program
	Window         *glfw.Window
	wPosX, wPosY   int
	wSizeX, wSizeY int

	Tick              *time.Ticker
	RefreshRate       float64
	UnlockedFrameRate bool
	Width, Height     int

	KeyPressDetection *KeyPressDetection
	Cmds              CmdChannels

	Recorder *Recorder

	vao uint32
	vbo uint32

	*KeyRegister
}

// NewRenderer Create new renderer
func NewRenderer(window *glfw.Window) *Renderer {
	r := &Renderer{
		Program: nil,
		Window:  window,

		KeyPressDetection: NewKeyPressDetection(),
		Cmds:              NewCmdChannels(),

		Recorder: NewRecorder(window),
	}

	r.KeyRegister = NewKeyRegister()
	registrable.RegisterMethods[KeyCallbackRegistration](r)

	return r
}

func (self *Renderer) Setup() {
	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	// get current resolution
	self.Width, self.Height = self.Window.GetSize()

	// get refresh rate
	self.RefreshRate = float64(glfw.GetPrimaryMonitor().GetVideoMode().RefreshRate)
	self.Tick = time.NewTicker(time.Duration(1000/self.RefreshRate) * time.Millisecond)

	// register key press channels
	self.Cmds.Register(CaptureCmd)

	// print some info
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
	fmt.Println("Refresh rate", self.RefreshRate)

	// register callbacks
	self.Window.SetKeyCallback(self.KeyCallback)
	self.Window.SetSizeCallback(self.ResizeCallback)

	// configure the vertex data
	gl.GenVertexArrays(1, &self.vao)
	gl.BindVertexArray(self.vao)

	gl.GenBuffers(1, &self.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, self.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, len(quadVertices)*4, gl.Ptr(quadVertices), gl.STATIC_DRAW)

	// Configure global settings
	gl.ClearColor(1.0, 1.0, 1.0, 1.0)
}

func (self *Renderer) SetTickRate(rr float64) {
	if rr <= 0.0 {
		self.Tick.Reset(1)
		glfw.SwapInterval(0)
	} else {
		self.Tick.Reset(time.Duration(1000/rr) * time.Millisecond)
		glfw.SwapInterval(1)
	}

	self.UnlockedFrameRate = rr != self.RefreshRate
}

var frames = 0
var lastTime time.Time

func (self *Renderer) Start(kill <-chan bool) {
	// self.Program = NewTurtleProgram()
	self.Program = NewPongProgram()
	self.Program.Load(self.Window, self.vao, self.vbo)
	for !self.Window.ShouldClose() {
		select {
		// kill
		case <-kill:
			return
		// capture frame
		case <-self.Cmds[CaptureCmd]:
			self.Capture()
		// frame limiter
		case <-self.Tick.C:
			t := glfw.GetTime()
			frames++
			currentTime := time.Now()
			delta := currentTime.Sub(lastTime)
			if delta > time.Second {
				fps := frames / int(delta.Seconds())
				self.Window.SetTitle(fmt.Sprintf("FPS: %v", fps))

				lastTime = currentTime
				frames = 0
			}

			// run
			self.Program.Render(t)

			// maintenance
			self.Window.SwapBuffers()
			glfw.PollEvents()

			// record
			if self.Recorder.On {
				self.Recorder.Capture()
			}
		}
	}
}

func (self *Renderer) ResizeCallback(w *glfw.Window, width int, height int) {
	self.Width, self.Height = width, height

	self.Program.ResizeCallback(w, width, height)
}

func (self *Renderer) SwitchToLife() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF1,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			self.Program = NewLifeProgram()
			self.Program.Load(self.Window, self.vao, self.vbo)
		},
	}
}

func (self *Renderer) SwitchToMandelbrot() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF2,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			self.Program = NewMandelbrotProgram()
			self.Program.Load(self.Window, self.vao, self.vbo)
		},
	}
}

func (self *Renderer) SwitchToJulia() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF3,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			self.Program = NewJuliaProgram()
			self.Program.Load(self.Window, self.vao, self.vbo)
		},
	}
}

func (self *Renderer) SwitchToLiveEdit() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF4,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			self.Program = NewLiveEditProgram("./assets/shaders/live_edit.glsl")
			self.Program.Load(self.Window, self.vao, self.vbo)
		},
	}
}

func (self *Renderer) SwitchToTurtle() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF5,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			self.Program = NewTurtleProgram()
			self.Program.Load(self.Window, self.vao, self.vbo)
		},
	}
}

func (self *Renderer) SwitchToPong() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF6,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			self.Program = NewPongProgram()
			self.Program.Load(self.Window, self.vao, self.vbo)
		},
	}
}

func (self *Renderer) UnlockFramerate() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyU,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if self.UnlockedFrameRate {
				self.SetTickRate(self.RefreshRate)
			} else {
				self.SetTickRate(0)
			}
		},
	}
}

func (self *Renderer) CloseProgram() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyW,
		mods:   glfw.ModSuper,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			w.SetShouldClose(true)
		},
	}
}

func (self *Renderer) Screencapture() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyP,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			self.Cmds.Issue(CaptureCmd)
		},
	}
}

func (self *Renderer) Record() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyQ,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if !self.Recorder.On {
				self.Recorder.Start()
			} else {
				self.Recorder.End()
			}
		},
	}
}

func (self *Renderer) ToggleCursor() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyEscape,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if self.Window.GetInputMode(glfw.CursorMode) == glfw.CursorDisabled {
				self.Window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
			} else {
				self.Window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
			}
		},
	}
}

func (self *Renderer) KeyCallback(
	w *glfw.Window,
	key glfw.Key,
	scancode int,
	action glfw.Action,
	mods glfw.ModifierKey,
) {
	self.Program.KeyCallback(w, key, scancode, action, mods)
	self.KeyRegister.KeyCallback(w, key, scancode, action, mods)
}

func (self *Renderer) Capture() error {
	// create sub-folders
	// folder := fmt.Sprintf("screencaptures/%v/", subFolder)
	folder := "screencaptures/"
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, 0700)
	}

	// create file
	name := folder + time.Now().Format("20060102150405") + ".png"
	f, err := os.Create(name)
	if err != nil {
		return err
	}

	// create image
	w, h := self.Window.GetFramebufferSize()
	img := *image.NewRGBA(image.Rect(0, 0, w, h))

	// set active frame buffer as main one
	gl.ReadPixels(
		0, 0,
		int32(w), int32(h),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(img.Pix),
	)

	// encode png
	fmt.Println("Saving", name)
	if err = png.Encode(f, &img); err != nil {
		return err
	}

	// cleanup
	if err = f.Close(); err != nil {
		return err
	}

	err = beeep.Notify("Screenshot Captured!", name, "applet.icns")
	if err != nil {
		return err
	}

	return nil
}

type CmdChannels map[string](chan interface{})

func NewCmdChannels() CmdChannels {
	return make(map[string](chan interface{}))
}

func (self CmdChannels) Issue(key string) error {
	if _, ok := self[key]; !ok {
		return errors.New("cmd is not registered")
	}

	go func(chan interface{}) {
		self[key] <- nil
	}(self[key])

	return nil
}

func (self CmdChannels) Register(key string) error {
	if _, ok := self[key]; ok {
		return errors.New("cmd already registered to channel")
	}

	self[key] = make(chan interface{})
	return nil
}

func (self CmdChannels) Unregister(key string) {
	delete(self, key)
}
