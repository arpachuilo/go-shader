package engine

import (
	"encoding/gob"
	"fmt"
	"log"
	"os"

	"github.com/go-gl/mathgl/mgl32"
)

// maybe refactor with custom getters and setters
type Camera struct {
	Yaw      float64
	Pitch    float64
	Position mgl32.Vec3
	Front    mgl32.Vec3
	Up       mgl32.Vec3
	// view mgl32.Mat4
	projection mgl32.Mat4

	width, height int
	aspect        float32
	fov           float32
	nearclip      float32
	farclip       float32
}

func NewCamera(width, height int) *Camera {
	c := &Camera{
		Yaw: 0, Pitch: 0,
		Position: mgl32.Vec3{0, 0, 3},
		Front:    mgl32.Vec3{0, 0, -1},
		Up:       mgl32.Vec3{0, 1, 0},

		width: width, height: height,
		fov:      70,
		nearclip: 0.1, farclip: 500,
	}

	// attempt to restore camera state
	c.Restore()

	c.SetPerspective()

	return c
}

func (self *Camera) Save() {
	// attempt to save
	td := os.TempDir()
	fp := fmt.Sprintf("%v%v", td, "camera.gob")
	file, err := os.Create(fp)
	if err != nil {
		return
	}

	defer file.Close()

	log.Printf("Camera.Save: saving camera state to %v\n", fp)
	enc := gob.NewEncoder(file)
	err = enc.Encode(self)
	if err != nil {
		log.Println(err)
	}
}

func (self *Camera) Restore() {
	// attempt to restore
	td := os.TempDir()
	fp := fmt.Sprintf("%v%v", td, "camera.gob")
	file, err := os.Open(fp)
	if err != nil {
		return
	}

	defer file.Close()

	log.Printf("Camera.Restore: restoring camera state from %v\n", fp)
	dec := gob.NewDecoder(file)
	err = dec.Decode(self)
	if err != nil {
		log.Println(err)
	}

	os.Remove(fp)
	if err != nil {
		log.Println(err)
	}
}

func (self *Camera) ShaderAppliactor(s Shader) Shader {
	view := self.View()
	return s.
		UniformMatrix4fv("ProjectionMatrix", &self.projection).
		UniformMatrix4fv("ViewMatrix", &view).
		Uniform1f("u_farclip", self.farclip)
}

func (self *Camera) SetPerspective() {
	self.aspect = float32(self.width) / float32(self.height)
	self.projection = mgl32.Perspective(
		self.fov, self.aspect,
		self.nearclip, self.farclip,
	)
}

func (self *Camera) SetOrtho() {
	// self.Projection = mgl32.Perspective(
	// 	70, float32(self.Width)/float32(self.Height),
	// 	0.01, 1000,
	// )
}

func (self *Camera) Resize(width, height int) {
	self.width = width
	self.height = height
	self.SetPerspective()
}

func (self *Camera) View() mgl32.Mat4 {
	return mgl32.LookAtV(
		self.Position,
		self.Position.Add(self.Front),
		self.Up,
	)
}
