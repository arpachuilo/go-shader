package main

import (
	"errors"
	"fmt"
	"image"
	"image/png"
	"os"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

var CaptureCmd = "capture"

// Renderer handles running our programs
type Renderer struct {
	Program Program
	Window  *glfw.Window

	Tick          <-chan time.Time
	RefreshRate   float64
	Width, Height int

	KeyPressDetection *KeyPressDetection
	Cmds              CmdChannels

	Recorder *Recorder

	vao uint32
	vbo uint32
}

// NewRenderer Create new renderer
func NewRenderer(window *glfw.Window) *Renderer {
	return &Renderer{
		Program: nil,
		Window:  window,

		KeyPressDetection: NewKeyPressDetection(),
		Cmds:              NewCmdChannels(),

		Recorder: NewRecorder(window),
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
	r.SetTickRate(r.RefreshRate)

	// register key press channels
	r.Cmds.Register(CaptureCmd)

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

func (r *Renderer) SetTickRate(rr float64) {
	r.Tick = time.Tick(time.Duration(1000/rr) * time.Millisecond)
}

var frames = 0
var lastTime time.Time

func (r *Renderer) Start() {
	r.Program = NewLifeProgram()
	r.Program.Load(r.Window, r.vao, r.vbo)
	for !r.Window.ShouldClose() {
		select {
		// capture frame
		case <-r.Cmds[CaptureCmd]:
			r.Capture()
		// frame limiter
		case <-r.Tick:
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

			// run
			r.Program.Render(t)

			// maintenance
			r.Window.SwapBuffers()
			glfw.PollEvents()

			// record
			if r.Recorder.On {
				r.Recorder.Capture()
			}
		}
	}

	glfw.Terminate()
}

func (r *Renderer) ResizeCallback(w *glfw.Window, width int, height int) {
	r.Width, r.Height = width, height

	r.Program.ResizeCallback(w, width, height)
}

func (r *Renderer) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	r.Program.KeyCallback(w, key, scancode, action, mods)

	// call renderer key callbacks
	if glfw.Release == action {
		if key == glfw.KeyP {
			r.Cmds.Issue(CaptureCmd)
		}

		if key == glfw.KeyQ {
			if !r.Recorder.On {
				r.Recorder.Start()
			} else {
				r.Recorder.End()
			}
		}
	}
}

func (r *Renderer) Capture() error {
	// create sub-folders
	subFolder := "cap"
	folder := fmt.Sprintf("screencaptures/%v/", subFolder)
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
	w, h := r.Window.GetFramebufferSize()
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

	return nil
}

type CmdChannels map[string](chan interface{})

func NewCmdChannels() CmdChannels {
	return make(map[string](chan interface{}))
}

func (lc CmdChannels) Issue(key string) error {
	if _, ok := lc[key]; !ok {
		return errors.New("cmd is not registered")
	}

	go func(chan interface{}) {
		lc[key] <- nil
	}(lc[key])

	return nil
}

func (lc CmdChannels) Register(key string) error {
	if _, ok := lc[key]; ok {
		return errors.New("cmd already registered to channel")
	}

	lc[key] = make(chan interface{})
	return nil
}

func (lc CmdChannels) Unregister(key string) {
	delete(lc, key)
}
