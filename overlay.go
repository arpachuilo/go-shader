package main

import (
	"image"
	"image/color"
	"strings"

	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

var defaultLabelX = 4
var defaultLabelY = 14

var labels = []string{}

var enabled = false

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
