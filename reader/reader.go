// Package reader.
package reader

import (
	"image"
)

// ImageReader interface
type ImageReader interface {
	// Read reads next frame from video and returns image.
	Read() (img image.Image, err error)

	// Close closes camera/video.
	Close() error
}
