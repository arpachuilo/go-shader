package engine

type Scene struct {
	Root *Transform
}

func NewScene() *Scene {
	return &Scene{
		Root: NewTransform(),
	}
}
