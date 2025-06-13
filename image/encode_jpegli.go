//go:build jpegli

// Package image.
package image

import (
	"image"
	"io"

	"github.com/gen2brain/jpegli"
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
	return jpegli.Encode(e.w, img, &jpegli.EncodingOptions{
		Quality:              75,
		ProgressiveLevel:     0,
		ChromaSubsampling:    image.YCbCrSubsampleRatio420,
		DCTMethod:            jpegli.DCTIFast,
		OptimizeCoding:       false,
		AdaptiveQuantization: false,
		StandardQuantTables:  false,
		FancyDownsampling:    false,
	})
}
