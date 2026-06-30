//go:build !android

package camera

import "testing"

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
