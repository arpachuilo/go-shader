//go:build darwin
// +build darwin

package main

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework Cocoa
#import <Cocoa/Cocoa.h>

int SetTitleTransparency(void *s, bool titleTransparency) {
  NSWindow *wnd = ((__unsafe_unretained NSWindow*)(s));

  // nice fullcontent sizing
	wnd.titlebarAppearsTransparent =  titleTransparency;
  wnd.styleMask = wnd.styleMask | NSWindowStyleMaskFullSizeContentView;

  // make that transparent buffer work
  wnd.opaque = false;
  wnd.backgroundColor = NSColor.clearColor;

  return 0;
}
*/
import "C"
import "github.com/go-gl/glfw/v3.3/glfw"

func SetupOSHint() {
	glfw.WindowHint(glfw.CocoaRetinaFramebuffer, glfw.True)
	glfw.WindowHintString(glfw.CocoaFrameNAME, "go-opengl")
}

func SetupOSWindow(window *glfw.Window, transparency bool) {
	C.SetTitleTransparency(window.GetCocoaWindow(), C.bool(transparency))
}
