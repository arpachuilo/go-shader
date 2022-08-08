package engine

import (
	"gogl/mathutil"

	"github.com/go-gl/mathgl/mgl32"
)

type Material struct {
	Ambient   mgl32.Vec3
	Diffuse   mgl32.Vec3
	Specular  mgl32.Vec3
	Shininess float32
}

func NewMaterial() *Material {
	return &Material{
		Ambient:   mgl32.Vec3{1, 1, 1},
		Diffuse:   mgl32.Vec3{1, 1, 1},
		Specular:  mgl32.Vec3{1, 1, 1},
		Shininess: 32,
	}
}

func NewRandomMaterial() *Material {
	return &Material{
		Ambient:   mathutil.RandMGL32Vec3(),
		Diffuse:   mathutil.RandMGL32Vec3(),
		Specular:  mathutil.RandMGL32Vec3(),
		Shininess: 32,
	}
}

func (self Material) ShaderAppliactor(s Shader) Shader {
	return s.
		UniformVec3("material.ambient", &self.Ambient).
		UniformVec3("material.diffuse", &self.Diffuse).
		UniformVec3("material.specular", &self.Specular).
		Uniform1f("material.shininess", self.Shininess)
}

// Phong lightning based on https://learnopengl.com/Lighting/Basic-Lighting
type DirectionalLight struct {
	Ambient  mgl32.Vec3
	Diffuse  mgl32.Vec3
	Specular mgl32.Vec3

	*Transform
}

func NewSimpleLight() *DirectionalLight {
	return &DirectionalLight{
		Ambient:   mgl32.Vec3{0.2, 0.2, 0.2},
		Diffuse:   mgl32.Vec3{0.5, 0.5, 0.5},
		Specular:  mgl32.Vec3{1.0, 1.0, 1.0},
		Transform: NewTransform(),
	}
}

func (self DirectionalLight) ShaderAppliactor(s Shader) Shader {
	return s.
		UniformVec3("light.ambient", &self.Ambient).
		UniformVec3("light.diffuse", &self.Diffuse).
		UniformVec3("light.specular", &self.Specular).
		Uniform3f("light.position",
			self.Translation[12],
			self.Translation[13],
			self.Translation[14],
		)
}
