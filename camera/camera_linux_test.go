//go:build !android

package camera

import (
	"testing"

	"github.com/korandiz/v4l"
	"unsafe"
)

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

func TestV4L2FormatLayout(t *testing.T) {
	want := uintptr(0xc0d05604)
	if unsafe.Sizeof(uintptr(0)) == 4 {
		want = 0xc0cc5604
	}

	f := v4l2Format{}
	if req := uintptr(3<<30 | unsafe.Sizeof(f)<<16 | 'V'<<8 | 4); req != want {
		t.Errorf("VIDIOC_G_FMT = %#x, want %#x", req, want)
	}

	if off := unsafe.Offsetof(f.fmt); off != unsafe.Sizeof(uintptr(0)) {
		t.Errorf("union offset = %d, want %d", off, unsafe.Sizeof(uintptr(0)))
	}
}

func TestNearestSize(t *testing.T) {
	configs := []v4l.DeviceConfig{
		{Format: mjpgFourCC, Width: 640, Height: 480},
		{Format: greyFourCC, Width: 640, Height: 360},
		{Format: yuyvFourCC, Width: 320, Height: 240},
		{Format: yuyvFourCC, Width: 640, Height: 480},
		{Format: yuyvFourCC, Width: 1280, Height: 720},
	}

	tests := []struct {
		format        uint32
		width, height int
		wantW, wantH  int
	}{
		{greyFourCC, 640, 480, 640, 360},
		{yuyvFourCC, 640, 480, 640, 480},
		{yuyvFourCC, 1920, 1080, 1280, 720},
		{yuyvFourCC, 300, 200, 320, 240},
		{nv12FourCC, 800, 600, 800, 600},
	}

	for _, tt := range tests {
		w, h := nearestSize(configs, tt.format, tt.width, tt.height)
		if w != tt.wantW || h != tt.wantH {
			t.Errorf("nearestSize(%s, %dx%d) = %dx%d, want %dx%d", fourccName(tt.format), tt.width, tt.height, w, h, tt.wantW, tt.wantH)
		}
	}
}
