package engine

import "github.com/go-gl/mathgl/mgl32"

type Scene struct {
	Root *Node
}

func NewScene() *Scene {
	return &Scene{
		Root: NewNode(),
	}
}

type Node struct {
	Translation mgl32.Mat4
	Scale       mgl32.Mat4
	Rotation    mgl32.Mat4

	Object BufferObject

	Parent   *Node
	Children []*Node
}

func NewNode() *Node {
	return &Node{
		Translation: mgl32.Translate3D(0, 0, 0),
		Scale:       mgl32.Scale3D(1, 1, 1),
		Rotation:    mgl32.QuatIdent().Mat4(),

		Parent:   nil,
		Children: make([]*Node, 0),
	}
}

func (self *Node) ModelMatrix() mgl32.Mat4 {
	return self.Translation.Mul4(self.Rotation).Mul4(self.Scale)
}

func (self *Node) Add(n *Node) *Node {
	n.Parent = self
	self.Children = append(self.Children, n)
	return self
}

func (self *Node) Walk(fn func(n *Node)) {
	if self.Object != nil {
		fn(self)
	}

	for _, n := range self.Children {
		n.Walk(fn)
	}
}
