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
	yuy2FourCC  = fourcc("YUY2")
	yuyvFourCC  = fourcc("YUYV")
	uyvyFourCC  = fourcc("UYVY")
	yvyuFourCC  = fourcc("YVYU")
	vyuyFourCC  = fourcc("VYUY")
	nv12FourCC  = fourcc("NV12")
	yu12FourCC  = fourcc("YU12")
	yv12FourCC  = fourcc("YV12")
	rgb24FourCC = fourcc("RGB3")
	bgr24FourCC = fourcc("BGR3")
	greyFourCC  = fourcc("GREY")
	mjpgFourCC  = fourcc("MJPG")
	jpegFourCC  = fourcc("JPEG")
)

func fourcc(b string) uint32 {
	return uint32(b[0]) | (uint32(b[1]) << 8) | (uint32(b[2]) << 16) | (uint32(b[3]) << 24)
}

func bmp24ToRgba(data []byte, dst *image.RGBA) error {
	r := bytes.NewReader(data)

	width := dst.Bounds().Dx()
	height := dst.Bounds().Dy()

	// 3 bytes per pixel, rows 4-byte aligned, stored bottom-up in BGR order.
	b := make([]byte, (3*width+3)&^3)

	for y := height - 1; y >= 0; y-- {
		_, err := r.Read(b)
		if err != nil {
			return err
		}

		p := dst.Pix[y*dst.Stride : y*dst.Stride+width*4]
		for i, j := 0, 0; i < len(p); i, j = i+4, j+3 {
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
	return packedYUV422ToYCbCr(data, dst, 0, 2, 1, 3)
}

// packedYUV422ToYCbCr converts packed 4:2:2 to image.YCbCr; y0, y1, cb, cr are the byte offsets within each macropixel.
func packedYUV422ToYCbCr(data []byte, dst *image.YCbCr, y0, y1, cb, cr int) error {
	if dst.SubsampleRatio != image.YCbCrSubsampleRatio422 {
		return fmt.Errorf("subsample ratio must be 422, got %s", dst.SubsampleRatio.String())
	}

	width := dst.Bounds().Dx()
	height := dst.Bounds().Dy()

	if width%2 != 0 {
		return fmt.Errorf("width must be even for 4:2:2")
	}

	if len(data) != width*height*2 {
		return fmt.Errorf("invalid data length for 4:2:2")
	}

	stride := width * 2

	for y := 0; y < height; y++ {
		for x := 0; x < width; x += 2 {
			idx := y*stride + x*2

			dst.Y[y*dst.YStride+x+0] = data[idx+y0]
			dst.Y[y*dst.YStride+x+1] = data[idx+y1]

			off := y*dst.CStride + x/2
			dst.Cb[off] = data[idx+cb]
			dst.Cr[off] = data[idx+cr]
		}
	}

	return nil
}

// planar420ToYCbCr converts planar 4:2:0 (YU12/I420, or YV12 when swapUV is set) to image.YCbCr.
func planar420ToYCbCr(data []byte, dst *image.YCbCr, swapUV bool) error {
	if dst.SubsampleRatio != image.YCbCrSubsampleRatio420 {
		return fmt.Errorf("subsample ratio must be 420, got %s", dst.SubsampleRatio.String())
	}

	width := dst.Bounds().Dx()
	height := dst.Bounds().Dy()

	if width%2 != 0 || height%2 != 0 {
		return fmt.Errorf("width and height must be even for 4:2:0")
	}

	cw := width / 2
	ch := height / 2
	ySize := width * height
	cSize := cw * ch

	if len(data) != ySize+2*cSize {
		return fmt.Errorf("invalid data length for 4:2:0")
	}

	yp := data[:ySize]
	up := data[ySize : ySize+cSize]
	vp := data[ySize+cSize : ySize+2*cSize]

	if swapUV {
		up, vp = vp, up
	}

	for r := 0; r < height; r++ {
		copy(dst.Y[r*dst.YStride:r*dst.YStride+width], yp[r*width:(r+1)*width])
	}

	for r := 0; r < ch; r++ {
		copy(dst.Cb[r*dst.CStride:r*dst.CStride+cw], up[r*cw:(r+1)*cw])
		copy(dst.Cr[r*dst.CStride:r*dst.CStride+cw], vp[r*cw:(r+1)*cw])
	}

	return nil
}

// nv12ToYCbCr converts NV12 (Y plane followed by an interleaved CbCr plane) to image.YCbCr.
func nv12ToYCbCr(data []byte, dst *image.YCbCr) error {
	if dst.SubsampleRatio != image.YCbCrSubsampleRatio420 {
		return fmt.Errorf("subsample ratio must be 420, got %s", dst.SubsampleRatio.String())
	}

	width := dst.Bounds().Dx()
	height := dst.Bounds().Dy()

	if width%2 != 0 || height%2 != 0 {
		return fmt.Errorf("width and height must be even for 4:2:0")
	}

	cw := width / 2
	ch := height / 2
	ySize := width * height

	if len(data) != ySize+2*cw*ch {
		return fmt.Errorf("invalid data length for NV12")
	}

	yp := data[:ySize]
	uv := data[ySize:]

	for r := 0; r < height; r++ {
		copy(dst.Y[r*dst.YStride:r*dst.YStride+width], yp[r*width:(r+1)*width])
	}

	for r := 0; r < ch; r++ {
		for c := 0; c < cw; c++ {
			i := (r*cw + c) * 2
			dst.Cb[r*dst.CStride+c] = uv[i]
			dst.Cr[r*dst.CStride+c] = uv[i+1]
		}
	}

	return nil
}

// rgb24ToRgba converts packed 24-bit RGB (BGR when bgr is set) to image.RGBA.
func rgb24ToRgba(data []byte, dst *image.RGBA, bgr bool) error {
	width := dst.Bounds().Dx()
	height := dst.Bounds().Dy()

	if len(data) != width*height*3 {
		return fmt.Errorf("invalid data length for RGB24")
	}

	for y := 0; y < height; y++ {
		src := data[y*width*3 : (y+1)*width*3]
		row := dst.Pix[y*dst.Stride : y*dst.Stride+width*4]

		for i, j := 0, 0; i < len(row); i, j = i+4, j+3 {
			if bgr {
				row[i+0] = src[j+2]
				row[i+1] = src[j+1]
				row[i+2] = src[j+0]
			} else {
				row[i+0] = src[j+0]
				row[i+1] = src[j+1]
				row[i+2] = src[j+2]
			}
			row[i+3] = 0xFF
		}
	}

	return nil
}

// greyToGray converts an 8-bit greyscale byte slice to an image.Gray.
func greyToGray(data []byte, dst *image.Gray) error {
	width := dst.Bounds().Dx()
	height := dst.Bounds().Dy()

	if len(data) != width*height {
		return fmt.Errorf("invalid data length for GREY")
	}

	for y := 0; y < height; y++ {
		copy(dst.Pix[y*dst.Stride:y*dst.Stride+width], data[y*width:(y+1)*width])
	}

	return nil
}
