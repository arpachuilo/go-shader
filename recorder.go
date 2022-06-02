package main

import (
	"bytes"
	"fmt"
	"image"
	"image/jpeg"
	"os"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/go-gl/gl/v4.1-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/icza/mjpeg"
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
	go func(r *Recorder) {

		// create sub-folders
		subFolder := "cap"
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
}
