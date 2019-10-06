// +build cv4,!cv2

// Package camera.
package camera

import (
	"fmt"
	"image"

	"github.com/disintegration/imaging"
	"gocv.io/x/gocv"
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
		err = fmt.Errorf("camera: can not open camera %d: %s", opts.Index, err.Error())
	}

	camera.SetProperty(PropFrameWidth, opts.Width)
	camera.SetProperty(PropFrameHeight, opts.Height)

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

	if c.frame == nil {
		err = fmt.Errorf("camera: can not retrieve frame")
		return
	}

	if c.opts.Rotate == 0 {
		return
	}

	switch c.opts.Rotate {
	case 90:
		img = imaging.Rotate90(img)
	case 180:
		img = imaging.Rotate180(img)
	case 270:
		img = imaging.Rotate270(img)
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
