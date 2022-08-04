package main

import (
	"image"
	"log"
	"math"
	"math/rand"
	"time"

	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	. "gogl"
	. "gogl/assets"
	. "gogl/mathutil"
)

func init() {
	HotProgram = NewLiveEditProgram("./cmd/shader_watch/live_vert.glsl", "./cmd/shader_watch/live_frag.glsl")
}

func HotProgramFn(kill <-chan bool, window *glfw.Window) {
	HotRender(kill, window)
}

type LiveEditProgram struct {
	Window        *glfw.Window
	width, height int

	paused bool
	frame  int

	watcher *ShaderWatcher

	// current texture
	texture *Texture

	// current shader
	vert   string
	frag   string
	shader *Shader

	// name of file to watch for updates
	vertFilename string
	fragFilename string

	// buffers
	bo  BufferObject
	fbo uint32

	muls       []float64
	scene      *Scene
	camera     *Camera
	mouseDelta *MouseDelta
}

func NewLiveEditProgram(vert, frag string) Program {
	return &LiveEditProgram{
		vertFilename: vert,
		fragFilename: frag,

		scene: NewScene(),
		muls:  make([]float64, 0),

		mouseDelta: NewMouseDelta(0.1),
	}
}

func (self *LiveEditProgram) Load(window *glfw.Window) {
	// setup window
	self.Window = window
	self.width, self.height = window.GetFramebufferSize()

	// setup buffer object
	self.bo = NewVIBuffer(CubeVertices, CubeIndices, 4, 36)

	// create textures
	img := *image.NewRGBA(image.Rect(0, 0, self.width, self.height))
	self.texture = LoadTexture(&img)

	// create+watch shaders
	shader := MustCompileShader(VertexShader, FragShader, self.bo)
	self.shader = &shader
	self.watcher = NewShaderWatcher()
	self.watcher.Add(self.shader, self.vertFilename, self.fragFilename, self.bo)

	// setup 3d things
	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LESS)

	gl.Enable(gl.CULL_FACE)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)

	// setup callbacks
	self.Window.SetCursorPosCallback(self.CursorPosCallback)

	// setup camera
	self.camera = NewCamera(self.width, self.height)

	// setup scene
	for i := 0; i < 1000; i++ {
		n := NewNode()
		n.Object = self.bo
		scale := rand.Float32() * 5
		n.Scale = mgl32.Scale3D(scale, scale, scale)

		rand.Seed(time.Now().UnixNano())
		x := -50.0 + rand.Float32()*(50.0 - -50.0)
		y := -50.0 + rand.Float32()*(50.0 - -50.0)
		z := -50.0 + rand.Float32()*(50.0 - -50.0)
		n.Translation = mgl32.Translate3D(x, y, z)

		self.scene.Root.Add(n)
		self.muls = append(self.muls, rand.Float64()*5.0+1.0)
	}

	// setup buffer passes
	gl.GenFramebuffers(1, &self.fbo)
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)

	gl.ColorMask(true, true, true, true)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
}

func (self *LiveEditProgram) run(t float64) {
	mx, my := self.Window.GetCursorPos()

	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	gl.Clear(gl.COLOR_BUFFER_BIT)

	// first pass to fbo
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.fbo)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, self.texture.Handle, 0)
	self.texture.Activate(gl.TEXTURE0)

	// walk scene
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	i := 0
	// todo frustum culling
	self.scene.Root.Walk(func(n *Node) {
		cubeRotation := 360.0 * math.Sin(t/self.muls[i]) / 2.0
		cubeAngle := float32(Deg2Rad(cubeRotation))
		n.Rotation = mgl32.AnglesToQuat(
			cubeAngle, cubeAngle, cubeAngle,
			mgl32.RotationOrder(mgl32.YXZ)).Mat4()

		modelMatrix := n.ModelMatrix()
		viewMatrix := self.camera.View()

		self.shader.Use().
			Uniform1i("p_buffer", 0).
			Uniform1i("p_use", 0).
			UniformMatrix4fv("ModelMatrix", &modelMatrix).
			UniformMatrix4fv("ViewMatrix", &viewMatrix).
			UniformMatrix4fv("ProjectionMatrix", &self.camera.Projection).
			Uniform1f("u_farclip", self.camera.Farclip).
			Uniform1i("u_frame", int32(self.frame)).
			Uniform1f("u_time", float32(t)).
			Uniform2f("u_mouse", float32(mx), float32(self.height)-float32(my)).
			Uniform2f("u_resolution", float32(self.width), float32(self.height))

		self.bo.Draw()
		i++
	})

	// second pass
	gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
}

func (self *LiveEditProgram) Render(t float64) {
	select {
	case event, ok := <-self.watcher.Events:
		shader, err := self.watcher.Handle(event, ok)
		if err != nil {
			log.Println(err)
		} else if shader != nil {
			gl.BindFragDataLocation(*shader.Program, 0, gl.Str("position\x00"))
			self.paused = false
		}
	default:
		if self.paused {
			return
		}

		self.frame = self.frame + 1
		self.run(t)
	}
}

func (self *LiveEditProgram) CursorPosCallback(w *glfw.Window, x, y float64) {
	if self.Window.GetInputMode(glfw.CursorMode) == glfw.CursorDisabled {
		// pan screen
		dx, dy := self.mouseDelta.Delta(x, y)
		self.camera.Yaw += dx
		self.camera.Pitch += dy

		if self.camera.Pitch > 89.0 {
			self.camera.Pitch = 89.0
		}
		if self.camera.Pitch < -89.0 {
			self.camera.Pitch = -89.0
		}

		direction := mgl32.Vec3{
			float32(math.Cos(Deg2Rad(self.camera.Yaw)) * math.Cos(Deg2Rad(self.camera.Pitch))),
			float32(math.Sin(Deg2Rad(self.camera.Pitch))),
			float32(math.Sin(Deg2Rad(self.camera.Yaw)) * math.Cos(Deg2Rad(self.camera.Pitch))),
		}
		self.camera.Front = direction.Normalize()
	}
}

func (self *LiveEditProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if key == glfw.KeySpace && action == glfw.Release && mods == glfw.ModControl {
		self.paused = !self.paused
	}

	cameraSpeed := float32(1.5)
	if key == glfw.KeyW {
		self.camera.Position = self.camera.Position.Add(self.camera.Front.Mul(cameraSpeed))
	}

	if key == glfw.KeyS {
		self.camera.Position = self.camera.Position.Sub(self.camera.Front.Mul(cameraSpeed))
	}

	if key == glfw.KeyA {
		self.camera.Position = self.camera.Position.Sub(
			self.camera.Front.
				Cross(self.camera.Up).
				Normalize().
				Mul(cameraSpeed),
		)
	}

	if key == glfw.KeyD {
		self.camera.Position = self.camera.Position.Add(
			self.camera.Front.
				Cross(self.camera.Up).
				Normalize().
				Mul(cameraSpeed),
		)
	}

	if key == glfw.KeySpace && mods != glfw.ModShift {
		// self.camera.Position[1] -= 0.1
	}

	if key == glfw.KeySpace && mods == glfw.ModShift {
		// self.camera.Position[1] += 0.1
	}
}

func (self *LiveEditProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.width, self.height = width, height
	self.texture.Resize(width, height)
	self.camera.Resize(width, height)
}
