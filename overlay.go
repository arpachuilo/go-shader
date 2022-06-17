package main

import (
	"errors"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/go-gl/gl/v4.1-core/gl"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var defaultLabelX = 4
var defaultLabelY = 14

var labels = []string{}

var enabled = false

type Character struct {
	Texture *Texture
	// Glyph
}

type Fontmap struct {
	Characters map[rune]*Character
}

func MustLoadFont() *Fontmap {
	f, err := LoadFont()
	if err != nil {
		panic(err)
	}

	return f
}

func LoadFont() (*Fontmap, error) {
	// would load font here, using basic font for now
	characters := make(map[rune]*Character)

	x := 4
	y := 4
	p := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}
	r := rune('a')
	for r < rune('Z') {
		dr, mask, _, _, ok := basicfont.Face7x13.Glyph(p, r)
		if !ok {
			return nil, errors.New("could not load font Face7x13")
		}

		rgba := image.NewRGBA(dr)
		draw.Draw(rgba, rgba.Bounds(), mask, mask.Bounds().Min, draw.Src)

		var texture uint32
		gl.GenTextures(1, &texture)
		gl.BindTexture(gl.TEXTURE_2D, texture)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, gl.CLAMP_TO_EDGE)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, gl.LINEAR)
		gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, gl.LINEAR)
		gl.TexImage2D(
			gl.TEXTURE_2D,
			0,
			gl.RGBA,
			int32(dr.Size().X),
			int32(dr.Size().Y),
			0,
			gl.RGBA,
			gl.UNSIGNED_BYTE,
			gl.Ptr(rgba.Pix),
		)

		character := Character{
			Texture: &Texture{
				rgba,
				texture,
			},
		}

		characters[r] = &character
		r++
	}

	return &Fontmap{
		Characters: characters,
	}, nil
}

func (self Fontmap) RenderText(s string) {
	// for _, r := range s {
	// 	ch := self.Characters[r]
	// 	bounds, advance, ok := basicfont.Face7x13.GlyphBounds(r)
	// 	if !ok {
	// 		continue
	// 	}
	//
	// 	// xpos := x + texture.
	// }
}

func AddLabel(label string) {
	if !enabled {
		return
	}

	labels = append(labels, strings.Split(label, "\n")...)
}

func PrependLabel(label string) {
	if !enabled {
		return
	}

	labels = append(strings.Split(label, "\n"), labels...)
}

func RenderOverlay(img *image.RGBA) {
	y := defaultLabelY
	for _, label := range labels {
		if label != "" {
			DrawLabel(img, defaultLabelX, y, label)
			y += 14
		}
	}

	// clear out labels
	labels = []string{}
}

func DrawLabel(img *image.RGBA, x, y int, label string) {
	col := color.Black
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}
	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: basicfont.Face7x13,
		Dot:  point,
	}

	length := d.MeasureString(label).Ceil()
	xStart := point.X.Floor()
	yStart := point.Y.Floor()
	for i := xStart - 1; i < xStart+length+1; i++ {
		for j := yStart - 13; j < y+1; j++ {
			img.Set(i, j, color.White)
		}
	}

	d.DrawString(label)
}
