// +build cv3

// Package camera.
package camera

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"
)

// Camera represents camera.
type Camera struct {
	camera *gocv.VideoCapture
	frame  *gocv.Mat
}

// New returns new Camera for given camera index.
func New(index int) (camera *Camera, err error) {
	camera = &Camera{}

	mat := gocv.NewMat()
	camera.frame = &mat

	camera.camera, err = gocv.VideoCaptureDevice(index)
	if err != nil {
		err = fmt.Errorf("camera: can not open camera %d: %s", index, err.Error())
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	ok := c.camera.Read(c.frame)
	if !ok {
		err = fmt.Errorf("camera: can not grab frame")
		return
	}

	img, e := c.frame.ToImage()
	if e != nil {
		err = fmt.Errorf("camera: %v", e)
		return
	}

	return
}

// GetProperty returns the specified camera property.
func (c *Camera) GetProperty(id int) float64 {
	return c.camera.Get(gocv.VideoCaptureProperties(id))
}

// SetProperty sets a camera property.
func (c *Camera) SetProperty(id int, value float64) {
	c.camera.Set(gocv.VideoCaptureProperties(id), value)
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	if c.camera == nil {
		err = fmt.Errorf("camera: camera is not opened")
		return
	}

	c.frame.Close()
	err = c.camera.Close()
	c.camera = nil
	return
}
