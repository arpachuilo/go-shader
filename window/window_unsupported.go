//go:build !darwin
// +build !darwin

package window

import "github.com/go-gl/glfw/v3.3/glfw"

func SetupOSHint() {

}

func SetupOSWindow(window *glfw.Window, transparency bool) {
}
