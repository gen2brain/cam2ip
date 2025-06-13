//go:build libjpeg

// Package image.
package image

import (
	"image"
	"io"

	"github.com/pixiv/go-libjpeg/jpeg"
)

// NewEncoder returns a new Encoder.
func NewEncoder(w io.Writer, quality int) *Encoder {
	return &Encoder{w, quality}
}

// Encoder struct.
type Encoder struct {
	w       io.Writer
	quality int
}

// Encode encodes image to JPEG.
func (e Encoder) Encode(img image.Image) error {
	return jpeg.Encode(e.w, img, &jpeg.EncoderOptions{
		Quality:         e.quality,
		DCTMethod:       jpeg.DCTIFast,
		ProgressiveMode: false,
		OptimizeCoding:  false,
	})
}
