package handlers

import (
	"net/http"
	"time"
)

const jpegTimeout = 5 * time.Second

// JPEG handler.
type JPEG struct {
	stream *Stream
}

// NewJPEG returns new JPEG handler.
func NewJPEG(stream *Stream) *JPEG {
	return &JPEG{stream}
}

// ServeHTTP handles requests on incoming connections.
func (j *JPEG) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" && r.Method != "HEAD" {
		http.Error(w, "405 Method Not Allowed", http.StatusMethodNotAllowed)

		return
	}

	ch := j.stream.subscribe()
	defer j.stream.unsubscribe(ch)

	timeout := time.NewTimer(jpegTimeout)
	defer timeout.Stop()

	select {
	case <-r.Context().Done():
		return

	case <-timeout.C:
		http.Error(w, "503 Service Unavailable", http.StatusServiceUnavailable)

		return

	case frame := <-ch:
		w.Header().Add("Connection", "close")
		w.Header().Add("Cache-Control", "no-store, no-cache")
		w.Header().Add("Content-Type", "image/jpeg")

		_, _ = w.Write(frame)
	}
}
