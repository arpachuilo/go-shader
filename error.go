package engine

import (
	"fmt"
	"runtime"

	"github.com/go-gl/gl/v4.1-core/gl"
)

func Trace() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(3, pc)
	frames := runtime.CallersFrames(pc[:n])
	frame, _ := frames.Next()
	return fmt.Sprintf("%s:%d %s\n", frame.File, frame.Line, frame.Function)
}

func DumpGLError(message string) {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		msg := fmt.Sprintf("%s: %v\n %v\n", message, Trace(), err)
		panic(msg)
	}
}

func PanicOnGLError(message string) {
	err := gl.GetError()
	if err != gl.NO_ERROR {
		msg := fmt.Sprintf("%s: %v\n %v\n", message, Trace(), err)
		panic(msg)
	}
}
