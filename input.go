package main

import "github.com/go-gl/glfw/v3.3/glfw"

// KeyPressDetection Allows detection of multiple down-presses at the same time
type KeyPressDetection struct {

	// Map of currently pressed keys
	Down map[glfw.Key]bool
}

// NewKeyPressDetection Create KeyPressDetection
func NewKeyPressDetection() *KeyPressDetection {
	return &KeyPressDetection{
		Down: make(map[glfw.Key]bool),
	}
}

// HandleKeyPress Handle whether or not a key press has been pressed/released. Call once per KeyCallback handler.
func (kpd *KeyPressDetection) HandleKeyPress(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		{
			kpd.Down[key] = true
		}

	case glfw.Release:
		{
			kpd.Down[key] = false
		}
	}
}
