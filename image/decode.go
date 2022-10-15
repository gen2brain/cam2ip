//go:build !jpeg
// +build !jpeg

// Package image.
package image

import (
	"image"
	"io"

	jpeg "github.com/antonini/golibjpegturbo"
)

// NewDecoder returns a new Decoder.
func NewDecoder(r io.Reader) *Decoder {
	return &Decoder{r}
}

// Decoder struct.
type Decoder struct {
	r io.Reader
}

// Decode decodes image from JPEG.
func (d Decoder) Decode() (image.Image, error) {
	return jpeg.Decode(d.r)
}
