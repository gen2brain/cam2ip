package handlers

import (
	"bytes"
	"encoding/base64"
	"log"
	"time"

	"golang.org/x/net/websocket"

	"github.com/gen2brain/cam2ip/camera"
)

// Socket handler.
type Socket struct {
	camera *camera.Camera
	delay  int
}

// NewSocket returns new socket handler.
func NewSocket(camera *camera.Camera, delay int) websocket.Handler {
	s := &Socket{camera, delay}
	return websocket.Handler(s.write)
}

// write writes images to socket
func (s *Socket) write(ws *websocket.Conn) {
	for {
		img, err := s.camera.Read()
		if err != nil {
			log.Printf("socket: read: %v", err)
			continue
		}

		w := new(bytes.Buffer)
		enc := camera.NewEncoder(w)

		err = enc.Encode(img)
		if err != nil {
			log.Printf("socket: encode: %v", err)
			continue
		}

		b64 := base64.StdEncoding.EncodeToString(w.Bytes())

		_, err = ws.Write([]byte(b64))
		if err != nil {
			break
		}

		time.Sleep(time.Duration(s.delay) * time.Millisecond)
	}
}
