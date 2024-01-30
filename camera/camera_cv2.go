//go:build cv2 && !cv4
// +build cv2,!cv4

// Package camera.
package camera

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/disintegration/imaging"
	"github.com/gen2brain/go-opencv/opencv"
	"github.com/pbnjay/pixfont"
)

// Camera represents camera.
type Camera struct {
	opts   Options
	camera *opencv.Capture
	frame  *opencv.IplImage
}

// New returns new Camera for given camera index.
func New(opts Options) (camera *Camera, err error) {
	camera = &Camera{}
	camera.opts = opts

	camera.camera = opencv.NewCameraCapture(opts.Index)
	if camera.camera == nil {
		err = fmt.Errorf("camera: can not open camera %d", opts.Index)
	}

	camera.SetProperty(PropFrameWidth, opts.Width)
	camera.SetProperty(PropFrameHeight, opts.Height)

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	if !c.camera.GrabFrame() {
		err = fmt.Errorf("camera: can not grab frame")
		return
	}

	c.frame = c.camera.RetrieveFrame(1)
	if c.frame == nil {
		err = fmt.Errorf("camera: can not retrieve frame")
		return
	}

	img = c.frame.ToImage()

	switch c.opts.Rotate {
	case 90:
		img = imaging.Rotate90(img)
	case 180:
		img = imaging.Rotate180(img)
	case 270:
		img = imaging.Rotate270(img)
	}

	if c.opts.Timestamp {
		dimg, ok := img.(draw.Image)
		if !ok {
			err = fmt.Errorf("camera: %T is not a drawable image type", img)
			return
		}

		pixfont.DrawString(dimg, 10, 10, time.Now().Format("2006-01-02 15:04:05"), color.White)
		img = dimg
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
