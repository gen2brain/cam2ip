package video

import (
	"fmt"
	"image/jpeg"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestVideo(t *testing.T) {
	video, err := New(video.Options{"test.mp4", 0})
	if err != nil {
		t.Fatal(err)
	}

	defer video.Close()

	tmpdir, err := ioutil.TempDir(os.TempDir(), "cam2ip")
	if err != nil {
		t.Error(err)
	}

	defer os.RemoveAll(tmpdir)

	var i int
	var n int = 10

	timeout := time.After(time.Duration(n) * time.Second)

	for {
		select {
		case <-timeout:
			//fmt.Printf("Fps: %d\n", i/n)
			return
		default:
			i += 1

			img, err := video.Read()
			if err != nil {
				t.Error(err)
			}

			file, err := os.Create(filepath.Join(tmpdir, fmt.Sprintf("%03d.jpg", i)))
			if err != nil {
				t.Error(err)
			}

			err = jpeg.Encode(file, img, &jpeg.Options{Quality: 75})
			if err != nil {
				t.Error(err)
			}

			err = file.Close()
			if err != nil {
				t.Error(err)
			}
		}
	}
}
