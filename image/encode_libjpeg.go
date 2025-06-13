//go:build libjpeg

// Package image.
package image

import (
	"image"
	"io"

	"github.com/pixiv/go-libjpeg/jpeg"
)

// NewEncoder returns a new Encoder.
func NewEncoder(w io.Writer) *Encoder {
	return &Encoder{w}
}

// Encoder struct.
type Encoder struct {
	w io.Writer
}

// Encode encodes image to JPEG.
func (e Encoder) Encode(img image.Image) error {
	return jpeg.Encode(e.w, img, &jpeg.EncoderOptions{
		Quality: 75,
	})
}
