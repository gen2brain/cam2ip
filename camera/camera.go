package camera

import (
	"bytes"
	"fmt"
	"image"
)

// Options .
type Options struct {
	Index      int
	Rotate     int
	Flip       string
	Width      float64
	Height     float64
	Timestamp  bool
	TimeFormat string
}

var (
	yuy2FourCC = fourcc("YUY2")
	yuyvFourCC = fourcc("YUYV")
	mjpgFourCC = fourcc("MJPG")
)

func fourcc(b string) uint32 {
	return uint32(b[0]) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}

func bmp24ToRgba(data []byte, dst *image.RGBA) error {
	r := bytes.NewReader(data)

	width := dst.Bounds().Dx()
	height := dst.Bounds().Dy()

	// There are 3 bytes per pixel, and each row is 4-byte aligned.
	b := make([]byte, (3*width+3)&^3)

	// BMP images are stored bottom-up rather than top-down.
	for y := height - 1; y >= 0; y-- {
		_, err := r.Read(b)
		if err != nil {
			return err
		}

		p := dst.Pix[y*dst.Stride : y*dst.Stride+width*4]
		for i, j := 0, 0; i < len(p); i, j = i+4, j+3 {
			// BMP images are stored in BGR order rather than RGB order.
			p[i+0] = b[j+2]
			p[i+1] = b[j+1]
			p[i+2] = b[j+0]
			p[i+3] = 0xFF
		}
	}

	return nil
}

// yuy2ToYCbCr422 converts a YUY2 (YUYV) byte slice to an image.YCbCr with YCbCrSubsampleRatio422 (I422).
func yuy2ToYCbCr422(data []byte, dst *image.YCbCr) error {
	if dst.SubsampleRatio != image.YCbCrSubsampleRatio422 {
		return fmt.Errorf("subsample ratio must be 422, got %s", dst.SubsampleRatio.String())
	}

	width := dst.Bounds().Dx()
	height := dst.Bounds().Dy()

	if width%2 != 0 {
		return fmt.Errorf("width must be even for YUY2")
	}

	if len(data) != width*height*2 {
		return fmt.Errorf("invalid data length for YUY2")
	}

	stride := width * 2 // 2 bytes per pixel

	for y := 0; y < height; y++ {
		for x := 0; x < width; x += 2 {
			idx := y*stride + x*2

			y0 := data[idx+0]
			cb := data[idx+1]
			y1 := data[idx+2]
			cr := data[idx+3]

			// Y plane: every pixel
			dst.Y[y*dst.YStride+x+0] = y0
			dst.Y[y*dst.YStride+x+1] = y1

			// Cb/Cr plane: every 2 pixels (422)
			off := y*dst.CStride + x/2
			dst.Cb[off] = cb
			dst.Cr[off] = cr
		}
	}

	return nil
}
