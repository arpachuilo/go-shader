package engine

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

var BuildDate = ""
var HotProgram Program

func HotRender(kill <-chan bool, window *glfw.Window) {
	fmt.Println(BuildDate)

	if HotProgram == nil {
		panic("hot program not set")
	}

	NewRenderer(window, HotProgram).Run(kill)
}

var CaptureCmd = "capture"

// Renderer handles running our programs
type Renderer struct {
	Program Program
	Window  *glfw.Window

	PauseBufferSwap   bool
	Wireframe         bool
	Tick              *time.Ticker
	RefreshRate       float64
	UnlockedFrameRate bool
	Width, Height     int

	Cmds CmdChannels

	*Recorder
	*KeyPressDetection
	*KeyRegister
	*Cleaner
}

// NewRenderer Create new renderer
func NewRenderer(window *glfw.Window, program Program) *Renderer {
	r := &Renderer{
		Program: nil,
		Window:  window,

		KeyPressDetection: NewKeyPressDetection(),
		Cmds:              NewCmdChannels(),

		Recorder: NewRecorder(window),

		Cleaner: &Cleaner{},
	}

	r.KeyRegister = NewKeyRegister()
	registrable.RegisterMethods[KeyCallbackRegistration](r)

	// Initialize Glow
	if err := gl.Init(); err != nil {
		panic(err)
	}

	// get current resolution
	r.Width, r.Height = r.Window.GetFramebufferSize()

	// get refresh rate
	r.RefreshRate = float64(glfw.GetPrimaryMonitor().GetVideoMode().RefreshRate)
	r.Tick = time.NewTicker(time.Duration(1000/r.RefreshRate) * time.Millisecond)

	// register key press channels
	r.Cmds.Register(CaptureCmd)

	// print some info
	version := gl.GoStr(gl.GetString(gl.VERSION))
	fmt.Println("OpenGL version", version)
	fmt.Println("Refresh rate", r.RefreshRate)

	// register callbacks
	r.Window.SetKeyCallback(r.KeyCallback)
	r.Window.SetSizeCallback(r.ResizeCallback)

	// Configure global settings
	gl.ColorMask(true, true, true, true)
	gl.ClearColor(0.0, 0.0, 0.0, 0.0)
	gl.Clear(gl.COLOR_BUFFER_BIT)

	r.Program = program
	r.Program.LoadR(r)
	return r
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

func (self *Renderer) Run(kill <-chan bool) {
	gl.Viewport(0, 0, int32(self.Width), int32(self.Height))
	defer self.Cleaner.Run()

	frames := 0.0
	previousTime := glfw.GetTime()
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
			currentTime := glfw.GetTime()
			frames++
			delta := currentTime - previousTime
			if delta > 1.0 {
				fps := frames / delta
				self.Window.SetTitle(fmt.Sprintf("%.2f FPS @ %v x %v", fps, self.Width, self.Height))

				previousTime = currentTime
				frames = 0
			}

			// run
			self.Program.Render(currentTime)

			// maintenance
			if !self.PauseBufferSwap {
				self.Window.SwapBuffers()
			}

			glfw.PollEvents()

			// record
			if self.Recorder.On {
				self.Recorder.Capture()
			}
		}
	}
}

func (self *Renderer) ResizeCallback(w *glfw.Window, width int, height int) {
	self.Width, self.Height = w.GetFramebufferSize()
	gl.Viewport(0, 0, int32(self.Width), int32(self.Height))

	self.Program.ResizeCallback(w, self.Width, self.Height)
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

func (self *Renderer) UnlockFramerate() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF1,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if self.UnlockedFrameRate {
				self.SetTickRate(self.RefreshRate)
			} else {
				self.SetTickRate(0)
			}
		},
	}
}

func (self *Renderer) Screencapture() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF2,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			self.Cmds.Issue(CaptureCmd)
		},
	}
}

func (self *Renderer) Record() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF3,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if !self.Recorder.On {
				self.Recorder.Start()
			} else {
				self.Recorder.End()
			}
		},
	}
}

func (self *Renderer) ToggleAlwaysOnTop() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF6,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			if self.Window.GetAttrib(glfw.Floating) == glfw.True {
				self.Window.SetAttrib(glfw.Floating, glfw.False)
			} else {
				self.Window.SetAttrib(glfw.Floating, glfw.True)
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

func (self *Renderer) ToggleWireframe() registrable.Registration {
	return KeyCallbackRegistration{
		action: glfw.Release,
		key:    glfw.KeyF10,
		callback: func(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
			self.Wireframe = !self.Wireframe

			if self.Wireframe {
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.LINE)
			} else {
				gl.PolygonMode(gl.FRONT_AND_BACK, gl.FILL)
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
