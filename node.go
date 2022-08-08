package engine

import "github.com/go-gl/mathgl/mgl32"

type Scene struct {
	Root *Transform
}

func NewScene() *Scene {
	return &Scene{
		Root: NewTransform(),
	}
}

type Transform struct {
	Translation mgl32.Mat4
	Scale       mgl32.Mat4
	Rotation    mgl32.Mat4

	Object BufferObject

	Parent *Transform
	// Children []*Transform
	Children map[*Transform]struct{}
}

func NewTransform() *Transform {
	return &Transform{
		Translation: mgl32.Translate3D(0, 0, 0),
		Scale:       mgl32.Scale3D(1, 1, 1),
		Rotation:    mgl32.QuatIdent().Mat4(),

		Parent:   nil,
		Children: map[*Transform]struct{}{},
		// Children: make([]*Transform, 0),
	}
}

func (self *Transform) ModelMatrix() mgl32.Mat4 {
	return self.Translation.Mul4(self.Rotation).Mul4(self.Scale)
}

func (self *Transform) Add(n *Transform) *Transform {
	n.Parent = self
	self.Children[n] = struct{}{}
	return self
}

func (self *Transform) Remove() map[*Transform]struct{} {
	delete(self.Parent.Children, self)
	self.Parent = nil
	return self.Children
}

// func (self *Transform) LocalModelMatrix() mgl32.Mat4 {
//
// }

// make local matrix space opt in?
func (self *Transform) walkHelper(space mgl32.Mat4, fn func(space mgl32.Mat4, n *Transform)) {
	lSpace := space.Mul4(self.ModelMatrix())
	if self.Object != nil {
		fn(lSpace, self)
	}

	for t := range self.Children {
		t.walkHelper(lSpace, fn)
	}
}

func (self *Transform) Walk(fn func(space mgl32.Mat4, n *Transform)) {
	lSpace := self.ModelMatrix()
	if self.Object != nil {
		fn(lSpace, self)
	}

	for t := range self.Children {
		t.walkHelper(lSpace, fn)
	}
}
