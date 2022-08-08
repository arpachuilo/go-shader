package main

import (
	"log"
	"math"
	"math/rand"

	"github.com/gen2brain/beeep"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/go-gl/mathgl/mgl32"

	. "gogl"
	. "gogl/assets"
	"gogl/mathutil"
	. "gogl/mathutil"
)

func init() {
	HotProgram = NewLiveEditProgram("./cmd/shader_watch/live_vert.glsl", "./cmd/shader_watch/live_frag.glsl")
}

func HotProgramFn(kill <-chan bool, window *glfw.Window) {
	HotRender(kill, window)
}

type LiveEditProgram struct {
	paused           bool
	frame            int
	deltaTime        float64
	currentFrameTime float64

	watcher *ShaderWatcher

	// current shader
	shader       *Shader
	post         *Shader
	postDisabled bool

	// name of file to watch for updates
	vertFilename string
	fragFilename string

	// buffers
	quad        BufferObject
	bo          BufferObject
	rbo         *Renderbuffer
	lightSource BufferObject
	light       *DirectionalLight

	muls []float64
	mats []*Material

	mx, my float64

	*MouseDelta
	*Renderer
	*Scene
	*Camera
}

func NewLiveEditProgram(vert, frag string) Program {
	return &LiveEditProgram{
		vertFilename: vert,
		fragFilename: frag,

		Scene: NewScene(),
		muls:  make([]float64, 0),
		mats:  make([]*Material, 0),
	}
}

func (self *LiveEditProgram) Load(window *glfw.Window) {}

func (self *LiveEditProgram) LoadR(r *Renderer) {
	// setup window
	self.Renderer = r

	// setup input
	self.MouseDelta = NewMouseDelta(self.Window, 0.1)

	// create renderbuffer for post processing
	self.rbo = NewRenderbuffer(self.Width, self.Height)

	// create watcher
	self.watcher = NewShaderWatcher()

	// create scene shader+vao (all cubes)
	// self.bo = NewVIBuffer(CubeVertices, CubeIndices, 4, 36)
	// self.bo = NewVIBuffer(CubeAltVertices, CubeAltIndices, 36)
	self.bo = NewModelBufferObject("./assets/teapot.obj")
	// self.bo = NewModelBufferObject("./assets/donut01.obj")
	shader := MustCompileShader(VertexShader, FragShader, self.bo)
	self.shader = &shader
	self.watcher.Add(self.shader, self.vertFilename, self.fragFilename, self.bo)

	// create post shader+vao (a quad)
	self.quad = NewV4Buffer(QuadVertices, 2, 4)
	post := MustCompileShader(VertexShader, FragShader, self.quad)
	self.post = &post
	self.watcher.Add(self.post,
		"./cmd/shader_watch/live_post_vert.glsl",
		"./cmd/shader_watch/live_post_frag.glsl",
		self.quad,
	)

	// create light shader+vao (a quad)
	self.lightSource = NewVIBuffer(CubeAltVertices, CubeAltIndices, 36)
	// post := MustCompileShader(VertexShader, FragShader, self.lightSource)
	// self.post = &post
	// self.watcher.Add(self.post,
	// 	"./cmd/shader_watch/live_post_vert.glsl",
	// 	"./cmd/shader_watch/live_post_frag.glsl",
	// 	self.quad,
	// )

	// setup 3d things
	gl.DepthFunc(gl.LESS)
	gl.CullFace(gl.BACK)
	gl.FrontFace(gl.CCW)

	// setup callbacks
	self.Window.SetCursorPosCallback(self.CursorPosCallback)

	// setup camera
	self.Camera = NewCamera(self.Width, self.Height)
	self.Cleaner.Add(self.Camera.Save)

	// setup lights
	self.light = NewSimpleLight()

	// setup scene
	// rand.Seed(time.Now().UnixNano())
	rand.Seed(42)
	var head *Transform
	for i := 0; i < 1000; i++ {
		n := NewTransform()
		n.Object = self.bo
		// scale := rand.Float32() * 20
		scale := float32(1.0)
		n.Scale = mgl32.Scale3D(scale, scale, scale)

		x, y, z := mathutil.RandPointInSphere[float32](5)
		n.Translation = mgl32.Translate3D(x, y, z)

		self.muls = append(self.muls, rand.Float64()*5.0+1.0)
		self.mats = append(self.mats, NewRandomMaterial())
		if head != nil {
			head.Add(n)
		} else {
			self.Scene.Root.Add(n)
		}

		head = n
	}

	gl.ColorMask(true, true, true, true)
	gl.ClearColor(0.0, 0.0, 0.0, 1.0)
}

func (self *LiveEditProgram) ShaderAppliactor(s Shader) Shader {
	return s.
		Uniform1i("u_frame", int32(self.frame)).
		Uniform1f("u_time", float32(self.currentFrameTime)).
		Uniform1f("u_delta", float32(self.deltaTime)).
		Uniform2f("u_mouse", float32(self.mx), float32(self.Height)-float32(self.my)).
		Uniform2f("u_resolution", float32(self.Width), float32(self.Height))
}

func (self *LiveEditProgram) run(t float64) {
	self.mx, self.my = self.Window.GetCursorPos()

	// first pass to fbo
	if !self.postDisabled {
		self.rbo.Bind()
	} else {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
	}

	gl.Enable(gl.DEPTH_TEST)
	gl.Enable(gl.CULL_FACE)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

	// setup part of shader that will be used for all scene objs
	self.shader.Use().
		Apply(self.ShaderAppliactor).
		Apply(self.Camera.ShaderAppliactor).
		Apply(self.light.ShaderAppliactor)

	// todo frustum culling
	// walk scene
	i := 0
	bufs := []uint32{uint32(gl.COLOR_ATTACHMENT0), uint32(gl.COLOR_ATTACHMENT1)}
	gl.DrawBuffers(2, &bufs[0])
	self.Scene.Root.Walk(func(mm mgl32.Mat4, n *Transform) {
		cubeRotation := 360.0 * math.Sin(t/self.muls[i]) / 2.0
		cubeAngle := float32(Deg2Rad(cubeRotation))
		rm := mm.Mul4(
			mgl32.AnglesToQuat(
				cubeAngle, cubeAngle, cubeAngle,
				mgl32.RotationOrder(mgl32.YXZ),
			).Mat4(),
		)

		self.shader.
			Apply(self.mats[i].ShaderAppliactor).
			UniformMatrix4fv("ModelMatrix", &rm)

		self.bo.Draw()
		i++
	})

	// move light
	self.light.Scale = mgl32.Scale3D(10, 10, 10)
	self.light.Translation = mgl32.Translate3D(0, float32(50*math.Sin(t/2)), 0)
	lm := self.light.ModelMatrix()
	mat := NewMaterial()
	self.shader.Use().
		Apply(mat.ShaderAppliactor).
		UniformMatrix4fv("ModelMatrix", &lm)
	self.lightSource.Draw()

	// second pass
	// TODO: handle rbo with multiple textures
	//  handle getting and activating textures
	if !self.postDisabled {
		gl.BindFramebuffer(gl.FRAMEBUFFER, 0)
		gl.Clear(gl.COLOR_BUFFER_BIT)
		self.rbo.Texture0.Activate(gl.TEXTURE0)
		self.rbo.Texture1.Activate(gl.TEXTURE1)
		self.post.Use().
			Uniform1i("color_buffer", 0).
			Uniform1i("normal_buffer", 1).
			Apply(self.ShaderAppliactor)

		gl.Disable(gl.CULL_FACE)
		gl.Disable(gl.DEPTH_TEST)
		self.quad.Draw()
	}
}

func (self *LiveEditProgram) Render(t float64) {
	select {
	case event, ok := <-self.watcher.Events:
		shader, err := self.watcher.Handle(event, ok)
		if err != nil {
			msg := err.Error()
			log.Println(msg)
			beeep.Notify("Shader Compilation Error", "", "")
		} else if shader != nil {
			gl.BindFragDataLocation(*shader.Program, 0, gl.Str("position\x00"))

			if self.paused {
				self.TogglePause()
			}
		}
	default:
		if self.paused {
			return
		}

		currentFrameTime := glfw.GetTime()
		self.deltaTime = currentFrameTime - self.currentFrameTime
		self.currentFrameTime = currentFrameTime

		self.frame = self.frame + 1
		self.ProcessInput()
		self.run(t)
	}
}

func (self *LiveEditProgram) TogglePause() {
	self.paused = !self.paused
	self.PauseBufferSwap = self.paused
	if !self.paused {
		glfw.SetTime(self.currentFrameTime)
	}
}

func (self *LiveEditProgram) CursorPosCallback(w *glfw.Window, x, y float64) {
	dx, dy := self.Delta(x, y)
	if self.Window.GetInputMode(glfw.CursorMode) == glfw.CursorDisabled {
		// pan screen
		self.Camera.Yaw += dx
		self.Camera.Pitch += dy

		if self.Camera.Pitch > 89.0 {
			self.Camera.Pitch = 89.0
		}
		if self.Camera.Pitch < -89.0 {
			self.Camera.Pitch = -89.0
		}

		direction := mgl32.Vec3{
			float32(math.Cos(Deg2Rad(self.Camera.Yaw)) * math.Cos(Deg2Rad(self.Camera.Pitch))),
			float32(math.Sin(Deg2Rad(self.Camera.Pitch))),
			float32(math.Sin(Deg2Rad(self.Camera.Yaw)) * math.Cos(Deg2Rad(self.Camera.Pitch))),
		}
		self.Camera.Front = direction.Normalize()
	}
}

func (self *LiveEditProgram) ProcessInput() {
	mul := 1.0
	if self.KeyPressDetection.Down[glfw.KeyLeftShift] {
		mul = 2.0
	}

	cameraSpeed := float32(32.0 * self.deltaTime * mul)

	if self.KeyPressDetection.Down[glfw.KeyW] {
		self.Camera.Position = self.Camera.Position.Add(self.Camera.Front.Mul(cameraSpeed))
	}

	if self.KeyPressDetection.Down[glfw.KeyS] {
		self.Camera.Position = self.Camera.Position.Sub(self.Camera.Front.Mul(cameraSpeed))
	}

	if self.KeyPressDetection.Down[glfw.KeyA] {
		self.Camera.Position = self.Camera.Position.Sub(
			self.Camera.Front.
				Cross(self.Camera.Up).
				Normalize().
				Mul(cameraSpeed),
		)
	}

	if self.KeyPressDetection.Down[glfw.KeyD] {
		self.Camera.Position = self.Camera.Position.Add(
			self.Camera.Front.
				Cross(self.Camera.Up).
				Normalize().
				Mul(cameraSpeed),
		)
	}

}

func (self *LiveEditProgram) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	self.KeyPressDetection.HandleKeyPress(key, action, mods)
	if key == glfw.KeySpace && action == glfw.Release {
		if mods == glfw.ModControl {
			self.TogglePause()
		} else {
			self.postDisabled = !self.postDisabled
		}
	}

	if key == glfw.KeySpace && mods != glfw.ModShift {
		// self.camera.Position[1] -= 0.1
	}

	if key == glfw.KeySpace && mods == glfw.ModShift {
		// self.camera.Position[1] += 0.1
	}
}

func (self *LiveEditProgram) ResizeCallback(w *glfw.Window, width int, height int) {
	self.Camera.Resize(width, height)
	self.rbo.Resize(width, height)
}
