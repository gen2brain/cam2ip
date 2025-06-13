package image_test

import (
	"bytes"
	_ "embed"
	"image/jpeg"
	"io"
	"testing"

	"github.com/gen2brain/cam2ip/image"
)

//go:embed testdata/test.jpg
var testJpg []byte

func BenchmarkDecode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, err := image.NewDecoder(bytes.NewReader(testJpg)).Decode()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	img, err := jpeg.Decode(bytes.NewReader(testJpg))
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		err := image.NewEncoder(io.Discard).Encode(img)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkBase64(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_ = image.EncodeToString(testJpg)
	}
}
