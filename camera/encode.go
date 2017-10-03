package camera

import (
	"image"
	//"image/jpeg"
	"io"

	jpeg "github.com/kjk/golibjpegturbo"
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
	err := jpeg.Encode(e.w, img, &jpeg.Options{Quality: 75})
	if err != nil {
		return err
	}

	return nil
}
