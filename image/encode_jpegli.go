//go:build jpegli

// Package image.
package image

import (
	"image"
	"io"

	"github.com/gen2brain/jpegli"
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
	return jpegli.Encode(e.w, img, &jpegli.EncodingOptions{
		Quality:              e.quality,
		ProgressiveLevel:     0,
		ChromaSubsampling:    image.YCbCrSubsampleRatio420,
		DCTMethod:            jpegli.DCTIFast,
		OptimizeCoding:       false,
		AdaptiveQuantization: false,
		StandardQuantTables:  false,
		FancyDownsampling:    false,
	})
}
