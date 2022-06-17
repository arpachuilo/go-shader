package main

import (
	"image"
	"image/color"
	"log"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// window settings
var windowWidth = 1280
var windowHeight = 720

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

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

	renderer := NewRenderer(window)
	renderer.Setup()
	renderer.Start()
}

var quadVertices = []float32{
	// positions   // texCoords
	-1.0, 1.0, 0.0, 1.0,
	-1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 0.0,

	-1.0, 1.0, 0.0, 1.0,
	1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, 1.0, 1.0,
}
