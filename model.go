package engine

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type ModelBufferObject struct {
	vao, vbo, ibo uint32
	*Model
}

func NewModelBufferObject(file string) *ModelBufferObject {
	model := NewModel(file)
	log.Printf(
		"loaded %v with %v vertices and %v vertex indices\n",
		file,
		len(model.Vecs),
		len(model.VecIndices),
	)

	var vao, vbo, ibo uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	// Vertices
	gl.GenBuffers(1, &vbo)
	gl.BindBuffer(gl.ARRAY_BUFFER, vbo)
	vecSize := len(model.Vecs) * F32_SIZE
	normalSize := len(model.Normals) * F32_SIZE
	uvSize := len(model.Uvs) * F32_SIZE
	gl.BufferData(gl.ARRAY_BUFFER, vecSize+uvSize+normalSize, nil, gl.STATIC_DRAW)
	gl.BufferSubData(gl.ARRAY_BUFFER, 0, vecSize, gl.Ptr(model.Vecs))
	gl.BufferSubData(gl.ARRAY_BUFFER, vecSize, normalSize, gl.Ptr(model.Normals))
	gl.BufferSubData(gl.ARRAY_BUFFER, vecSize+normalSize, uvSize, gl.Ptr(model.Uvs))

	// Indices
	gl.GenBuffers(1, &ibo)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, ibo)
	vecISize := len(model.VecIndices) * F32_SIZE
	// normalISize := len(model.NormalIndices) * F32_SIZE
	// uvISize := len(model.UvIndices) * F32_SIZE
	gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, vecISize, gl.Ptr(model.VecIndices), gl.STATIC_DRAW)
	// gl.BufferData(gl.ELEMENT_ARRAY_BUFFER, vecISize+uvISize+normalISize, nil, gl.STATIC_DRAW)
	// gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, 0, vecISize, gl.Ptr(model.VecIndices))
	// gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, vecISize, normalISize, gl.Ptr(model.NormalIndices))
	// gl.BufferSubData(gl.ELEMENT_ARRAY_BUFFER, vecISize+normalISize, uvISize, gl.Ptr(model.UvIndices))

	// bind positions
	gl.EnableVertexAttribArray(0)
	gl.VertexAttribPointerWithOffset(0, 3, gl.FLOAT, false, 0, uintptr(0))

	// bind normals
	gl.EnableVertexAttribArray(2)
	gl.VertexAttribPointerWithOffset(2, 3, gl.FLOAT, false, 0, uintptr(vecSize))

	// bind uvs
	gl.EnableVertexAttribArray(1)
	gl.VertexAttribPointerWithOffset(1, 2, gl.FLOAT, false, 0, uintptr(vecSize+normalSize))

	return &ModelBufferObject{
		vao, vbo, ibo,
		model,
	}
}

func (self ModelBufferObject) Draw() {
	gl.BindVertexArray(self.vao)
	gl.BindBuffer(gl.ELEMENT_ARRAY_BUFFER, self.ibo)
	gl.DrawElements(gl.TRIANGLES, int32(len(self.VecIndices)), gl.UNSIGNED_BYTE, nil)
}

func (self ModelBufferObject) VAO() uint32 {
	return self.vao
}

func (self ModelBufferObject) VBO() uint32 {
	return self.vbo
}

func (self ModelBufferObject) IBO() uint32 {
	return self.ibo
}

// Model is a renderable collection of vecs.
type Model struct {
	// For the v, vt and vn in the obj file.
	// Vecs, Normals []mgl32.Vec3
	// Uvs           []mgl32.Vec2
	Vecs, Normals, Uvs []float32

	// For the fun "f" in the obj file.
	VecIndices, NormalIndices, UvIndices []uint8
}

// NewModel will read an OBJ model file and create a Model from its contents
func NewModel(file string) *Model {
	// Open the file for reading and check for errors.
	objF, err := os.Open(file)
	if err != nil {
		panic(err)
	}

	defer objF.Close()

	objI := bufio.NewReader(objF)

	// Create a model to store stuff.
	model := Model{}

	// Read the file and get it's contents.
	for {
		var lineType string

		// Scan the type field.
		_, err := fmt.Fscanf(objI, "%s", &lineType)

		// Check if it's the end of the file
		// and break out of the loop.
		if err != nil {
			if err == io.EOF {
				break
			}
		}

		// Check the type.
		switch lineType {
		// VERTICES.
		case "v":
			// Create a vec to assign digits to.
			var x, y, z float32

			// Get the digits from the file.
			fmt.Fscanf(objI, "%f %f %f\n", &x, &y, &z)

			// Add the vector to the model.
			model.Vecs = append(model.Vecs, x, y, z)

		// NORMALS.
		case "vn":
			// Create a vec to assign digits to.
			var x, y, z float32

			// Get the digits from the file.
			fmt.Fscanf(objI, "%f %f %f\n", &x, &y, &z)

			// Add the vector to the model.
			model.Normals = append(model.Normals, x, y, z)

		// TEXTURE VERTICES.
		case "vt":
			// Create a Uv pair.
			var x, y float32

			// Get the digits from the file.
			fmt.Fscanf(objI, "%f %f\n", &x, &y)

			// Add the uv to the model.
			model.Uvs = append(model.Uvs, x, y)

		// INDICES.
		case "f":
			var vx, vy, vz int
			var nx, ny, nz int
			var ux, uy, uz int

			// todo: support wider range of formats
			matches, _ := fmt.Fscanf(objI, "%d/%d/%d %d/%d/%d %d/%d/%d\n", &vx, &ux, &nx, &vy, &uy, &ny, &vz, &uz, &nz)
			if matches != 9 {
				panic("Cannot read your file")
			}

			// add indices
			// quick fix w/ minus 1. look into proper handling of this another time
			model.VecIndices = append(model.VecIndices, uint8(vx)-1)
			model.VecIndices = append(model.VecIndices, uint8(vy)-1)
			model.VecIndices = append(model.VecIndices, uint8(vz)-1)

			model.NormalIndices = append(model.NormalIndices, uint8(nx)-1)
			model.NormalIndices = append(model.NormalIndices, uint8(ny)-1)
			model.NormalIndices = append(model.NormalIndices, uint8(nz)-1)

			model.UvIndices = append(model.UvIndices, uint8(ux)-1)
			model.UvIndices = append(model.UvIndices, uint8(uy)-1)
			model.UvIndices = append(model.UvIndices, uint8(uz)-1)
		}
	}

	// Return the newly created Model.
	return &model
}
