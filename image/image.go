package image

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"time"

	"github.com/anthonynsimon/bild/transform"
	"github.com/pbnjay/pixfont"
)

func Rotate(img image.Image, angle int) image.Image {
	switch angle {
	case 90:
		img = transform.Rotate(img, 90, &transform.RotationOptions{ResizeBounds: true})
	case 180:
		img = transform.Rotate(img, 180, &transform.RotationOptions{ResizeBounds: true})
	case 270:
		img = transform.Rotate(img, 270, &transform.RotationOptions{ResizeBounds: true})
	}

	return img
}

func Flip(img image.Image, dir string) image.Image {
	switch dir {
	case "horizontal":
		img = transform.FlipH(img)
	case "vertical":
		img = transform.FlipV(img)
	}

	return img
}

func Timestamp(img image.Image, format string) (image.Image, error) {
	dimg, ok := img.(draw.Image)
	if !ok {
		return img, fmt.Errorf("camera: %T is not a drawable image type", img)
	}

	pixfont.DrawString(dimg, 10, 10, time.Now().Format(format), color.White)

	return dimg, nil
}
