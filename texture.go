package engine

import (
	"image"

	"github.com/go-gl/gl/v4.1-core/gl"
	"golang.org/x/image/draw"
)

type Texture struct {
	Handle uint32
	Image  *image.RGBA
}

var LastActiveTexture0 uint32

func LoadTexture(rgba *image.RGBA) *Texture {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
	defer gl.BindTexture(gl.TEXTURE_2D, LastActiveTexture0)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.REPEAT)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.NEAREST)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.NEAREST)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(rgba.Rect.Size().X),
		int32(rgba.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(rgba.Pix))

	// bind back 0 texture
	return &Texture{
		texture,
		rgba,
	}
}

func (self *Texture) Resize(width, height int) *Texture {
	// Create new dst image
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Resize
	draw.NearestNeighbor.Scale(dst, dst.Rect, self.Image, self.Image.Bounds(), draw.Over, nil)

	// Override
	self.Image = dst

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, self.Handle)
	gl.TexImage2D(
		gl.TEXTURE_2D,
		0,
		gl.RGBA,
		int32(dst.Rect.Size().X),
		int32(dst.Rect.Size().Y),
		0,
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(dst.Pix))
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, LastActiveTexture0)

	return self
}

func (self *Texture) Activate(tex uint32) *Texture {
	gl.ActiveTexture(tex)
	gl.BindTexture(gl.TEXTURE_2D, self.Handle)
	if tex == gl.TEXTURE0 {
		LastActiveTexture0 = self.Handle
	}

	return self
}
