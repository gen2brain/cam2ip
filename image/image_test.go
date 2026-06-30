package image_test

import (
	"bytes"
	_ "embed"
	stdimage "image"
	"image/color"
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
		err := image.NewEncoder(io.Discard, 75).Encode(img)
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

// labeled returns a 3x2 image whose pixels encode their coordinates as R=x, G=y.
func labeled() *stdimage.RGBA {
	m := stdimage.NewRGBA(stdimage.Rect(0, 0, 3, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 3; x++ {
			m.Set(x, y, color.RGBA{uint8(x), uint8(y), 0, 0xFF})
		}
	}

	return m
}

func at(t *testing.T, img stdimage.Image, x, y int) (int, int) {
	t.Helper()
	r, g, _, _ := img.At(x, y).RGBA()

	return int(r >> 8), int(g >> 8)
}

func TestRotate(t *testing.T) {
	src := labeled()

	r90 := image.Rotate(src, 90)
	if b := r90.Bounds(); b.Dx() != 2 || b.Dy() != 3 {
		t.Fatalf("rotate 90 size = %v, want 2x3", b)
	}
	// Clockwise: source (0,0) ends up top-right of the 2-wide result.
	if x, y := at(t, r90, 1, 0); x != 0 || y != 0 {
		t.Errorf("rotate 90 top-right = (%d,%d), want source (0,0)", x, y)
	}

	r180 := image.Rotate(src, 180)
	if b := r180.Bounds(); b.Dx() != 3 || b.Dy() != 2 {
		t.Fatalf("rotate 180 size = %v, want 3x2", b)
	}
	if x, y := at(t, r180, 0, 0); x != 2 || y != 1 {
		t.Errorf("rotate 180 (0,0) = (%d,%d), want source (2,1)", x, y)
	}

	if got := image.Rotate(src, 45); got != stdimage.Image(src) {
		t.Error("rotate with unsupported angle should return the source unchanged")
	}
}

func TestFlip(t *testing.T) {
	src := labeled()

	h := image.Flip(src, "horizontal")
	if x, y := at(t, h, 0, 0); x != 2 || y != 0 {
		t.Errorf("flipH (0,0) = (%d,%d), want source (2,0)", x, y)
	}

	v := image.Flip(src, "vertical")
	if x, y := at(t, v, 0, 0); x != 0 || y != 1 {
		t.Errorf("flipV (0,0) = (%d,%d), want source (0,1)", x, y)
	}

	if got := image.Flip(src, ""); got != stdimage.Image(src) {
		t.Error("flip with empty direction should return the source unchanged")
	}
}
