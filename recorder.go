package engine

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/gif"
	"image/jpeg"
	"os"
	"time"

	"github.com/ericpauley/go-quantize/quantize"
	"github.com/gen2brain/beeep"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/icza/mjpeg"
	"golang.org/x/image/draw"
)

type Recorder struct {
	On     bool
	Window *glfw.Window

	frames    []*image.RGBA
	startTime time.Time
	endTime   time.Time
}

func NewRecorder(window *glfw.Window) *Recorder {
	return &Recorder{
		On:     false,
		Window: window,
	}
}

func (self *Recorder) Start() {
	self.On = true

	self.frames = make([]*image.RGBA, 0)
	self.startTime = time.Now()

	beeep.Notify("Video Recording Started", "Press Q to end recording", "")
}

func (self *Recorder) Capture() {
	w, h := self.Window.GetFramebufferSize()
	img := *image.NewRGBA(image.Rect(0, 0, w, h))
	gl.ReadPixels(
		0, 0,
		int32(w), int32(h),
		gl.RGBA,
		gl.UNSIGNED_BYTE,
		gl.Ptr(img.Pix),
	)

	self.frames = append(self.frames, &img)
}

func (self *Recorder) End() {
	self.On = false
	self.endTime = time.Now()
	beeep.Notify("Video Recording Finished", "Please wait before closing while your video is encoded", "")

	// create video
	go func(r *Recorder) {

		// create sub-folders
		subFolder := "videos"
		folder := fmt.Sprintf("screencaptures/%v/", subFolder)
		if _, err := os.Stat(folder); os.IsNotExist(err) {
			os.MkdirAll(folder, 0700)
		}

		// create file
		framerate := r.endTime.Sub(r.startTime).Milliseconds() / int64(len(r.frames))
		w, h := r.Window.GetFramebufferSize()
		name := folder + time.Now().Format("20060102150405") + ".avi"
		video, err := mjpeg.New(name, int32(w), int32(h), int32(framerate))
		if err != nil {
			fmt.Println(err)
			return
		}

		// encode frames
		fmt.Println("encoding frames to video")
		for _, frame := range r.frames {
			buf := &bytes.Buffer{}
			err = jpeg.Encode(buf, frame, nil)
			if err != nil {
				fmt.Println(err)
				return
			}

			err = video.AddFrame(buf.Bytes())
			if err != nil {
				fmt.Println(err)
				return
			}
		}

		fmt.Println("video saved")
		beeep.Notify("Video Recording Saved!", name, "")
		err = os.Remove(name + ".idx_")
		if err != nil {
			fmt.Println(err)
			return
		}
	}(self)

	// create gif
	go func(r *Recorder) {
		// create gif
		var delays []int
		var disposal []byte
		var images = make([]*image.Paletted, 0)

		var p = make([]color.Color, 0, 256)
		var q = quantize.MedianCutQuantizer{}

		for _, img := range r.frames {
			p = q.Quantize(p, img)
			p[0] = color.RGBA{0, 0, 0, 0}

			b := img.Bounds()
			frame := image.NewPaletted(b, p)
			draw.Draw(frame, frame.Rect, img, b.Min, draw.Over)
			images = append(images, frame)
			delays = append(delays, 2)
			d := gif.DisposalBackground
			disposal = append(disposal, byte(d))
		}

		out := &gif.GIF{
			Image:           images,
			Delay:           delays,
			Disposal:        disposal,
			BackgroundIndex: byte(0),
		}

		// create sub-folders
		subFolder := "gifs"
		folder := fmt.Sprintf("screencaptures/%v/", subFolder)
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

		// encode gif
		fmt.Println("Saving GIF", name)
		gif.EncodeAll(f, out)

		// cleanup
		beeep.Notify("GIF Saved!", name, "")
		fmt.Println("gif saved")
		if err = f.Close(); err != nil {
			fmt.Println(err)
		}
	}(self)
}
