package camera

import (
	"image"
	"testing"
)

func TestPackedYUV422ToYCbCr(t *testing.T) {
	// 2x2 image, one macropixel per row.
	wantY := []byte{10, 30, 50, 70}
	wantCb := []byte{20, 60}
	wantCr := []byte{40, 80}

	tests := []struct {
		name           string
		y0, y1, cb, cr int
		data           []byte
	}{
		{"YUYV", 0, 2, 1, 3, []byte{10, 20, 30, 40, 50, 60, 70, 80}},
		{"UYVY", 1, 3, 0, 2, []byte{20, 10, 40, 30, 60, 50, 80, 70}},
		{"YVYU", 0, 2, 3, 1, []byte{10, 40, 30, 20, 50, 80, 70, 60}},
		{"VYUY", 1, 3, 2, 0, []byte{40, 10, 20, 30, 80, 50, 60, 70}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := image.NewYCbCr(image.Rect(0, 0, 2, 2), image.YCbCrSubsampleRatio422)

			if err := packedYUV422ToYCbCr(tt.data, dst, tt.y0, tt.y1, tt.cb, tt.cr); err != nil {
				t.Fatal(err)
			}

			assertBytes(t, "Y", dst.Y, wantY)
			assertBytes(t, "Cb", dst.Cb, wantCb)
			assertBytes(t, "Cr", dst.Cr, wantCr)
		})
	}
}

func TestPackedYUV422Errors(t *testing.T) {
	dst422 := image.NewYCbCr(image.Rect(0, 0, 2, 2), image.YCbCrSubsampleRatio422)
	if err := packedYUV422ToYCbCr(make([]byte, 7), dst422, 0, 2, 1, 3); err == nil {
		t.Error("expected error for wrong data length")
	}

	dst420 := image.NewYCbCr(image.Rect(0, 0, 2, 2), image.YCbCrSubsampleRatio420)
	if err := packedYUV422ToYCbCr(make([]byte, 8), dst420, 0, 2, 1, 3); err == nil {
		t.Error("expected error for wrong subsample ratio")
	}
}

func TestPlanar420ToYCbCr(t *testing.T) {
	wantY := []byte{11, 12, 13, 14}
	wantCb := []byte{21}
	wantCr := []byte{22}

	tests := []struct {
		name   string
		swapUV bool
		data   []byte
	}{
		{"YU12", false, []byte{11, 12, 13, 14, 21, 22}},
		{"YV12", true, []byte{11, 12, 13, 14, 22, 21}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := image.NewYCbCr(image.Rect(0, 0, 2, 2), image.YCbCrSubsampleRatio420)

			if err := planar420ToYCbCr(tt.data, dst, tt.swapUV); err != nil {
				t.Fatal(err)
			}

			assertBytes(t, "Y", dst.Y, wantY)
			assertBytes(t, "Cb", dst.Cb, wantCb)
			assertBytes(t, "Cr", dst.Cr, wantCr)
		})
	}
}

func TestNV12ToYCbCr(t *testing.T) {
	dst := image.NewYCbCr(image.Rect(0, 0, 2, 2), image.YCbCrSubsampleRatio420)

	if err := nv12ToYCbCr([]byte{11, 12, 13, 14, 21, 22}, dst); err != nil {
		t.Fatal(err)
	}

	assertBytes(t, "Y", dst.Y, []byte{11, 12, 13, 14})
	assertBytes(t, "Cb", dst.Cb, []byte{21})
	assertBytes(t, "Cr", dst.Cr, []byte{22})
}

func TestRgb24ToRgba(t *testing.T) {
	tests := []struct {
		name string
		bgr  bool
		data []byte
	}{
		{"RGB", false, []byte{10, 20, 30, 40, 50, 60}},
		{"BGR", true, []byte{30, 20, 10, 60, 50, 40}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dst := image.NewRGBA(image.Rect(0, 0, 2, 1))

			if err := rgb24ToRgba(tt.data, dst, tt.bgr); err != nil {
				t.Fatal(err)
			}

			assertBytes(t, "Pix", dst.Pix, []byte{10, 20, 30, 0xFF, 40, 50, 60, 0xFF})
		})
	}
}

func TestGreyToGray(t *testing.T) {
	dst := image.NewGray(image.Rect(0, 0, 2, 2))

	if err := greyToGray([]byte{1, 2, 3, 4}, dst); err != nil {
		t.Fatal(err)
	}

	assertBytes(t, "Pix", dst.Pix, []byte{1, 2, 3, 4})
}

func TestSelectFormat(t *testing.T) {
	if f, ok := selectFormat([]uint32{yuyvFourCC, mjpgFourCC}); !ok || f != mjpgFourCC {
		t.Errorf("expected MJPG to win over YUYV, got %d ok=%v", f, ok)
	}

	if f, ok := selectFormat([]uint32{greyFourCC, yuyvFourCC}); !ok || f != yuyvFourCC {
		t.Errorf("expected YUYV to win over GREY, got %d ok=%v", f, ok)
	}

	if _, ok := selectFormat([]uint32{fourcc("XVID")}); ok {
		t.Error("expected no supported format")
	}
}

func assertBytes(t *testing.T, name string, got, want []byte) {
	t.Helper()

	if len(got) != len(want) {
		t.Fatalf("%s: length = %d, want %d", name, len(got), len(want))
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("%s[%d] = %d, want %d", name, i, got[i], want[i])
		}
	}
}
