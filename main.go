package main

import (
	"log"
	"runtime"

	"github.com/go-gl/glfw/v3.3/glfw"
)

// window settings
var windowWidth = 800
var windowHeight = 600

func init() {
	// GLFW event handling must run on the main OS thread
	runtime.LockOSThread()
}

func main() {
	// init glfw
	if err := glfw.Init(); err != nil {
		log.Fatalln("failed to initialize glfw:", err)
	}
	defer glfw.Terminate()

	// setup window
	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, 4)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.True)
	window, err := glfw.CreateWindow(windowWidth, windowHeight, "Visuals", nil, nil)
	if err != nil {
		panic(err)
	}
	window.MakeContextCurrent()
	// TODO: Add program icon
	// window.SetIcon()

	renderer := NewRenderer(window)
	renderer.Setup()
	renderer.Start()
}

var vertexShader = `
#version 410

in vec2 vert;
in vec2 vertTexCoord;

out vec2 fragTexCoord;

void main() {
    gl_Position = vec4(vert,0,1);
    fragTexCoord = vertTexCoord;
}
` + "\x00"

var fragmentShader = `
#version 410
uniform sampler2D tex;

in vec2 fragTexCoord;

out vec4 outputColor;

void main() 
    outputColor = texture(tex, fragTexCoord);
}
` + "\x00"

var golShader = `
#version 410
uniform int s[9];
uniform int b[9];
uniform sampler2D state;
uniform vec2 scale;

in vec2 fragTexCoord;

out vec4 outputColor;

vec4 get(vec2 coord) {
    return texture(state, vec2(gl_FragCoord.xy + coord) / scale, 0);
}

ivec4 alive(vec4 cell) {
	return ivec4(
	  cell.r > 0.0 ? 1 : 0,
	  cell.g > 0.0 ? 1 : 0,
	  cell.b > 0.0 ? 1 : 0,
	  cell.a > 0.0 ? 1 : 0
	);
}

float op(float c, int n) {
    if (n == b[0] || n == b[1] || n == b[2] || n == b[3] || n == b[4] || n == b[5] || n == b[6] || n == b[7] || n == b[8]) {
    	return 1.0;
    } else if (n == s[0] || n == s[1] || n == s[2] || n == s[3] || n == s[4] || n == s[5] || n == s[6] || n == s[7] || n == s[8]) {
      	return c;
    }

    return 0.0;
}

void main() {
    // sum each channel alive
    ivec4 sum = alive(get(vec2(-1, -1))) +
               alive(get(vec2(-1,  0))) +
               alive(get(vec2(-1,  1))) +
               alive(get(vec2( 0, -1))) +
               alive(get(vec2( 0,  1))) +
               alive(get(vec2( 1, -1))) +
               alive(get(vec2( 1,  0))) +
               alive(get(vec2( 1,  1)));

    vec4 current = get(vec2(0, 0));
    outputColor = vec4(
    	op(current.r, sum.r),
    	op(current.g, sum.g),
    	op(current.b, sum.b),
    	op(current.a, sum.a)
    );
}
` + "\x00"

var gainShader = `
#version 410
uniform sampler2D state;
uniform sampler2D self;
uniform vec2 scale;

in vec2 fragTexCoord;
out vec4 outputColor;

const float gain = 0.3;
const float decay = -0.01;

vec4 getCell(vec2 coord) {
    return texture(state, vec2(gl_FragCoord.xy + coord) / scale, 0);
}

vec4 getSelf(vec2 coord) {
    return texture(self, vec2(gl_FragCoord.xy + coord) / scale, 0);
}

float update(float cell, float current) {
	float offset = cell > 0.0 ? gain : decay;
	return current + offset;
}

void main() {
    vec4 cell = getCell(vec2(0, 0));
    vec4 self = getSelf(vec2(0, 0));
    outputColor = vec4(
    	update(cell.r, self.r),
    	update(cell.g, self.g),
    	update(cell.b, self.b),
    	update(cell.a, self.a)
    );
}
` + "\x00"

var copyShader = `
#version 410
uniform float time;
uniform sampler2D state;
uniform vec2 scale;

vec3 viridis(float t) {

    const vec3 c0 = vec3(0.2777273272234177, 0.005407344544966578, 0.3340998053353061);
    const vec3 c1 = vec3(0.1050930431085774, 1.404613529898575, 1.384590162594685);
    const vec3 c2 = vec3(-0.3308618287255563, 0.214847559468213, 0.09509516302823659);
    const vec3 c3 = vec3(-4.634230498983486, -5.799100973351585, -19.33244095627987);
    const vec3 c4 = vec3(6.228269936347081, 14.17993336680509, 56.69055260068105);
    const vec3 c5 = vec3(4.776384997670288, -13.74514537774601, -65.35303263337234);
    const vec3 c6 = vec3(-5.435455855934631, 4.645852612178535, 26.3124352495832);

    return c0+t*(c1+t*(c2+t*(c3+t*(c4+t*(c5+t*c6)))));
}

in vec2 fragTexCoord;

out vec4 outputColor;

void main() {
    vec4 color = texture(state, fragTexCoord.xy);
    // float pct = abs(sin(time));
    // if (color.r > 0.0) {
    // 	color = vec4(viridis(pct), 1.0);
    // }
    // outputColor = color;
    // outputColor = vec4(
    //   1.0 - color.r,
    //   1.0 - color.g,
    //   1.0 - color.b,
    //   1.0
    // );
    outputColor = vec4(color.rgb, 1.0);
}
` + "\x00"

var quadVertices = []float32{
	// positions   // texCoords
	-1.0, 1.0, 0.0, 1.0,
	-1.0, -1.0, 0.0, 0.0,
	1.0, -1.0, 1.0, 0.0,

	-1.0, 1.0, 0.0, 1.0,
	1.0, -1.0, 1.0, 0.0,
	1.0, 1.0, 1.0, 1.0,
}
