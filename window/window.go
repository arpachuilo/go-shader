package window

import (
	"fmt"
	"image"
	"image/color"
	"log"
	"plugin"
	"runtime"
	"runtime/debug"

	"github.com/fsnotify/fsnotify"
	"github.com/go-gl/glfw/v3.3/glfw"
)

// window settings
var windowWidth = 1280
var windowHeight = 720

type PlugRender func(<-chan bool, *glfw.Window)

// TODO: make registrable keys to print which keys do what
// TODO: ability to render text overlays
// TODO: command to generate embedded asset listings

type Window struct {
	*glfw.Window

	pluginFolder, pluginFile, pluginLatest string
}

func init() {
	runtime.LockOSThread()
}

func NewWindow() *Window {
	// init glfw
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}

	// setup window
	// opengl
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// basic
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.CenterCursor, glfw.True)
	glfw.WindowHint(glfw.TransparentFramebuffer, glfw.True)

	// setup os specific hints
	SetupOSHint()

	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Visuals", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()

	// setup os specific window
	SetupOSWindow(window, true)

	// disable cursor
	window.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)

	img := image.NewNRGBA(image.Rect(0, 0, 64, 64))
	for x := 0; x < 64; x++ {
		for y := 0; y < 64; y++ {
			img.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	window.SetIcon([]image.Image{img})

	return &Window{Window: window}
}

func (self *Window) Close() {
	glfw.Terminate()
}

func (self *Window) HotWindow(_pluginFolder, _pluginFile string) {
	self.pluginFolder = _pluginFolder
	self.pluginFile = _pluginFile
	self.pluginLatest = _pluginFolder + _pluginFile

	kill := make(chan bool)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	watcher.Add(self.pluginFolder)

	// check for updates to plugin
	latestMod := self.pluginLatest
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

	successiveFails := 0
	for !self.Window.ShouldClose() {
		err := self.HotRender(latestMod, kill)
		if err != nil {
			successiveFails++
			latestMod = self.pluginLatest
			fmt.Println(err)
		} else {
			successiveFails = 0
		}

		if successiveFails > 3 {
			self.SetShouldClose(true)
		}
	}
}

func (self *Window) HotRender(latestMod string, kill <-chan bool) (err error) {
	// load module
	// 1. open the so file to load the symbols
	plug, err := plugin.Open(latestMod)
	if err != nil {
		return err
	}

	// 2. look up a symbol (an exported function or variable)
	symRenderer, err := plug.Lookup("HotProgramFn")
	if err != nil {
		return err
	}

	// 3. Assert that loaded symbol is of a desired type
	var plugRender PlugRender
	plugRender, ok := symRenderer.(func(<-chan bool, *glfw.Window))
	if !ok {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			debug.PrintStack()
			err = fmt.Errorf("recovered: %v", r)
		} else {
			self.pluginLatest = latestMod
		}
	}()

	// 4. use the module
	fmt.Printf("Running Plug: %v\n", latestMod)
	plugRender(kill, self.Window)

	return nil
}
