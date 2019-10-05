// +build cv2,!cv4

// Package video.
package video

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/go-opencv/opencv"
)

// Options.
type Options struct {
	Filename string
	Rotate   int
}

// Video represents video.
type Video struct {
	opts  Options
	video *opencv.Capture
	frame *opencv.IplImage
}

// New returns new Video for given path.
func New(opts Options) (video *Video, err error) {
	video = &Video{}
	video.opts = opts

	video.video = opencv.NewFileCapture(opts.Filename)
	if video.video == nil {
		err = fmt.Errorf("video: can not open video %s", opts.Filename)
	}

	return
}

// Read reads next frame from video and returns image.
func (v *Video) Read() (img image.Image, err error) {
	if v.video.GrabFrame() {
		v.frame = v.video.RetrieveFrame(1)
		if v.frame == nil {
			err = fmt.Errorf("video: can not retrieve frame")
			return
		}

		img = v.frame.ToImage()
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
	} else {
		err = fmt.Errorf("video: can not grab frame")
	}

	return
}

// Close closes video.
func (v *Video) Close() (err error) {
	if v.video == nil {
		err = fmt.Errorf("video: video is not opened")
		return
	}

	v.frame.Release()
	v.video.Release()
	v.video = nil
	return
}
