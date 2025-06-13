//go:build opencv && !android

// Package camera.
package camera

import (
	"fmt"
	"image"

	"gocv.io/x/gocv"

	im "github.com/gen2brain/cam2ip/image"
)

// Property identifiers.
const (
	PropPosMsec = iota
	PropPosFrames
	PropPosAviRatio
	PropFrameWidth
	PropFrameHeight
	PropFps
	PropFourcc
	PropFrameCount
	PropFormat
	PropMode
	PropBrightness
	PropContrast
	PropSaturation
	PropHue
	PropGain
	PropExposure
	PropConvertRgb
	PropWhiteBalanceU
	PropRectification
	PropMonocrome
	PropSharpness
	PropAutoExposure
	PropGamma
	PropTemperature
	PropTrigger
	PropTriggerDelay
	PropWhiteBalanceV
	PropZoom
	PropFocus
	PropGuid
	PropIsoSpeed
	PropMaxDc1394
	PropBacklight
	PropPan
	PropTilt
	PropRoll
	PropIris
	PropSettings
	PropBuffersize
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

	if c.opts.Timestamp {
		img, err = im.Timestamp(img, "")
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

	err = c.frame.Close()
	if err != nil {
		err = fmt.Errorf("camera: %w", err)

		return
	}

	err = c.camera.Close()
	c.camera = nil

	return
}
