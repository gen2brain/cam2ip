//go:build opencv && !android

// Package camera.
package camera

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"

	im "github.com/gen2brain/cam2ip/image"
)

const (
	propFrameWidth  = 3
	propFrameHeight = 4
)

// Camera represents camera.
type Camera struct {
	opts   Options
	camera *gocv.VideoCapture
	frame  *gocv.Mat
}

// New returns new Camera for given camera index.
func New(opts Options) (camera *Camera, err error) {
	camera = &Camera{}
	camera.opts = opts

	mat := gocv.NewMat()
	camera.frame = &mat

	camera.camera, err = gocv.VideoCaptureDevice(opts.Index)
	if err != nil {
		err = fmt.Errorf("camera: can not open camera %d: %w", opts.Index, err)
	}

	camera.camera.Set(gocv.VideoCaptureProperties(propFrameWidth), opts.Width)
	camera.camera.Set(gocv.VideoCaptureProperties(propFrameHeight), opts.Height)

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	ok := c.camera.Read(c.frame)
	if !ok {
		err = fmt.Errorf("camera: can not grab frame")

		return
	}

	img, err = c.frame.ToImage()
	if err != nil {
		err = fmt.Errorf("camera: %w", err)

		return
	}

	if c.frame == nil {
		err = fmt.Errorf("camera: can not retrieve frame")

		return
	}

	if c.opts.Rotate != 0 {
		img = im.Rotate(img, c.opts.Rotate)
	}

	if c.opts.Flip != "" {
		img = im.Flip(img, c.opts.Flip)
	}

	if c.opts.Timestamp {
		img, err = im.Timestamp(img, "")
	}

	return
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	if c.camera == nil {
		err = fmt.Errorf("camera: camera is not opened")

		return
	}

	err = c.frame.Close()
	if err != nil {
		err = fmt.Errorf("camera: %w", err)

		return
	}

	err = c.camera.Close()
	c.camera = nil

	return
}
