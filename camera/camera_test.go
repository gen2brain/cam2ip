package camera

import (
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/gen2brain/cam2ip/image"
)

func TestCamera(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping camera test in short mode")
	}

	camera, err := New(Options{0, 0, "", 640, 480, false, ""})
	if err != nil {
		t.Skipf("no camera available: %v", err)
	}

	defer func(camera *Camera) {
		err := camera.Close()
		if err != nil {
			t.Error(err)
		}
	}(camera)

	var i int
	var n = 10

	start := time.Now()
	timeout := time.After(time.Duration(n) * time.Second)

	for {
		select {
		case <-timeout:
			fmt.Printf("FPS: %.2f\n", float64(i)/time.Since(start).Seconds())
			return
		default:
			img, err := camera.Read()
			if err != nil {
				t.Error(err)
				continue
			}

			err = image.NewEncoder(io.Discard, 75).Encode(img)
			if err != nil {
				t.Error(err)
				continue
			}

			i += 1
		}
	}
}
