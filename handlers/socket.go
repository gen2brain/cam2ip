// +build !amd64

package handlers

import (
	"bytes"
	"encoding/base64"
	"log"
	"time"

	"golang.org/x/net/websocket"

	"github.com/gen2brain/cam2ip/image"
	"github.com/gen2brain/cam2ip/reader"
)

// Socket handler.
type Socket struct {
	reader reader.ImageReader
	delay  int
}

// NewSocket returns new socket handler.
func NewSocket(reader reader.ImageReader, delay int) websocket.Handler {
	s := &Socket{reader, delay}
	return websocket.Handler(s.write)
}

// write writes images to socket
func (s *Socket) write(ws *websocket.Conn) {
	for {
		img, err := s.reader.Read()
		if err != nil {
			log.Printf("socket: read: %v", err)
			break
		}

		w := new(bytes.Buffer)

		err = image.NewEncoder(w).Encode(img)
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
