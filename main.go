package main

import (
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"os"
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

var pluginPath string = "./bin/plugins/"
var lastWorkingMod string = pluginPath + "plug.so"

// TODO: refactor into packages
// TODO: make registrable keys to print which keys do what
// TODO: ability to render text overlays
// TODO: command to generate embedded asset listings

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
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

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

	kill := make(chan bool)

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	defer watcher.Close()

	watcher.Add(pluginPath)

	// check for updates to plugin
	latestMod := lastWorkingMod
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
	for !window.ShouldClose() {
		err := Render(latestMod, kill, window)
		if err != nil {
			successiveFails++
			latestMod = lastWorkingMod
			fmt.Println(err)
		} else {
			successiveFails = 0
		}

		if successiveFails > 3 {
			window.SetShouldClose(true)
		}
	}
}

func copy(src, dst string, BUFFERSIZE int64) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file.", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if err != nil {
		panic(err)
	}

	buf := make([]byte, BUFFERSIZE)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}
		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	return err
}

func Render(latestMod string, kill <-chan bool, window *glfw.Window) (err error) {
	// load module
	// 1. open the so file to load the symbols
	plug, err := plugin.Open(latestMod)
	if err != nil {
		return err
	}

	// 2. look up a symbol (an exported function or variable)
	symRenderer, err := plug.Lookup("PlugRender")
	if err != nil {
		return err
	}

	// 3. Assert that loaded symbol is of a desired type
	var plugRender PluggableRender
	plugRender, ok := symRenderer.(func(<-chan bool, *glfw.Window))
	if !ok {
		return err
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered: %v", r)
		} else {
			lastWorkingMod = latestMod

			// overwrite existing plugin.so with new valid one
			// err = copy(latestMod, pluginPath+"plug.so", 1024)
		}
	}()

	// 4. use the module
	fmt.Printf("Running Plug: %v\n", latestMod)
	plugRender(kill, window)

	return nil
}
