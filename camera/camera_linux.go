//go:build !android

// Package camera.
package camera

import (
	"fmt"
	"image"
	"io"
	"slices"
	"unsafe"

	"github.com/korandiz/v4l"
	"golang.org/x/sys/unix"

	im "github.com/gen2brain/cam2ip/image"
)

// supportedFormats lists the formats the camera can decode, in order of preference.
var supportedFormats = []uint32{
	mjpgFourCC, jpegFourCC,
	yuyvFourCC, uyvyFourCC, yvyuFourCC, vyuyFourCC,
	nv12FourCC, yu12FourCC, yv12FourCC,
	rgb24FourCC, bgr24FourCC,
	greyFourCC,
}

// Camera represents camera.
type Camera struct {
	opts    Options
	camera  *v4l.Device
	config  v4l.DeviceConfig
	limited bool
	ycbcr   *image.YCbCr
	rgba    *image.RGBA
	gray    *image.Gray
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

	path := devices[opts.Index].Path

	c.camera, err = v4l.Open(path)
	if err != nil {
		err = fmt.Errorf("camera: %w", err)

		return
	}

	if c.camera == nil {
		err = fmt.Errorf("camera: can not open camera %d", opts.Index)

		return
	}

	defer func() {
		if err != nil {
			c.camera.Close()
			c.camera = nil
		}
	}()

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

	format, ok := selectFormat(formats)
	if !ok {
		err = fmt.Errorf("camera: no supported pixel format")

		return
	}
	c.config.Format = format

	c.config.Width, c.config.Height = nearestSize(configs, format, int(opts.Width), int(opts.Height))

	if fps, ok := maxFrameRate(configs, format, c.config.Width, c.config.Height); ok {
		c.config.FPS = fps
	}

	err = c.camera.SetConfig(c.config)
	if err != nil {
		err = fmt.Errorf("camera: format %d: can not set config: %w", c.config.Format, err)

		return
	}

	c.config, err = c.camera.GetConfig()
	if err != nil {
		err = fmt.Errorf("camera: can not get config: %w", err)

		return
	}

	c.limited = limitedRange(path)

	rect := image.Rect(0, 0, c.config.Width, c.config.Height)

	switch c.config.Format {
	case yuy2FourCC, yuyvFourCC, uyvyFourCC, yvyuFourCC, vyuyFourCC:
		c.ycbcr = image.NewYCbCr(rect, image.YCbCrSubsampleRatio422)
	case nv12FourCC, yu12FourCC, yv12FourCC:
		c.ycbcr = image.NewYCbCr(rect, image.YCbCrSubsampleRatio420)
	case rgb24FourCC, bgr24FourCC:
		c.rgba = image.NewRGBA(rect)
	case greyFourCC:
		c.gray = image.NewGray(rect)
	}

	err = c.camera.TurnOn()
	if err != nil {
		err = fmt.Errorf("camera: format %d: can not turn on: %w", c.config.Format, err)

		return
	}

	return
}

// nearestSize returns the size the device offers for format closest to width x height.
func nearestSize(configs []v4l.DeviceConfig, format uint32, width, height int) (int, int) {
	bw, bh := width, height
	best := -1

	for _, cfg := range configs {
		if cfg.Format != format {
			continue
		}

		d := abs(cfg.Width-width) + abs(cfg.Height-height)
		if best < 0 || d < best {
			bw, bh, best = cfg.Width, cfg.Height, d
		}
	}

	return bw, bh
}

func abs(v int) int {
	if v < 0 {
		return -v
	}

	return v
}

// maxFrameRate returns the highest frame rate the device offers for the given format and resolution.
func maxFrameRate(configs []v4l.DeviceConfig, format uint32, width, height int) (v4l.Frac, bool) {
	var best v4l.Frac
	found := false

	for _, cfg := range configs {
		if cfg.Format != format || cfg.Width != width || cfg.Height != height {
			continue
		}

		if !found || cfg.FPS.Cmp(best) > 0 {
			best = cfg.FPS
			found = true
		}
	}

	return best, found
}

// selectFormat returns the first supported format the device offers.
func selectFormat(formats []uint32) (uint32, bool) {
	for _, f := range supportedFormats {
		if slices.Contains(formats, f) {
			return f, true
		}
	}

	return 0, false
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	buffer, err := c.camera.Capture()
	if err != nil {
		err = fmt.Errorf("camera: format %d: can not grab frame: %w", c.config.Format, err)

		return
	}

	if c.config.Format == mjpgFourCC || c.config.Format == jpegFourCC {
		img, err = im.NewDecoder(buffer).Decode()
		if err != nil {
			err = fmt.Errorf("camera: format %d: can not decode frame: %w", c.config.Format, err)

			return
		}
	} else {
		data, e := io.ReadAll(buffer)
		if e != nil {
			err = fmt.Errorf("camera: format %d: can not read buffer: %w", c.config.Format, e)

			return
		}

		img, err = c.convert(data)
		if err != nil {
			err = fmt.Errorf("camera: format %d: can not retrieve frame: %w", c.config.Format, err)

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
		img = im.Timestamp(img, c.opts.TimeFormat)
	}

	return
}

// convert converts a raw frame in the negotiated format to an image.
func (c *Camera) convert(data []byte) (image.Image, error) {
	if c.ycbcr != nil {
		if err := c.convertYUV(data); err != nil {
			return nil, err
		}

		if c.limited {
			expandRange(c.ycbcr)
		}

		return c.ycbcr, nil
	}

	switch c.config.Format {
	case rgb24FourCC:
		return c.rgba, rgb24ToRgba(data, c.rgba, false)
	case bgr24FourCC:
		return c.rgba, rgb24ToRgba(data, c.rgba, true)
	case greyFourCC:
		return c.gray, greyToGray(data, c.gray)
	}

	return nil, fmt.Errorf("unsupported format %d", c.config.Format)
}

// convertYUV converts a raw YUV frame in the negotiated format into c.ycbcr.
func (c *Camera) convertYUV(data []byte) error {
	if y0, y1, cb, cr, ok := packed422Offsets(c.config.Format); ok {
		return packedYUV422ToYCbCr(data, c.ycbcr, y0, y1, cb, cr)
	}

	switch c.config.Format {
	case yu12FourCC:
		return planar420ToYCbCr(data, c.ycbcr, false)
	case yv12FourCC:
		return planar420ToYCbCr(data, c.ycbcr, true)
	case nv12FourCC:
		return nv12ToYCbCr(data, c.ycbcr)
	}

	return fmt.Errorf("unsupported format %d", c.config.Format)
}

// v4l2Format is struct v4l2_format; the union is pointer aligned, 208 bytes in all on 64-bit and 204 on 32-bit.
type v4l2Format struct {
	typ uint32
	fmt [200 / unsafe.Sizeof(uintptr(0))]uintptr
}

const (
	v4l2BufTypeVideoCapture = 1
	v4l2ColorspaceJPEG      = 7
	v4l2QuantizationFull    = 1
	v4l2QuantizationLimited = 2
)

// limitedRange reports whether the device delivers YUV in limited range, applying V4L2_MAP_QUANTIZATION_DEFAULT.
func limitedRange(path string) bool {
	fd, err := unix.Open(path, unix.O_RDWR|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return true
	}
	defer unix.Close(fd)

	f := v4l2Format{typ: v4l2BufTypeVideoCapture}
	req := uintptr(3<<30 | unsafe.Sizeof(f)<<16 | 'V'<<8 | 4)

	if _, _, e := unix.Syscall(unix.SYS_IOCTL, uintptr(fd), req, uintptr(unsafe.Pointer(&f))); e != 0 {
		return true
	}

	pix := (*[12]uint32)(unsafe.Pointer(&f.fmt))

	switch pix[10] {
	case v4l2QuantizationFull:
		return false
	case v4l2QuantizationLimited:
		return true
	}

	return pix[6] != v4l2ColorspaceJPEG
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

// Info returns the negotiated capture format.
func (c *Camera) Info() Info {
	return Info{Format: fourccName(c.config.Format), Width: c.config.Width, Height: c.config.Height}
}

// Devices returns the available capture devices.
func Devices() ([]DeviceInfo, error) {
	infos := v4l.FindDevices()
	devices := make([]DeviceInfo, 0, len(infos))

	for i, d := range infos {
		name := d.DeviceName
		if name == "" {
			name = d.Path
		}

		devices = append(devices, DeviceInfo{Index: i, Name: name, Path: d.Path})
	}

	return devices, nil
}
