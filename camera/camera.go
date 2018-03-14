// +build !cv3

// Package camera.
package camera

import (
	"fmt"
	"image"

	"github.com/lazywei/go-opencv/opencv"
)

// Camera represents camera.
type Camera struct {
	camera *opencv.Capture
	frame  *opencv.IplImage
}

// New returns new Camera for given camera index.
func New(index int) (camera *Camera, err error) {
	camera = &Camera{}

	camera.camera = opencv.NewCameraCapture(index)
	if camera.camera == nil {
		err = fmt.Errorf("camera: can not open camera %d", index)
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	if c.camera.GrabFrame() {
		c.frame = c.camera.RetrieveFrame(1)
		img = c.frame.ToImage()
	} else {
		err = fmt.Errorf("camera: can not grab frame")
	}

	return
}

// GetProperty returns the specified camera property.
func (c *Camera) GetProperty(id int) float64 {
	return c.camera.GetProperty(id)
}

// SetProperty sets a camera property.
func (c *Camera) SetProperty(id int, value float64) {
	c.camera.SetProperty(id, value)
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	if c.camera == nil {
		err = fmt.Errorf("camera: camera is not opened")
		return
	}

	c.frame.Release()
	c.camera.Release()
	c.camera = nil
	return
}
