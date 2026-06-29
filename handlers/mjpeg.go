package handlers

import (
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/textproto"
)

// MJPEG handler.
type MJPEG struct {
	stream *Stream
}

// NewMJPEG returns new MJPEG handler.
func NewMJPEG(stream *Stream) *MJPEG {
	return &MJPEG{stream}
}

// ServeHTTP handles requests on incoming connections.
func (m *MJPEG) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)

		return
	}

	mimeWriter := multipart.NewWriter(w)
	_ = mimeWriter.SetBoundary("--boundary")

	w.Header().Add("Connection", "close")
	w.Header().Add("Cache-Control", "no-store, no-cache")
	w.Header().Add("Content-Type", fmt.Sprintf("multipart/x-mixed-replace;boundary=%s", mimeWriter.Boundary()))

	flusher, _ := w.(http.Flusher)

	ch := m.stream.subscribe()
	defer m.stream.unsubscribe(ch)

	done := r.Context().Done()

	partHeader := make(textproto.MIMEHeader)
	partHeader.Add("Content-Type", "image/jpeg")

	for {
		select {
		case <-done:
			_ = mimeWriter.Close()

			return

		case frame := <-ch:
			partWriter, err := mimeWriter.CreatePart(partHeader)
			if err != nil {
				log.Printf("mjpeg: createPart: %v", err)

				return
			}

			if _, err := partWriter.Write(frame); err != nil {
				log.Printf("mjpeg: write: %v", err)

				return
			}

			if flusher != nil {
				flusher.Flush()
			}
		}
	}
}
