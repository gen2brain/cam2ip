// Package image.
package image

import (
	"image"
	"io"

	"github.com/gen2brain/jpegn"
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
	return jpegn.Encode(e.w, img, &jpegn.EncodeOptions{Quality: e.quality})
}
