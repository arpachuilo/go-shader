package engine

import "github.com/go-gl/glfw/v3.3/glfw"

type Program interface {
	Render(t float64)
	LoadR(*Renderer)
	Load(window *glfw.Window)
	ResizeCallback(w *glfw.Window, width int, height int)
	KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)
}
