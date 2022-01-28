package main

import (
	"image"

	"github.com/go-gl/glfw/v3.3/glfw"
)

type NoopProgram struct{}

func NewNoopProgram() Program {
	p := NoopProgram{}
	return &p
}

func (p *NoopProgram) Render() bool {
	return false
}

func (c *NoopProgram) ResizeCallback(img *image.RGBA, width int, height int) {}

func (c *NoopProgram) KeyCallback(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
}
