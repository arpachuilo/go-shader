package main

import (
	"image"

	"github.com/go-gl/gl/v4.1-core/gl"
	"golang.org/x/image/draw"
)

type Texture struct {
	Image  *image.RGBA
	Handle uint32
}

func LoadTexture(rgba *image.RGBA) *Texture {
	var texture uint32
	gl.GenTextures(1, &texture)
	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, texture)
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

	return &Texture{
		rgba,
		texture,
	}
}

var LastActiveTexture0 uint32

func (t *Texture) Resize(width, height int) *Texture {
	// Create new dst image
	dst := image.NewRGBA(image.Rect(0, 0, width, height))

	// Resize
	draw.NearestNeighbor.Scale(dst, dst.Rect, t.Image, t.Image.Bounds(), draw.Over, nil)

	// Override
	t.Image = dst

	gl.ActiveTexture(gl.TEXTURE0)
	gl.BindTexture(gl.TEXTURE_2D, t.Handle)
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

	return t
}

func (t *Texture) Activate(tex uint32) *Texture {
	gl.ActiveTexture(tex)
	gl.BindTexture(gl.TEXTURE_2D, t.Handle)
	if tex == gl.TEXTURE0 {
		LastActiveTexture0 = t.Handle
	}

	return t
}
