package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/png"
	"os"
	"time"

	"github.com/ericpauley/go-quantize/quantize"
	"github.com/go-gl/glfw/v3.3/glfw"
)

type Program interface {
	Render() bool

	ResizeCallback(image *image.RGBA, width int, height int)
	KeyCallback(key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey)
}

type Container struct {
	Program Program
	Image   image.RGBA

	debug bool

	// for capture
	recording   bool
	quantizer   quantize.MedianCutQuantizer
	palette     []color.Color
	framebuffer []*image.Paletted
}

func NewContainer(width, height int, quadVertices *([]float32)) *Container {
	bounds := image.Rect(0, 0, width, height)
	image := image.NewRGBA(bounds)
	c := Container{
		Program: NewNoopProgram(),
		Image:   *image,

		debug: true,
	}

	enabled = c.debug

	return &c
}

var frames = 0
var lastTime time.Time

func (c *Container) Render(w *glfw.Window) {
	didUpdate := c.Program.Render()

	// record before overlay
	if c.recording && didUpdate {
		bounds := c.Image.Bounds()
		c.palette = c.quantizer.Quantize(c.palette, &c.Image)
		cp := image.NewPaletted(bounds, c.palette)
		//cp.Pix = make([]uint8, len(c.Image.Pix))
		//copy(cp.Pix, c.Image.Pix)
		draw.Draw(cp, cp.Rect, &c.Image, bounds.Min, draw.Over)
		c.framebuffer = append(c.framebuffer, cp)
	}

	// perform debug ops
	if c.debug {
		frames++
		currentTime := time.Now()
		delta := currentTime.Sub(lastTime)
		if delta > time.Second {
			fps := frames / int(delta.Seconds())
			w.SetTitle(fmt.Sprintf("FPS: %v", fps))

			lastTime = currentTime
			frames = 0
		}

		PrependLabel(fmt.Sprintf("Resolution: %v", c.Image.Rect.Size()))
		PrependLabel(fmt.Sprintf("Program: %v", c.Program))
		RenderOverlay(&c.Image)
	}
}

func (c *Container) ResizeCallback(w *glfw.Window, width int, height int) {
	// c.Image = *image.NewRGBA(image.Rect(0, 0, width, height))
	// c.Program.ResizeCallback(&c.Image, width, height)
}

func (c *Container) KeyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	defer c.Program.KeyCallback(key, scancode, action, mods)

	// // enable/disable debug
	// if key == glfw.KeyF12 && action == glfw.Release {
	// 	c.debug = !c.debug
	// 	enabled = c.debug

	// 	if !c.debug {
	// 		w.SetTitle(fmt.Sprintf("Program: %v", c.Program))
	// 	}
	// }

	// // close window
	// if key == glfw.KeyEscape && action == glfw.Release {
	// 	w.SetShouldClose(true)
	// }

	// // capture/record/flip
	// if action == glfw.Release {
	// 	switch key {
	// 	case glfw.KeyP:
	// 		c.Capture()
	// 	case glfw.KeyO:
	// 		if !c.recording {
	// 			c.StartRecording()
	// 		} else {
	// 			c.EndRecording()
	// 		}
	// 	}
	// }

	// // switch programs
	// if action == glfw.Release {
	// 	switch key {
	// 	case glfw.KeyF1:
	// 		// c.Program = NewNoopProgram()
	// 	}

	// 	if !c.debug {
	// 		w.SetTitle(fmt.Sprintf("Program: %v", c.Program))
	// 	}
	// }

}

func (c *Container) Capture() {
	// create sub-folders
	folder := fmt.Sprintf("screencaptures/%v/", c.Program)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, 0700)
	}

	// create file
	name := folder + time.Now().Format("20060102150405") + ".png"
	f, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	// encode png
	fmt.Println("Saving", name)
	if err = png.Encode(f, &c.Image); err != nil {
		fmt.Println(err)
	}

	// cleanup
	if err = f.Close(); err != nil {
		fmt.Println(err)
	}
}

func (c *Container) StartRecording() {
	c.quantizer = quantize.MedianCutQuantizer{}
	c.palette = make([]color.Color, 0, 256)
	c.framebuffer = make([]*image.Paletted, 0)
	c.recording = true
}

func (c *Container) EndRecording() {
	// create GIF
	var delays []int
	for range c.framebuffer {
		//p = q.Quantize(p, img)
		//b := img.Bounds()
		//frame := image.NewPaletted(b, p)
		//draw.Draw(frame, frame.Rect, img, b.Min, draw.Over)
		//images = append(images, frame)
		delays = append(delays, 0)
	}
	out := &gif.GIF{
		Image: c.framebuffer,
		Delay: delays,
	}

	// create sub-folders
	folder := fmt.Sprintf("screencaptures/%v/", c.Program)
	if _, err := os.Stat(folder); os.IsNotExist(err) {
		os.MkdirAll(folder, 0700)
	}

	// create file
	name := folder + time.Now().Format("20060102150405") + ".gif"
	f, err := os.Create(name)
	if err != nil {
		fmt.Println(err)
		return
	}

	// encode GIF
	fmt.Println("Saving", name)
	gif.EncodeAll(f, out)

	// cleanup
	if err = f.Close(); err != nil {
		fmt.Println(err)
	}

	c.recording = false
}
