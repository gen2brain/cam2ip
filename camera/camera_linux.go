//go:build !opencv && !android

// Package camera.
package camera

import (
	"fmt"
	"image"

	"github.com/korandiz/v4l"
	"github.com/korandiz/v4l/fmt/mjpeg"

	im "github.com/gen2brain/cam2ip/image"
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
		err = fmt.Errorf("camera: %w", err)

		return
	}

	if camera.camera == nil {
		err = fmt.Errorf("camera: can not open camera %d", opts.Index)

		return
	}

	config, err := camera.camera.GetConfig()
	if err != nil {
		err = fmt.Errorf("camera: %w", err)

		return
	}

	config.Format = mjpeg.FourCC
	config.Width = int(opts.Width)
	config.Height = int(opts.Height)

	err = camera.camera.SetConfig(config)
	if err != nil {
		err = fmt.Errorf("camera: %w", err)

		return
	}

	err = camera.camera.TurnOn()
	if err != nil {
		err = fmt.Errorf("camera: %w", err)

		return
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	buffer, err := c.camera.Capture()
	if err != nil {
		err = fmt.Errorf("camera: can not grab frame: %w", err)

		return
	}

	img, err = im.NewDecoder(buffer).Decode()
	if err != nil {
		err = fmt.Errorf("camera: %w", err)

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
