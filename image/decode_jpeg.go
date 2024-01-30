//go:build !turbo
// +build !turbo

// Package image.
package image

import (
	"image"
	"image/jpeg"
	"io"
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
