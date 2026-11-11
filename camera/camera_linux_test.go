//go:build !android

package camera

import (
	"testing"
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
