package main

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

type NoopProgram struct {
}

func NewNoopProgram() Program {
	return &NoopProgram{}
}

func (self *NoopProgram) Load(window *glfw.Window, vao, vbo uint32) {
}

func (self *NoopProgram) Render(t float64) {

}

func (self *NoopProgram) ResizeCallback(w *glfw.Window, width int, height int) {

}

func (self *NoopProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {

}
