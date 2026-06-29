package handlers

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/gen2brain/cam2ip/image"
)

const errorBackoff = 100 * time.Millisecond

// Stream captures frames in a single loop and broadcasts the encoded JPEG to all subscribers.
type Stream struct {
	reader  ImageReader
	delay   int
	quality int

	mu   sync.Mutex
	cond *sync.Cond
	subs map[chan []byte]struct{}
}

// NewStream returns a new Stream and starts its capture loop.
func NewStream(reader ImageReader, delay, quality int) *Stream {
	s := &Stream{
		reader:  reader,
		delay:   delay,
		quality: quality,
		subs:    make(map[chan []byte]struct{}),
	}
	s.cond = sync.NewCond(&s.mu)

	go s.capture()

	return s
}

// capture reads, encodes and broadcasts frames while subscribers are connected.
func (s *Stream) capture() {
	for {
		s.mu.Lock()
		for len(s.subs) == 0 {
			s.cond.Wait()
		}
		s.mu.Unlock()

		img, err := s.reader.Read()
		if err != nil {
			log.Printf("stream: read: %v", err)
			time.Sleep(errorBackoff)

			continue
		}

		buf := new(bytes.Buffer)
		if err := image.NewEncoder(buf, s.quality).Encode(img); err != nil {
			log.Printf("stream: encode: %v", err)
			time.Sleep(errorBackoff)

			continue
		}

		s.broadcast(buf.Bytes())

		if s.delay > 0 {
			time.Sleep(time.Duration(s.delay) * time.Millisecond)
		}
	}
}

// broadcast sends the frame to every subscriber, dropping stale frames so slow clients never block.
func (s *Stream) broadcast(frame []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for ch := range s.subs {
		select {
		case <-ch:
		default:
		}

		select {
		case ch <- frame:
		default:
		}
	}
}

// subscribe registers a new subscriber and returns its frame channel.
func (s *Stream) subscribe() chan []byte {
	ch := make(chan []byte, 1)

	s.mu.Lock()
	s.subs[ch] = struct{}{}
	s.mu.Unlock()

	s.cond.Signal()

	return ch
}

// unsubscribe removes a subscriber.
func (s *Stream) unsubscribe(ch chan []byte) {
	s.mu.Lock()
	delete(s.subs, ch)
	s.mu.Unlock()
}
