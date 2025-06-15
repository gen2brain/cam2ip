package handlers

import (
	"log"
	"net/http"

	"github.com/gen2brain/cam2ip/image"
)

// JPEG handler.
type JPEG struct {
	reader  ImageReader
	quality int
}

// NewJPEG returns new JPEG handler.
func NewJPEG(reader ImageReader, quality int) *JPEG {
	return &JPEG{reader, quality}
}

// ServeHTTP handles requests on incoming connections.
func (j *JPEG) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)

		return
	}

	w.Header().Add("Connection", "close")
	w.Header().Add("Cache-Control", "no-store, no-cache")
	w.Header().Add("Content-Type", "image/jpeg")

	img, err := j.reader.Read()
	if err != nil {
		log.Printf("jpeg: read: %v", err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)

		return
	}

	err = image.NewEncoder(w, j.quality).Encode(img)
	if err != nil {
		log.Printf("jpeg: encode: %v", err)
		http.Error(w, "500 Internal Server Error", http.StatusInternalServerError)

		return
	}
}
