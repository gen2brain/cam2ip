//go:build !libjpeg

package image_test

import (
	"bytes"
	"image/jpeg"
	"testing"

	"github.com/gen2brain/jpegn"

	"github.com/gen2brain/cam2ip/image"
)

// stripDHT removes the Huffman table segments, like webcams that rely on the standard tables (issue #20).
func stripDHT(data []byte) []byte {
	out := append([]byte{}, data[0], data[1])

	for i := 2; i+4 <= len(data); {
		if data[i] != 0xFF {
			break
		}

		marker := data[i+1]
		if marker == 0xDA {
			return append(out, data[i:]...)
		}

		length := int(data[i+2])<<8 | int(data[i+3])
		if marker != 0xC4 {
			out = append(out, data[i:i+2+length]...)
		}

		i += 2 + length
	}

	return out
}

func TestDecodeMissingDHT(t *testing.T) {
	src, err := jpeg.Decode(bytes.NewReader(testJpg))
	if err != nil {
		t.Fatal(err)
	}

	// stdlib uses the standard tables; strip the DHT so the stream relies on the decoder's defaults.
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, src, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	data := stripDHT(buf.Bytes())

	if _, err := jpeg.Decode(bytes.NewReader(data)); err == nil {
		t.Log("note: stdlib decoded a DHT-less stream")
	}

	img, err := image.NewDecoder(bytes.NewReader(data)).Decode()
	if err != nil {
		t.Fatalf("decode without DHT: %v", err)
	}

	if img.Bounds() != src.Bounds() {
		t.Fatalf("bounds = %v, want %v", img.Bounds(), src.Bounds())
	}
}

func TestDecodeLeadingGarbage(t *testing.T) {
	data := append([]byte{0x00, 0x00, 0xFF, 0x12, 0x34}, testJpg...)

	if _, err := jpegn.Decode(bytes.NewReader(data)); err == nil {
		t.Log("note: jpegn decoded leading garbage directly")
	}

	img, err := image.NewDecoder(bytes.NewReader(data)).Decode()
	if err != nil {
		t.Fatalf("decode with leading garbage: %v", err)
	}

	if img.Bounds().Dx() == 0 || img.Bounds().Dy() == 0 {
		t.Fatal("empty image")
	}
}
