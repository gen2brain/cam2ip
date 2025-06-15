//go:build !opencv && !android

// Package camera.
package camera

import (
	"fmt"
	"image"
	"io"
	"slices"

	"github.com/korandiz/v4l"

	im "github.com/gen2brain/cam2ip/image"
)

// Camera represents camera.
type Camera struct {
	opts   Options
	camera *v4l.Device
	config v4l.DeviceConfig
	ycbcr  *image.YCbCr
}

// New returns new Camera for given camera index.
func New(opts Options) (c *Camera, err error) {
	c = &Camera{}
	c.opts = opts

	devices := v4l.FindDevices()
	if len(devices) < opts.Index+1 {
		err = fmt.Errorf("camera: no camera at index %d", opts.Index)

		return
	}

	c.camera, err = v4l.Open(devices[opts.Index].Path)
	if err != nil {
		err = fmt.Errorf("camera: %w", err)

		return
	}

	if c.camera == nil {
		err = fmt.Errorf("camera: can not open camera %d", opts.Index)

		return
	}

	configs, e := c.camera.ListConfigs()
	if e != nil {
		err = fmt.Errorf("camera: can not list configs: %w", e)

		return
	}

	formats := make([]uint32, 0)
	for _, config := range configs {
		formats = append(formats, config.Format)
	}

	c.config, err = c.camera.GetConfig()
	if err != nil {
		err = fmt.Errorf("camera: can not get config: %w", err)

		return
	}

	if slices.Contains(formats, mjpgFourCC) {
		c.config.Format = mjpgFourCC
	} else if slices.Contains(formats, yuyvFourCC) {
		c.config.Format = yuyvFourCC
	} else {
		err = fmt.Errorf("camera: unsupported format %d", c.config.Format)

		return
	}

	c.config.Width = int(opts.Width)
	c.config.Height = int(opts.Height)

	err = c.camera.SetConfig(c.config)
	if err != nil {
		err = fmt.Errorf("camera: format %d: can not set config: %w", c.config.Format, err)

		return
	}

	if c.config.Format == yuyvFourCC {
		c.ycbcr = image.NewYCbCr(image.Rect(0, 0, int(c.opts.Width), int(c.opts.Height)), image.YCbCrSubsampleRatio422)
	}

	err = c.camera.TurnOn()
	if err != nil {
		err = fmt.Errorf("camera: format %d: can not turn on: %w", c.config.Format, err)

		return
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	buffer, err := c.camera.Capture()
	if err != nil {
		err = fmt.Errorf("camera: format %d: can not grab frame: %w", c.config.Format, err)

		return
	}

	switch c.config.Format {
	case yuy2FourCC, yuyvFourCC:
		data, e := io.ReadAll(buffer)
		if e != nil {
			err = fmt.Errorf("camera: format %d: can not read buffer: %w", c.config.Format, e)

			return
		}

		e = yuy2ToYCbCr422(data, c.ycbcr)
		if e != nil {
			err = fmt.Errorf("camera: format %d: can not retrieve frame: %w", c.config.Format, e)

			return
		}

		img = c.ycbcr
	case mjpgFourCC:
		img, err = im.NewDecoder(buffer).Decode()
		if err != nil {
			err = fmt.Errorf("camera: format %d: can not decode frame: %w", c.config.Format, err)

			return
		}
	}

	if c.opts.Rotate != 0 {
		img = im.Rotate(img, c.opts.Rotate)
	}

	if c.opts.Flip != "" {
		img = im.Flip(img, c.opts.Flip)
	}

	if c.opts.Timestamp {
		img, err = im.Timestamp(img, c.opts.TimeFormat)
	}

	return
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	if c.camera == nil {
		err = fmt.Errorf("camera: close: camera is not opened")

		return
	}

	c.camera.TurnOff()
	c.camera.Close()
	c.camera = nil

	return
}
