//go:build !libjpeg

// Package image.
package image

import (
	"bytes"
	"image"
	"io"

	"github.com/gen2brain/jpegn"
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
	data, err := io.ReadAll(d.r)
	if err != nil {
		return nil, err
	}

	// Some cameras emit MJPEG frames with leading bytes before the SOI marker.
	if i := bytes.Index(data, []byte{0xFF, 0xD8}); i > 0 {
		data = data[i:]
	}

	return jpegn.Decode(bytes.NewReader(data))
}
