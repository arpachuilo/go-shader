package engine

import "github.com/go-gl/mathgl/mgl32"

// maybe refactor with custom getters and setters
type Camera struct {
	Position mgl32.Vec3
	Front    mgl32.Vec3
	Up       mgl32.Vec3

	Width, Height int
	Aspect        float32
	Fov           float32
	Nearclip      float32
	Farclip       float32

	Yaw   float64
	Pitch float64

	// View mgl32.Mat4
	Projection mgl32.Mat4
}

func NewCamera(width, height int) *Camera {
	return &Camera{
		Position: mgl32.Vec3{0, 0, 3},
		Front:    mgl32.Vec3{0, 0, -1},
		Up:       mgl32.Vec3{0, 1, 0},

		Width: width, Height: height,
		Fov:      70,
		Nearclip: 0.1, Farclip: 100,
		Yaw: 0, Pitch: 0,

		Projection: mgl32.Perspective(
			70, float32(width)/float32(height),
			0.01, 100,
		),
	}
}

func (self *Camera) Resize(width, height int) {
	self.Width = width
	self.Height = height
	self.Projection = mgl32.Perspective(
		70, float32(width)/float32(height),
		0.01, 100,
	)
}

func (self *Camera) View() mgl32.Mat4 {
	return mgl32.LookAtV(
		self.Position,
		self.Position.Add(self.Front),
		self.Up,
	)
}
