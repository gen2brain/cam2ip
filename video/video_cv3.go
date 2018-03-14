// +build cv3

// Package video.
package video

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// Video represents video.
type Video struct {
	video *gocv.VideoCapture
	frame *gocv.Mat
}

// New returns new Video for given path.
func New(filename string) (video *Video, err error) {
	video = &Video{}

	mat := gocv.NewMat()
	video.frame = &mat

	video.video, err = gocv.VideoCaptureFile(filename)
	if err != nil {
		err = fmt.Errorf("video: can not open video %s: %s", filename, err.Error())
	}

	return
}

// Read reads next frame from video and returns image.
func (v *Video) Read() (img image.Image, err error) {
	ok := v.video.Read(*v.frame)
	if !ok {
		err = fmt.Errorf("video: can not grab frame")
		return
	}

	img, e := v.frame.ToImage()
	if e != nil {
		err = fmt.Errorf("video: %v", e)
		return
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
