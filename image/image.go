package image

import (
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/pbnjay/pixfont"
)

// Rotate rotates the image clockwise by 90, 180 or 270 degrees.
func Rotate(img image.Image, angle int) image.Image {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()

	var dst *image.RGBA
	switch angle {
	case 90, 270:
		dst = image.NewRGBA(image.Rect(0, 0, h, w))
	case 180:
		dst = image.NewRGBA(image.Rect(0, 0, w, h))
	default:
		return img
	}

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.At(b.Min.X+x, b.Min.Y+y)
			switch angle {
			case 90:
				dst.Set(h-1-y, x, c)
			case 180:
				dst.Set(w-1-x, h-1-y, c)
			case 270:
				dst.Set(y, w-1-x, c)
			}
		}
	}

	return dst
}

// Flip mirrors the image horizontally or vertically.
func Flip(img image.Image, dir string) image.Image {
	if dir != "horizontal" && dir != "vertical" {
		return img
	}

	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	dst := image.NewRGBA(image.Rect(0, 0, w, h))

	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := img.At(b.Min.X+x, b.Min.Y+y)
			if dir == "horizontal" {
				dst.Set(w-1-x, y, c)
			} else {
				dst.Set(x, h-1-y, c)
			}
		}
	}

	return dst
}

func Timestamp(img image.Image, format string) image.Image {
	dimg, ok := img.(draw.Image)
	if !ok {
		b := img.Bounds()
		dimg = image.NewRGBA(image.Rect(0, 0, b.Dx(), b.Dy()))
		draw.Draw(dimg, b, img, b.Min, draw.Src)
	}

	pixfont.DrawString(dimg, 10, 10, time.Now().Format(format), color.White)

	return dimg
}
