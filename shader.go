package main

import (
	"fmt"
	"math"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Shader uint32

// MustCompileShader create a new shader program that must compile.
func MustCompileShader(vertexShaderSource, fragmentShaderSource string, bo BufferObject) Shader {
	shader, err := CompileShader(vertexShaderSource, fragmentShaderSource, bo)
	if err != nil {
		panic(err)
	}

	return shader
}

// CompileShader create a new shader program.
// TODO: stop hard coding to the quad
func CompileShader(vertexShaderSource, fragmentShaderSource string, bo BufferObject) (Shader, error) {
	vertexShader, err := compileShader(vertexShaderSource+"\x00", gl.VERTEX_SHADER)
	if err != nil {
		return Shader(math.MaxUint32), err
	}

	fragmentShader, err := compileShader(fragmentShaderSource+"\x00", gl.FRAGMENT_SHADER)
	if err != nil {
		return Shader(math.MaxUint32), err
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

		return Shader(math.MaxUint32), fmt.Errorf("failed to link program: %v", log)
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
	gl.VertexAttribPointerWithOffset(vertCoords, 2, gl.FLOAT, false, bo.GetVertices().VertexSize(), 0)

	// bind texture coordinates
	texCoords := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(texCoords)
	gl.VertexAttribPointerWithOffset(texCoords, 2, gl.FLOAT, false, bo.GetVertices().VertexSize(), uintptr(bo.GetVertices().TexOffset()))

	return Shader(program), nil
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
	gl.UseProgram(uint32(self))
	gl.DeleteProgram(uint32(self))
}

func (self Shader) Apply(applyFN func(Shader) Shader) Shader {
	return applyFN(self)
}

func (self Shader) Use() Shader {
	gl.UseProgram(uint32(self))
	return self
}

func (self Shader) Uniform1d(name string, value float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform1d(uint32(self), location, value)
	return self
}

func (self Shader) Uniform1dv(name string, values []float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform1dv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform1f(name string, value float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform1f(uint32(self), location, value)
	return self
}

func (self Shader) Uniform1fv(name string, values []float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform1fv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform1i(name string, value int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform1i(uint32(self), location, value)
	return self
}

func (self Shader) Uniform1iv(name string, values []int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform1iv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform2d(name string, v0, v1 float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform2d(uint32(self), location, v0, v1)
	return self
}

func (self Shader) Uniform2dv(name string, values []float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform2dv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform2f(name string, v0, v1 float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform2f(uint32(self), location, v0, v1)
	return self
}

func (self Shader) Uniform2fv(name string, values []float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform2fv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform2i(name string, v0, v1 int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform2i(uint32(self), location, v0, v1)
	return self
}

func (self Shader) Uniform2iv(name string, values []int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform2iv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform3d(name string, v0, v1, v2 float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform3d(uint32(self), location, v0, v1, v2)
	return self
}

func (self Shader) Uniform3dv(name string, values []float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform3dv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform3f(name string, v0, v1, v2 float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform3f(uint32(self), location, v0, v1, v2)
	return self
}

func (self Shader) Uniform3fv(name string, values []float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform3fv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform3i(name string, v0, v1, v2 int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform3i(uint32(self), location, v0, v1, v2)
	return self
}

func (self Shader) Uniform3iv(name string, values []int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform3iv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform4d(name string, v0, v1, v2, v3 float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform4d(uint32(self), location, v0, v1, v2, v3)
	return self
}

func (self Shader) Uniform4dv(name string, values []float64) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform4dv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform4f(name string, v0, v1, v2, v3 float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform4f(uint32(self), location, v0, v1, v2, v3)
	return self
}

func (self Shader) Uniform4fv(name string, values []float32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform4fv(uint32(self), location, int32(len(values)), &values[0])
	return self
}

func (self Shader) Uniform4i(name string, v0, v1, v2, v3 int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform4i(uint32(self), location, v0, v1, v2, v3)
	return self
}

func (self Shader) Uniform4iv(name string, values []int32) Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(uint32(self), gl.Str(attr))
	gl.ProgramUniform4iv(uint32(self), location, int32(len(values)), &values[0])
	return self
}
