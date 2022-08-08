package engine

import (
	"image"

	"github.com/go-gl/gl/v4.1-core/gl"
)

type Framebuffer struct {
	Handle uint32
}

func NewFramebuffer() *Framebuffer {
	var fbo uint32
	gl.GenFramebuffers(1, &fbo)
	return &Framebuffer{fbo}
}

var LastActiveFramebuffer uint32 = 0

func (self Framebuffer) Bind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, self.Handle)
	LastActiveFramebuffer = self.Handle
}

func (self Framebuffer) Unbind() {
	gl.BindFramebuffer(gl.FRAMEBUFFER, LastActiveFramebuffer)
}

type Renderbuffer struct {
	Handle   uint32
	Texture0 *Texture
	Texture1 *Texture

	*Framebuffer
}

func NewRenderbuffer(width, height int) *Renderbuffer {
	var rbo uint32

	fbo := NewFramebuffer()
	fbo.Bind()
	defer fbo.Unbind()

	gl.BindFramebuffer(gl.FRAMEBUFFER, fbo.Handle)

	// color attachment0
	img0 := image.NewRGBA(image.Rect(0, 0, width, height))
	tex0 := LoadTexture(img0)
	tex0.Activate(gl.TEXTURE0)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT0, gl.TEXTURE_2D, tex0.Handle, 0)

	// color attachment1
	img1 := image.NewRGBA(image.Rect(0, 0, width, height))
	tex1 := LoadTexture(img1)
	tex1.Activate(gl.TEXTURE1)
	gl.FramebufferTexture2D(gl.FRAMEBUFFER, gl.COLOR_ATTACHMENT1, gl.TEXTURE_2D, tex1.Handle, 0)

	// create render buffer
	gl.GenRenderbuffers(1, &rbo)
	gl.BindRenderbuffer(gl.RENDERBUFFER, fbo.Handle)
	defer gl.BindRenderbuffer(gl.RENDERBUFFER, LastActiveRenderbuffer)
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, int32(width), int32(height))
	gl.FramebufferRenderbuffer(gl.FRAMEBUFFER, gl.DEPTH_STENCIL_ATTACHMENT, gl.RENDERBUFFER, rbo)
	if gl.CheckFramebufferStatus(gl.FRAMEBUFFER) != gl.FRAMEBUFFER_COMPLETE {
		panic("ERROR: Framebuffer is not complete")
	}

	return &Renderbuffer{
		Handle:      rbo,
		Framebuffer: fbo,
		Texture0:    tex0,
		Texture1:    tex1,
	}
}

var LastActiveRenderbuffer uint32 = 0

func (self Renderbuffer) Bind() {
	gl.BindRenderbuffer(gl.RENDERBUFFER, self.Handle)
	self.Framebuffer.Bind()
	LastActiveRenderbuffer = self.Handle
}

func (self Renderbuffer) Unbind() {
	gl.BindFramebuffer(gl.RENDERBUFFER, LastActiveRenderbuffer)
	self.Framebuffer.Unbind()
}

func (self Renderbuffer) Resize(width, height int) {
	self.Texture0.Resize(width, height)
	self.Texture1.Resize(width, height)
	self.Bind()
	gl.RenderbufferStorage(gl.RENDERBUFFER, gl.DEPTH24_STENCIL8, int32(width), int32(height))
	defer self.Unbind()
}
