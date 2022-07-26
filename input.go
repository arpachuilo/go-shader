package main

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

type KeyCallbackRegistration struct {
	mods        glfw.ModifierKey
	action      glfw.Action
	key         glfw.Key
	callback    glfw.KeyCallback
	description string
}

type KeyCallback struct {
	callback    glfw.KeyCallback
	description string
}

type KeyRegister struct {
	// TODO: allow multiple callbacks for a key?
	// TODO: used orderer registration?
	callbacks map[glfw.Action]map[glfw.Key]map[glfw.ModifierKey]KeyCallback
}

func (self *KeyRegister) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if a, ok := self.callbacks[action]; ok {
		if b, ok := a[key]; ok {
			if c, ok := b[mods]; ok {
				c.callback(w, key, scancode, action, mods)
			}
		}
	}
}

func NewKeyRegister() *KeyRegister {
	kr := &KeyRegister{
		callbacks: make(map[glfw.Action]map[glfw.Key]map[glfw.ModifierKey]KeyCallback),
	}

	return kr
}

func (self *KeyRegister) Register(r KeyCallbackRegistration) {
	if self.callbacks[r.action] == nil {
		self.callbacks[r.action] = make(map[glfw.Key]map[glfw.ModifierKey]KeyCallback)
	}

	if self.callbacks[r.action][r.key] == nil {
		self.callbacks[r.action][r.key] = make(map[glfw.ModifierKey]KeyCallback)
	}

	self.callbacks[r.action][r.key][r.mods] = KeyCallback{
		callback: r.callback,
	}
}

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
func (self *KeyPressDetection) HandleKeyPress(key glfw.Key, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		{
			self.Down[key] = true
		}

	case glfw.Release:
		{
			self.Down[key] = false
		}
	}
}

type MouseDelta struct {
	previousX, previousY float64

	Scale float64
}

func NewMouseDelta(scale float64) *MouseDelta {
	return &MouseDelta{
		previousX: -1,
		previousY: -1,
		Scale:     scale,
	}
}

func (self *MouseDelta) DeltaX(currentX float64) float64 {
	deltaX := currentX - self.previousX
	if self.previousX == -1 {
		deltaX = 0
	}

	self.previousX = currentX
	return (deltaX) * self.Scale
}

func (self *MouseDelta) DeltaY(currentY float64) float64 {
	deltaY := currentY - self.previousY
	if self.previousY == -1 {
		deltaY = 0
	}

	self.previousY = currentY
	return (deltaY) * self.Scale
}

func (self *MouseDelta) Delta(currentX, currentY float64) (float64, float64) {
	return self.DeltaX(currentX), self.DeltaY(currentY)
}
