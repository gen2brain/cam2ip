//go:build libjpeg

// Package image.
package image

import (
	"image"
	"io"

	"github.com/pixiv/go-libjpeg/jpeg"
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
	return jpeg.Decode(d.r, &jpeg.DecoderOptions{
		DCTMethod:              jpeg.DCTFloat,
		DisableFancyUpsampling: true,
		DisableBlockSmoothing:  true,
	})
}
