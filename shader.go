package engine

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/mathgl/mgl32"
)

type HotShader struct {
	VertFilename string
	VertSource   string

	FragFilename string
	FragSource   string

	BufferObject BufferObject
	*Shader
}

type ShaderWatcher struct {
	*fsnotify.Watcher

	// map of watched files to their shader
	WatchedFiles map[string]*HotShader
}

func NewShaderWatcher() *ShaderWatcher {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}

	return &ShaderWatcher{
		Watcher:      watcher,
		WatchedFiles: make(map[string]*HotShader),
	}
}

func (self *ShaderWatcher) Add(shader *Shader, vertFilename, fragFilename string, bo BufferObject) {
	log.Println("watching: ", vertFilename)
	log.Println("watching: ", fragFilename)
	self.Watcher.Add(vertFilename)
	self.Watcher.Add(fragFilename)

	// initial file load
	// ignore read errors
	vert, err := os.ReadFile(vertFilename)
	frag, err := os.ReadFile(fragFilename)

	vfb := filepath.Clean(vertFilename)
	ffb := filepath.Clean(fragFilename)
	hs := HotShader{
		VertFilename: vfb,
		VertSource:   string(vert),

		FragFilename: ffb,
		FragSource:   string(frag),

		BufferObject: bo,
		Shader:       shader,
	}

	if err == nil {
		// attempt compile ignore any possible errors
		hs.Shader.Compile(hs.VertSource, hs.FragSource, bo)
	}

	self.WatchedFiles[vfb] = &hs
	self.WatchedFiles[ffb] = &hs
}

func (self *ShaderWatcher) Handle(event fsnotify.Event, ok bool) (*Shader, error) {
	if !ok {
		return nil, fmt.Errorf("ShaderWatcher.Handle: not okay")
	}

	// have to check "rename" for some reason?!
	if event.Op == fsnotify.Create || event.Op == fsnotify.Write || event.Op == fsnotify.Rename {
		log.Println("modified file:", event.Name)
		self.Watcher.Add(event.Name)

		hs, ok := self.WatchedFiles[event.Name]
		if !ok {
			return nil, fmt.Errorf("ShaderWatcher.Handle:  could not find watched file")
		}

		vert := hs.VertSource
		frag := hs.FragSource

		// read file in
		data, err := os.ReadFile(event.Name)
		if err != nil {
			return hs.Shader, err
		}

		// if strings.ContainsAny
		enb := filepath.Clean(event.Name)
		vfb := filepath.Clean(hs.VertFilename)
		ffb := filepath.Clean(hs.FragFilename)
		switch enb {
		case vfb:
			vert = string(data)
		case ffb:
			frag = string(data)
		}

		// attempt to create new shader
		err = hs.Compile(vert, frag, hs.BufferObject)
		if err != nil {
			return hs.Shader, err
		}

		if hs.VertSource != vert {
			hs.VertSource = vert
		}

		if hs.FragSource != frag {
			hs.FragSource = frag
		}

		return hs.Shader, nil
	}

	return nil, nil
}

type Shader struct {
	Program *uint32
}

func NewShader() *Shader {
	return &Shader{Program: nil}
}

// MustCompileShader create a new shader program that must compile.
func MustCompileShader(vertexShaderSource, fragmentShaderSource string, bo BufferObject) Shader {
	shader, err := CompileShader(vertexShaderSource, fragmentShaderSource, bo)
	if err != nil {
		panic(err)
	}

	return shader
}

// CompileShader create a new shader program.
func CompileShader(vertexShaderSource, fragmentShaderSource string, bo BufferObject) (Shader, error) {
	shader := Shader{}
	err := shader.Compile(vertexShaderSource, fragmentShaderSource, bo)
	return shader, err
}

func (self *Shader) Compile(vertexShaderSource, fragmentShaderSource string, bo BufferObject) error {
	vertexShader, err := compileShader(vertexShaderSource+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		return err
	}

	fragmentShader, err := compileShader(fragmentShaderSource+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		return err
	}

	program := gl.CreateProgram()
	gl.AttachShader(program, vertexShader)
	gl.AttachShader(program, fragmentShader)
	gl.LinkProgram(program)

	var status int32
	gl.GetProgramiv(program, gl.LINK_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetProgramiv(program, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(program, logLength, nil, gl.Str(log))

		return fmt.Errorf("failed to link program: %v", log)
	}

	gl.DetachShader(program, vertexShader)
	gl.DetachShader(program, fragmentShader)

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	// bind output color location
	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// bind buffer object
	// bind vertex coordinates
	vertCoords := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertCoords)
	gl.VertexAttribPointerWithOffset(
		vertCoords, bo.GetSize(), gl.FLOAT, false,
		bo.VertexSize(),
		uintptr(bo.PosOffset()),
	)

	// bind texture coordinates
	texCoords := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoords)
	gl.VertexAttribPointerWithOffset(
		texCoords, bo.GetSize(), gl.FLOAT, false,
		bo.VertexSize(),
		uintptr(bo.TexOffset()),
	)

	if self.Program != nil {
		self.Cleanup()
	}

	self.Program = &program

	return nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source)
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		log := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(log))

		return 0, fmt.Errorf("failed to compile %v: %v", source, log)
	}

	return shader, nil
}

func (self Shader) Cleanup() {
	gl.UseProgram(*self.Program)
	gl.DeleteProgram(*self.Program)
}

func (self Shader) Apply(applyFN func(Shader) Shader) Shader {
	return applyFN(self)
}

func (self Shader) Use() Shader {
	gl.UseProgram(*self.Program)
	return self
}

func (self Shader) Uniform1d(name string, value float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform1d(*self.Program, location, value)
	return self
}

func (self Shader) Uniform1dv(name string, values []float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform1dv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform1f(name string, value float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform1f(*self.Program, location, value)
	return self
}

func (self Shader) Uniform1fv(name string, values []float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform1fv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform1i(name string, value int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform1i(*self.Program, location, value)
	return self
}

func (self Shader) Uniform1iv(name string, values []int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform1iv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform2d(name string, v0, v1 float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform2d(*self.Program, location, v0, v1)
	return self
}

func (self Shader) Uniform2dv(name string, values []float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform2dv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform2f(name string, v0, v1 float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform2f(*self.Program, location, v0, v1)
	return self
}

func (self Shader) Uniform2fv(name string, values []float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform2fv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform2i(name string, v0, v1 int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform2i(*self.Program, location, v0, v1)
	return self
}

func (self Shader) Uniform2iv(name string, values []int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform2iv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform3d(name string, v0, v1, v2 float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform3d(*self.Program, location, v0, v1, v2)
	return self
}

func (self Shader) Uniform3dv(name string, values []float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform3dv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform3f(name string, v0, v1, v2 float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform3f(*self.Program, location, v0, v1, v2)
	return self
}

func (self Shader) Uniform3fv(name string, values []float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform3fv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform3i(name string, v0, v1, v2 int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform3i(*self.Program, location, v0, v1, v2)
	return self
}

func (self Shader) Uniform3iv(name string, values []int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform3iv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform4d(name string, v0, v1, v2, v3 float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform4d(*self.Program, location, v0, v1, v2, v3)
	return self
}

func (self Shader) Uniform4dv(name string, values []float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform4dv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform4f(name string, v0, v1, v2, v3 float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform4f(*self.Program, location, v0, v1, v2, v3)
	return self
}

func (self Shader) Uniform4fv(name string, values []float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform4fv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform4i(name string, v0, v1, v2, v3 int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform4i(*self.Program, location, v0, v1, v2, v3)
	return self
}

func (self Shader) Uniform4iv(name string, values []int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniform4iv(*self.Program, location, int32(len(values)), &values[0])
	return self
}

func (self Shader) UniformMatrix4fv(name string, values *mgl32.Mat4) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(*self.Program, gl.Str(attr))
	gl.ProgramUniformMatrix4fv(*self.Program, location, 1, false, &values[0])
	return self
}
