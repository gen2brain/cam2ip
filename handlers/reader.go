package handlers

import (
	"image"
)

// ImageReader interface
type ImageReader interface {
	// Read reads next frame from camera/video and returns image.
	Read() (img image.Image, err error)

	// Close closes camera/video.
	Close() error
}
