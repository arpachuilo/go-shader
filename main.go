package main

import (
	"image"
	"image/color"
	"log"
	"plugin"
	"runtime"

	"github.com/fsnotify/fsnotify"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// window settings
var windowWidth = 1280
var windowHeight = 720

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

type PluggableRender func(<-chan bool, *glfw.Window)

// TODO: refactor into packages
// TODO: make registrable keys to print which keys do what
// TODO: ability to render text overlays

func main() {
	// init glfw
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	// setup window
	// opengl
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// basic
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.CenterCursor, glfw.True)

	// mac os
	glfw.WindowHint(glfw.CocoaRetinaFramebuffer, glfw.True)
	glfw.WindowHintString(glfw.CocoaFrameNAME, "go-opengl")

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Visuals", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// disable cursor
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for x := 0; x < 64; x++ {
		for y := 0; y < 64; y++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	window.SetIcon([]image.Image{img})

	kill := make(chan bool)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	watcher.Add("./bin/plugins/")

	// check for updates to plugin
	latestMod := "./bin/plugins/plug.so"
	go func() {
		for {
			select {
			case event, ok := <-watcher.Events:
				if ok {
					if event.Op == fsnotify.Create {
						latestMod = event.Name
						kill <- true
					}
				}
			}
		}
	}()

	for {
		Render(latestMod, kill, window)
	}
}

func Render(latestMod string, kill <-chan bool, window *glfw.Window) {
	// load module
	// 1. open the so file to load the symbols
	plug, err := plugin.Open(latestMod)
	if err != nil {
		panic(err)
	}

	// 2. look up a symbol (an exported function or variable)
	symRenderer, err := plug.Lookup("PlugRender")
	if err != nil {
		panic(err)
	}

	// 3. Assert that loaded symbol is of a desired type
	var plugRender PluggableRender
	plugRender, ok := symRenderer.(func(<-chan bool, *glfw.Window))
	if !ok {
		panic("unexpected type from module symbol")
	}

	// 4. use the module
	plugRender(kill, window)
}
