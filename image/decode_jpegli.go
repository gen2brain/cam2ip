//go:build jpegli

// Package image.
package image

import (
	"image"
	"io"

	"github.com/gen2brain/jpegli"
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
	return jpegli.DecodeWithOptions(d.r, &jpegli.DecodingOptions{
		DCTMethod:       jpegli.DCTIFast,
		FancyUpsampling: false,
		BlockSmoothing:  false,
		ArithCode:       true,
	})
}
