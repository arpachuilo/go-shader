package main

import "github.com/go-gl/glfw/v3.3/glfw"

var BuildDate = ""

func PlugRender(kill <-chan bool, window *glfw.Window) {
	renderer := NewRenderer(window)
	renderer.Setup()
	renderer.Start(kill)
}
