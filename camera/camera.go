// Package camera.
package camera

import (
	"fmt"
	"image"

	"github.com/lazywei/go-opencv/opencv"
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
	Index  int
	camera *opencv.Capture
}

// New returns new Camera for given camera index.
func New(index int) (camera *Camera, err error) {
	camera = &Camera{}
	camera.Index = index

	camera.camera = opencv.NewCameraCapture(index)
	if camera.camera == nil {
		err = fmt.Errorf("camera: can not open camera %d", index)
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	if c.camera.GrabFrame() {
		frame := c.camera.RetrieveFrame(1)
		img = frame.ToImage()
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
func (c *Camera) SetProperty(id int, value float64) int {
	return c.camera.SetProperty(id, value)
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	if c.camera == nil {
		err = fmt.Errorf("camera: camera is not opened")
		return
	}

	c.camera.Release()
	c.camera = nil
	return
}
