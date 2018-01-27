package handlers

import (
	"log"
	"net/http"

	"github.com/gen2brain/cam2ip/encoder"
	"github.com/gen2brain/cam2ip/reader"
)

// JPEG handler.
type JPEG struct {
	reader reader.ImageReader
}

// NewJPEG returns new JPEG handler.
func NewJPEG(reader reader.ImageReader) *JPEG {
	return &JPEG{reader}
}

// ServeHTTP handles requests on incoming connections.
func (j *JPEG) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Add("Connection", "close")
	w.Header().Add("Cache-Control", "no-store, no-cache")
	w.Header().Add("Content-Type", "image/jpeg")

	img, err := j.reader.Read()
	if err != nil {
		log.Printf("jpeg: read: %v", err)
		return
	}

	err = encoder.New(w).Encode(img)
	if err != nil {
		log.Printf("jpeg: encode: %v", err)
		return
	}
}
