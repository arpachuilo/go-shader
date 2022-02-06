package main

import "github.com/go-gl/glfw/v3.3/glfw"

type Program interface {
	Render(t float64)
	Load(window *glfw.Window, vao, vbo uint32)
	ResizeCallback(w *glfw.Window, width int, height int)
	KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)
}
