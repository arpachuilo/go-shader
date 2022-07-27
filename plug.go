package main

import (
	"fmt"

	"github.com/go-gl/glfw/v3.3/glfw"
)

var BuildDate = ""

func PlugRender(kill <-chan bool, window *glfw.Window) {
	fmt.Println(BuildDate)
	renderer := NewRenderer(window)
	renderer.Setup()
	renderer.Start(kill)
}
