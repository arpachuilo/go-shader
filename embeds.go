package main

import (
	_ "embed"
)

// basic shaders

//go:embed shaders/vertex.glsl
var vertexShader string

//go:embed shaders/frag.glsl
var fragShader string

// compute shaders
//go:embed shaders/life/cyclic.glsl
var cyclicShader string

//go:embed shaders/life/life.glsl
var golShader string

//go:embed shaders/life/growth_decay.glsl
var gainShader string

// 3 channel mixers
//go:embed shaders/rgba_sampler.glsl
var rgbaShader string

//go:embed shaders/rgb_sampler.glsl
var rgbShader string

// 1 channel gradients
//go:embed shaders/gradients/viridis.glsl
var viridisShader string

//go:embed shaders/gradients/magma.glsl
var magmaShader string

//go:embed shaders/gradients/inferno.glsl
var infernoShader string

//go:embed shaders/gradients/plasma.glsl
var plasmaShader string

//go:embed shaders/gradients/cividis.glsl
var cividisShader string

//go:embed shaders/gradients/turbo.glsl
var turboShader string

//go:embed shaders/gradients/sinebow.glsl
var sinebowShader string
