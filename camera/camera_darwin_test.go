//go:build darwin

package camera

import (
	"image"
	"testing"
)

func TestBgraToRgba(t *testing.T) {
	dst := image.NewRGBA(image.Rect(0, 0, 2, 1))

	bgraToRgba([]byte{30, 20, 10, 0, 60, 50, 40, 0}, dst)

	assertBytes(t, "Pix", dst.Pix, []byte{10, 20, 30, 0xFF, 40, 50, 60, 0xFF})
}
