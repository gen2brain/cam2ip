//go:build !opencv && !android

// Package camera.
package camera

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/anthonynsimon/bild/transform"
	"github.com/korandiz/v4l"
	"github.com/korandiz/v4l/fmt/mjpeg"
	"github.com/pbnjay/pixfont"

	im "github.com/gen2brain/cam2ip/image"
)

// Property identifiers.
const (
	PropBrightness    = v4l.CtrlBrightness
	PropContrast      = v4l.CtrlContrast
	PropSaturation    = v4l.CtrlSaturation
	PropHue           = v4l.CtrlHue
	PropGain          = v4l.CtrlGain
	PropExposure      = v4l.CtrlExposure
	PropWhiteBalanceU = v4l.CtrlWhiteBalance
	PropSharpness     = v4l.CtrlSharpness
	PropWhiteBalanceV = v4l.CtrlDoWhiteBalance
	PropBacklight     = v4l.CtrlBacklightCompensation
)

// Camera represents camera.
type Camera struct {
	opts   Options
	camera *v4l.Device
}

// New returns new Camera for given camera index.
func New(opts Options) (camera *Camera, err error) {
	camera = &Camera{}
	camera.opts = opts

	devices := v4l.FindDevices()
	if len(devices) < opts.Index+1 {
		err = fmt.Errorf("camera: no camera at index %d", opts.Index)
		return
	}

	camera.camera, err = v4l.Open(devices[opts.Index].Path)
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	if camera.camera == nil {
		err = fmt.Errorf("camera: can not open camera %d", opts.Index)
		return
	}

	config, err := camera.camera.GetConfig()
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	config.Format = mjpeg.FourCC
	config.Width = int(opts.Width)
	config.Height = int(opts.Height)

	err = camera.camera.SetConfig(config)
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	err = camera.camera.TurnOn()
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {

	buffer, err := c.camera.Capture()
	if err != nil {
		err = fmt.Errorf("camera: can not grab frame: %s", err.Error())
		return
	}

	img, err = im.NewDecoder(buffer).Decode()
	if err != nil {
		err = fmt.Errorf("camera: %s", err.Error())
		return
	}

	switch c.opts.Rotate {
	case 90:
		img = transform.Rotate(img, 90, &transform.RotationOptions{ResizeBounds: true})
	case 180:
		img = transform.Rotate(img, 180, &transform.RotationOptions{ResizeBounds: true})
	case 270:
		img = transform.Rotate(img, 270, &transform.RotationOptions{ResizeBounds: true})
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
	ret, _ := c.camera.GetControl(uint32(id))
	return float64(ret)
}

// SetProperty sets a camera property.
func (c *Camera) SetProperty(id int, value float64) {
	_ = c.camera.SetControl(uint32(id), int32(value))
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	if c.camera == nil {
		err = fmt.Errorf("camera: camera is not opened")
		return
	}

	c.camera.TurnOff()

	c.camera.Close()
	c.camera = nil
	return
}
