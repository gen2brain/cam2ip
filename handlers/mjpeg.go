// Package handlers.
package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"time"

	"github.com/gen2brain/cam2ip/camera"
	"github.com/gen2brain/cam2ip/encoder"
)

// MJPEG handler.
type MJPEG struct {
	camera *camera.Camera
	delay  int
}

// NewMJPEG returns new MJPEG handler.
func NewMJPEG(camera *camera.Camera, delay int) *MJPEG {
	return &MJPEG{camera, delay}
}

// ServeHTTP handles requests on incoming connections.
func (m *MJPEG) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	mimeWriter := multipart.NewWriter(w)
	mimeWriter.SetBoundary("--boundary")

	w.Header().Add("Connection", "close")
	w.Header().Add("Cache-Control", "no-store, no-cache")
	w.Header().Add("Content-Type", fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary()))

	cn := w.(http.CloseNotifier).CloseNotify()

loop:
	for {
		select {
		case <-cn:
			break loop

		default:
			partHeader := make(textproto.MIMEHeader)
			partHeader.Add("Content-Type", "image/jpeg")

			partWriter, err := mimeWriter.CreatePart(partHeader)
			if err != nil {
				log.Printf("mjpeg: createPart: %v", err)
				continue
			}

			img, err := m.camera.Read()
			if err != nil {
				log.Printf("mjpeg: read: %v", err)
				continue
			}

			err = encoder.New(partWriter).Encode(img)
			if err != nil {
				log.Printf("mjpeg: encode: %v", err)
				continue
			}

			time.Sleep(time.Duration(m.delay) * time.Millisecond)
		}
	}

	mimeWriter.Close()
}
