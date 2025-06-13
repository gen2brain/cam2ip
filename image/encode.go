//go:build !libjpeg && !jpegli

// Package image.
package image

import (
	"image"
	"image/jpeg"
	"io"
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
	err := jpeg.Encode(e.w, img, &jpeg.Options{Quality: e.quality})
	if err != nil {
		return err
	}

	return nil
}
