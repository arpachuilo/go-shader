package assets

import (
	_ "embed"
)

// basic shaders

//go:embed shaders/vertex.glsl
var VertexShader string

//go:embed shaders/frag.glsl
var FragShader string

// compute shaders
//go:embed shaders/gaussianX.glsl
var GaussXShader string

//go:embed shaders/gaussianY.glsl
var GaussYShader string

// 4 channel mixers
//go:embed shaders/rgba_sampler.glsl
var RGBAShader string

// 3 channel mixers
//go:embed shaders/rgb_sampler.glsl
var RGBShader string

// 1 channel gradients
//go:embed shaders/gradients/viridis.glsl
var ViridisShader string

//go:embed shaders/gradients/magma.glsl
var MagmaShader string

//go:embed shaders/gradients/inferno.glsl
var InfernoShader string

//go:embed shaders/gradients/plasma.glsl
var PlasmaShader string

//go:embed shaders/gradients/cividis.glsl
var CividisShader string

//go:embed shaders/gradients/turbo.glsl
var TurboShader string

//go:embed shaders/gradients/sinebow.glsl
var SinebowShader string
