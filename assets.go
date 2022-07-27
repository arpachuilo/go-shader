package main

import (
	_ "embed"
)

// basic shaders

//go:embed assets/shaders/vertex.glsl
var VertexShader string

//go:embed assets/shaders/frag.glsl
var FragShader string

// compute shaders
//go:embed assets/shaders/pong.glsl
var PongShader string

//go:embed assets/shaders/turtle.glsl
var TurtleShader string

//go:embed assets/shaders/fractals/mandelbrot.glsl
var MandelbrotShader string

//go:embed assets/shaders/fractals/julia.glsl
var JuliaShader string

//go:embed assets/shaders/life/smooth_out.glsl
var SmoothOutputShader string

//go:embed assets/shaders/gaussianX.glsl
var GaussXShader string

//go:embed assets/shaders/gaussianY.glsl
var GaussYShader string

//go:embed assets/shaders/life/smooth.glsl
var SmoothShader string

//go:embed assets/shaders/life/cyclic.glsl
var CyclicShader string

//go:embed assets/shaders/life/life.glsl
var GOLShader string

//go:embed assets/shaders/life/growth_decay.glsl
var GainShader string

// 3 channel mixers
//go:embed assets/shaders/rgba_sampler.glsl
var RGBAShader string

//go:embed assets/shaders/rgb_sampler.glsl
var RGBShader string

// 1 channel gradients
//go:embed assets/shaders/gradients/viridis.glsl
var ViridisShader string

//go:embed assets/shaders/gradients/magma.glsl
var MagmaShader string

//go:embed assets/shaders/gradients/inferno.glsl
var InfernoShader string

//go:embed assets/shaders/gradients/plasma.glsl
var PlasmaShader string

//go:embed assets/shaders/gradients/cividis.glsl
var CividisShader string

//go:embed assets/shaders/gradients/turbo.glsl
var TurboShader string

//go:embed assets/shaders/gradients/sinebow.glsl
var SinebowShader string
