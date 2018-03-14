// +build !cv3

// Package video.
package video

import (
	"fmt"
	"image"

	"github.com/lazywei/go-opencv/opencv"
)

// Video represents video.
type Video struct {
	video *opencv.Capture
	frame *opencv.IplImage
}

// New returns new Video for given path.
func New(filename string) (video *Video, err error) {
	video = &Video{}

	video.video = opencv.NewFileCapture(filename)
	if video.video == nil {
		err = fmt.Errorf("video: can not open video %s", filename)
	}

	return
}

// Read reads next frame from video and returns image.
func (v *Video) Read() (img image.Image, err error) {
	if v.video.GrabFrame() {
		v.frame = v.video.RetrieveFrame(1)
		img = v.frame.ToImage()
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
