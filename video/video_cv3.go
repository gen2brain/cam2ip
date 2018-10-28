// +build cv3,!native

// Package video.
package video

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
	"gocv.io/x/gocv"
)

// Options.
type Options struct {
	Filename string
	Rotate   int
}

// Video represents video.
type Video struct {
	opts  Options
	video *gocv.VideoCapture
	frame *gocv.Mat
}

// New returns new Video for given path.
func New(opts Options) (video *Video, err error) {
	video = &Video{}
	video.opts = opts

	mat := gocv.NewMat()
	video.frame = &mat

	video.video, err = gocv.VideoCaptureFile(opts.Filename)
	if err != nil {
		err = fmt.Errorf("video: can not open video %s: %s", opts.Filename, err.Error())
	}

	return
}

// Read reads next frame from video and returns image.
func (v *Video) Read() (img image.Image, err error) {
	ok := v.video.Read(v.frame)
	if !ok {
		err = fmt.Errorf("video: can not grab frame")
		return
	}

	if v.frame == nil {
		err = fmt.Errorf("video: can not retrieve frame")
		return
	}

	img, e := v.frame.ToImage()
	if e != nil {
		err = fmt.Errorf("video: %v", e)
		return
	}

	if v.opts.Rotate == 0 {
		return
	}

	switch v.opts.Rotate {
	case 90:
		img = imaging.Rotate90(img)
	case 180:
		img = imaging.Rotate180(img)
	case 270:
		img = imaging.Rotate270(img)
	}

	return
}

// Close closes video.
func (v *Video) Close() (err error) {
	if v.video == nil {
		err = fmt.Errorf("video: video is not opened")
		return
	}

	v.frame.Close()
	err = v.video.Close()
	v.video = nil
	return
}
