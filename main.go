package main

import (
	"log"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// window settings
var windowWidth = 800
var windowHeight = 600

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
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)

	// glfw.WindowHint(glfw.CocoaRetinaFramebuffer, glfw.True)
	// glfw.WindowHint(glfw.SRGBCapable, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Visuals", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	// TODO: Add program icon
	// window.SetIcon()

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
