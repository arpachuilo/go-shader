package main

import (
	"unsafe"

	"github.com/go-gl/gl/v4.1-core/gl"
)

var FLOAT32_SIZE int = 4

type Vertices []float32
type Indices []uint8

func (self Vertices) VertexSize() int32 {
	return 4 * int32(FLOAT32_SIZE)
}

func (self Vertices) BufferSize() int {
	return len(self) * FLOAT32_SIZE
}

func (self Vertices) PosOffset() int {
	return 0 * FLOAT32_SIZE
}

func (self Vertices) TexOffset() int {
	return 2 * FLOAT32_SIZE
}

type BufferObject interface {
	Draw()
	VAO() uint32
	VBO() uint32
	IBO() uint32
	GetVertices() Vertices
	GetIndices() Indices
}

type VIBuffer struct {
	Vertices      Vertices
	Indicies      Indices
	Tris          int32
	vao, vbo, ibo uint32
}

func NewVIBuffer(vertices Vertices, indices Indices, tris int32) *VIBuffer {
	buf := &VIBuffer{Vertices: vertices, Indicies: indices, Tris: tris}

	gl.GenVertexArrays(1, &buf.vao)
	gl.BindVertexArray(buf.vao)

	gl.GenBuffers(1, &buf.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, buf.vbo)
	// Vertices
	gl.BufferData(gl.ARRAY_BUFFER, vertices.BufferSize(), gl.Ptr(vertices), gl.STATIC_DRAW)

	gl.GenBuffers(1, &buf.ibo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, buf.ibo)
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, len(indices)*1, gl.Ptr(indices), gl.STATIC_DRAW)

	return buf
}

func (self VIBuffer) Draw() {
	gl.BindVertexArray(self.vao)
	gl.DrawElements(gl.TRIANGLES, self.Tris, gl.UNSIGNED_BYTE, unsafe.Pointer(nil))
	// gl.DrawElements(gl.LINE_STRIP, self.Tris+1, gl.UNSIGNED_BYTE, unsafe.Pointer(nil))
	// gl.PointSize(12.0)
	// gl.DrawElements(gl.POINTS, self.Tris, gl.UNSIGNED_BYTE, unsafe.Pointer(nil))
}

func (self VIBuffer) VAO() uint32 {
	return self.vao
}

func (self VIBuffer) VBO() uint32 {
	return self.vbo
}

func (self VIBuffer) IBO() uint32 {
	return self.ibo
}

func (self VIBuffer) GetVertices() Vertices {
	return self.Vertices
}

func (self VIBuffer) GetIndices() Indices {
	return self.Indicies
}

type VBuffer struct {
	Vertices      Vertices
	Tris          int32
	vao, vbo, ibo uint32
}

func NewVBuffer(vertices Vertices, tris int32) *VBuffer {
	buf := &VBuffer{Vertices: vertices, Tris: tris}

	gl.GenVertexArrays(1, &buf.vao)
	gl.BindVertexArray(buf.vao)

	gl.GenBuffers(1, &buf.vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, buf.vbo)
	gl.BufferData(gl.ARRAY_BUFFER, vertices.BufferSize(), gl.Ptr(vertices), gl.STATIC_DRAW)

	return buf
}

func (self VBuffer) Draw() {
	gl.BindVertexArray(self.vao)
	gl.DrawArrays(gl.TRIANGLE_STRIP, 0, self.Tris)
}

func (self VBuffer) VAO() uint32 {
	return self.vao
}

func (self VBuffer) VBO() uint32 {
	return self.vbo
}

func (self VBuffer) IBO() uint32 {
	return 0
}

func (self VBuffer) GetVertices() Vertices {
	return self.Vertices
}

func (self VBuffer) GetIndices() Indices {
	return nil
}

var TriangleVertices = Vertices{
	-0.8, -0.8, 0.0, 0.0,
	0.0, 0.8, 0.0, 0.5,
	0.8, -0.8, 1.0, 0.0,
}

var QuadVertices = Vertices{
	-1.0, 1.0, 0.0, 1.0,
	1.0, 1.0, 1.0, 1.0,
	-1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 0.0,
}

var C1Vertices = Vertices{
	// Center
	0.0, 0.0, 0.5, 0.5,
	// Top
	-0.2, 0.8, 0.0, 1.0,
	0.2, 0.8, 1.0, 1.0,
	0.0, 0.8, 0.5, 0.8,
	0.0, 1.0, 0.5, 1.0,
	// Bottom
	-0.2, -0.8, 0.0, 0.0,
	0.2, -0.8, 1.0, 0.0,
	0.0, -0.8, 0.5, 0.2,
	0.0, -1.0, 0.5, 0.0,
	// Left
	-0.8, -0.2, 0.0, 0.0,
	-0.8, 0.2, 0.0, 1.0,
	-0.8, 0.0, 0.2, 0.5,
	-1.0, 0.0, 0.0, 0.5,
	// Right
	0.8, -0.2, 1.0, 0.0,
	0.8, 0.2, 1.0, 1.0,
	0.8, 0.0, 0.8, 0.5,
	1.0, 0.0, 1.0, 0.5,
}

var C1Indices = Indices{
	// Top
	0, 1, 3,
	0, 3, 2,
	3, 1, 4,
	3, 4, 2,
	// Bottom
	0, 5, 7,
	0, 7, 6,
	7, 5, 8,
	7, 8, 6,
	// Left
	0, 9, 11,
	0, 11, 10,
	11, 9, 12,
	11, 12, 10,
	// Right
	0, 13, 15,
	0, 15, 14,
	15, 13, 16,
	15, 16, 14,
}

var C1AltIndices = Indices{
	// Outer square border:
	3, 4, 16,
	3, 15, 16,
	15, 16, 8,
	15, 7, 8,
	7, 8, 12,
	7, 11, 12,
	11, 12, 4,
	11, 3, 4,

	// Inner square
	0, 11, 3,
	0, 3, 15,
	0, 15, 7,
	0, 7, 11,
}
