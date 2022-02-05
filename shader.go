package main

import (
	"fmt"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Shader struct {
	Program uint32
}

// MustCompileShader create a new shader program that must compile.
func MustCompileShader(vertexShaderSource, fragmentShaderSource string) *Shader {
	shader, err := CompileShader(vertexShaderSource, fragmentShaderSource)
	if err != nil {
		panic(err)
	}

	return shader
}

// CompileShader create a new shader program.
func CompileShader(vertexShaderSource, fragmentShaderSource string) (*Shader, error) {
	vertexShader, err := compileShader(vertexShaderSource, gl.VERTEX_SHADER)
	if err != nil {
		return nil, err
	}

	fragmentShader, err := compileShader(fragmentShaderSource, gl.FRAGMENT_SHADER)
	if err != nil {
		return nil, err
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

		return nil, fmt.Errorf("failed to link program: %v", log)
	}

	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	shader := &Shader{
		Program: program,
	}

	// bind output color location
	gl.BindFragDataLocation(program, 0, gl.Str("outputColor\x00"))

	// bind vertex coordinates
	vertCoords := uint32(gl.GetAttribLocation(program, gl.Str("vert\x00")))
	gl.EnableVertexAttribArray(vertCoords)
	gl.VertexAttribPointerWithOffset(vertCoords, 2, gl.FLOAT, false, 4*4, 0)

	// bind texture coordinates
	taxCoords := uint32(gl.GetAttribLocation(program, gl.Str("vertTexCoord\x00")))
	gl.EnableVertexAttribArray(taxCoords)
	gl.VertexAttribPointerWithOffset(taxCoords, 2, gl.FLOAT, false, 4*4, 2*4)

	return shader, nil
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

func (s *Shader) Use() *Shader {
	gl.UseProgram(s.Program)
	return s
}

func (s *Shader) Uniform1d(name string, value float64) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform1d(s.Program, location, value)
	return s
}

func (s *Shader) Uniform1dv(name string, values []float64) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform1dv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform1f(name string, value float32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform1f(s.Program, location, value)
	return s
}

func (s *Shader) Uniform1fv(name string, values []float32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform1fv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform1i(name string, value int32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform1i(s.Program, location, value)
	return s
}

func (s *Shader) Uniform1iv(name string, values []int32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform1iv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform2d(name string, v0, v1 float64) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform2d(s.Program, location, v0, v1)
	return s
}

func (s *Shader) Uniform2dv(name string, values []float64) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform2dv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform2f(name string, v0, v1 float32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform2f(s.Program, location, v0, v1)
	return s
}

func (s *Shader) Uniform2fv(name string, values []float32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform2fv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform2i(name string, v0, v1 int32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform2i(s.Program, location, v0, v1)
	return s
}

func (s *Shader) Uniform2iv(name string, values []int32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform2iv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform3d(name string, v0, v1, v2 float64) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform3d(s.Program, location, v0, v1, v2)
	return s
}

func (s *Shader) Uniform3dv(name string, values []float64) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform3dv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform3f(name string, v0, v1, v2 float32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform3f(s.Program, location, v0, v1, v2)
	return s
}

func (s *Shader) Uniform3fv(name string, values []float32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform3fv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform3i(name string, v0, v1, v2 int32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform3i(s.Program, location, v0, v1, v2)
	return s
}

func (s *Shader) Uniform3iv(name string, values []int32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform3iv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform4d(name string, v0, v1, v2, v3 float64) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform4d(s.Program, location, v0, v1, v2, v3)
	return s
}

func (s *Shader) Uniform4dv(name string, values []float64) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform4dv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform4f(name string, v0, v1, v2, v3 float32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform4f(s.Program, location, v0, v1, v2, v3)
	return s
}

func (s *Shader) Uniform4fv(name string, values []float32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform4fv(s.Program, location, int32(len(values)), &values[0])
	return s
}

func (s *Shader) Uniform4i(name string, v0, v1, v2, v3 int32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform4i(s.Program, location, v0, v1, v2, v3)
	return s
}

func (s *Shader) Uniform4iv(name string, values []int32) *Shader {
	attr := fmt.Sprintf("%v\x00", name)
	location := gl.GetUniformLocation(s.Program, gl.Str(attr))
	gl.ProgramUniform4iv(s.Program, location, int32(len(values)), &values[0])
	return s
}
