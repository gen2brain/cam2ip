package main

import (
	"bytes"
	"context"
	_ "embed"
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gen2brain/iv"

	"github.com/gen2brain/cam2ip/camera"
)

const name = "camview"

//go:embed icon.png
var iconPNG []byte

var version string

func init() {
	if version != "" {
		return
	}

	buildInfo, ok := debug.ReadBuildInfo()
	if !ok {
		return
	}

	if buildInfo.Main.Version != "" {
		version = buildInfo.Main.Version
	}

	for _, kv := range buildInfo.Settings {
		if kv.Key == "vcs.revision" && kv.Value != "" {
			version = kv.Value
			if len(version) > 7 {
				version = version[:7]
			}
		}
	}
}

func main() {
	runtime.LockOSThread()

	var (
		index       int
		device      string
		listDevices bool
		width       float64
		height      float64
		rotate      int
		flip        string
		timestamp   bool
		timeFormat  string
		fps         int
		fullscreen  bool
		showVersion bool
	)

	flag.IntVar(&index, "index", 0, "Camera index")
	flag.StringVar(&device, "device", "", "Camera name to use, matched as substring, overrides index")
	flag.BoolVar(&listDevices, "list-devices", false, "List available cameras and exit")
	flag.Float64Var(&width, "width", 640, "Frame width")
	flag.Float64Var(&height, "height", 480, "Frame height")
	flag.IntVar(&rotate, "rotate", 0, "Rotate image, valid values are 90, 180, 270")
	flag.StringVar(&flip, "flip", "", "Flip image, valid values are horizontal and vertical")
	flag.BoolVar(&timestamp, "timestamp", false, "Draws timestamp on image")
	flag.StringVar(&timeFormat, "time-format", "2006-01-02 15:04:05", "Time format")
	flag.IntVar(&fps, "fps", 30, "Maximum frames per second to display")
	flag.BoolVar(&fullscreen, "fullscreen", false, "Start in fullscreen")
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")
	flag.Parse()

	if showVersion {
		fmt.Printf("%s %s\n", name, version)
		os.Exit(0)
	}

	if listDevices {
		devices, err := camera.Devices()
		if err != nil {
			stderr("%s\n", err.Error())
			os.Exit(1)
		}

		for _, d := range devices {
			if d.Path != "" {
				fmt.Printf("%d: %s (%s)\n", d.Index, d.Name, d.Path)
			} else {
				fmt.Printf("%d: %s\n", d.Index, d.Name)
			}
		}

		os.Exit(0)
	}

	if fps < 1 {
		fps = 1
	}

	if device != "" {
		idx, err := deviceIndex(device)
		if err != nil {
			stderr("%s\n", err.Error())
			os.Exit(1)
		}

		index = idx
	}

	cam, err := camera.New(camera.Options{
		Index:      index,
		Rotate:     rotate,
		Flip:       flip,
		Width:      width,
		Height:     height,
		Timestamp:  timestamp,
		TimeFormat: timeFormat,
	})
	if err != nil {
		stderr("%s\n", err.Error())
		os.Exit(1)
	}

	banner(index, cam.Info())

	view, err := iv.New(iv.Options{AppID: name, Width: int(width), Height: int(height)})
	if err != nil {
		cam.Close()
		stderr("%s\n", err.Error())
		os.Exit(1)
	}

	_ = view.SetTitle(title(index))

	if icon := appIcon(); icon != nil {
		_ = view.SetIcon(icon)
	}

	if fullscreen {
		_ = view.ToggleFullscreen()
	}

	s := &stream{}
	done := make(chan struct{})

	var closeOnce sync.Once
	quit := func() {
		closeOnce.Do(func() { close(done) })
		s.wake()
	}

	view.SetClosedHandler(quit)
	view.SetKeyPressHandler(func(key int) {
		switch key {
		case iv.KeyEscape, iv.KeyQ:
			quit()
		case iv.KeyF, iv.KeyF11:
			_ = view.ToggleFullscreen()
		}
	})

	captureDone := make(chan struct{})
	go func() {
		defer close(captureDone)
		defer cam.Close()

		capture(cam, s, fps, isTerminal(os.Stderr), done)
	}()

	blank := image.NewRGBA(image.Rect(0, 0, 1, 1))

	for {
		select {
		case <-done:
		default:
			img, ctx := s.frame()
			if img == nil {
				img = blank
			}

			if err := view.Display(ctx, img, "", true); err != nil {
				stderr("%s\n", err.Error())
				quit()
			}

			continue
		}

		break
	}

	// Capture owns and closes the camera; bound the wait so a stall cannot hang exit.
	select {
	case <-captureDone:
	case <-time.After(time.Second):
	}

	view.Close()
}

// stream holds the latest frame and wakes the display loop when a new one arrives.
type stream struct {
	mu     sync.Mutex
	img    image.Image
	shown  image.Image
	free   []image.Image
	cancel context.CancelFunc
}

// buffer returns a recycled frame buffer, or nil.
func (s *stream) buffer() image.Image {
	s.mu.Lock()
	defer s.mu.Unlock()

	if n := len(s.free); n > 0 {
		b := s.free[n-1]
		s.free = s.free[:n-1]

		return b
	}

	return nil
}

// set stores a new frame and wakes the in-flight Display.
func (s *stream) set(img image.Image) {
	s.mu.Lock()
	if s.img != nil {
		s.free = append(s.free, s.img)
	}
	s.img = img
	s.mu.Unlock()

	s.wake()
}

// wake cancels the in-flight Display so the display loop advances.
func (s *stream) wake() {
	s.mu.Lock()
	cancel := s.cancel
	s.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

// frame returns the latest image and a context that stays live until the next
// frame arrives or the viewer is asked to quit.
func (s *stream) frame() (image.Image, context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.img != nil {
		if s.shown != nil {
			s.free = append(s.free, s.shown)
		}
		s.shown, s.img = s.img, nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel

	return s.shown, ctx
}

// copyFrame copies src into dst, reusing dst when it matches.
func copyFrame(dst, src image.Image) image.Image {
	switch s := src.(type) {
	case *image.YCbCr:
		d, ok := dst.(*image.YCbCr)
		if !ok || len(d.Y) != len(s.Y) || len(d.Cb) != len(s.Cb) {
			d = &image.YCbCr{Y: make([]byte, len(s.Y)), Cb: make([]byte, len(s.Cb)), Cr: make([]byte, len(s.Cr))}
		}
		copy(d.Y, s.Y)
		copy(d.Cb, s.Cb)
		copy(d.Cr, s.Cr)
		d.YStride, d.CStride, d.SubsampleRatio, d.Rect = s.YStride, s.CStride, s.SubsampleRatio, s.Rect

		return d
	case *image.RGBA:
		d, ok := dst.(*image.RGBA)
		if !ok || len(d.Pix) != len(s.Pix) {
			d = &image.RGBA{Pix: make([]byte, len(s.Pix))}
		}
		copy(d.Pix, s.Pix)
		d.Stride, d.Rect = s.Stride, s.Rect

		return d
	case *image.Gray:
		d, ok := dst.(*image.Gray)
		if !ok || len(d.Pix) != len(s.Pix) {
			d = &image.Gray{Pix: make([]byte, len(s.Pix))}
		}
		copy(d.Pix, s.Pix)
		d.Stride, d.Rect = s.Stride, s.Rect

		return d
	}

	return src
}

// capture reads the newest frame until quit; Read() paces the loop, fps caps it.
// When showFPS is set it reports the measured rate on a single overwritten line.
func capture(cam *camera.Camera, s *stream, fps int, showFPS bool, done <-chan struct{}) {
	interval := time.Second / time.Duration(fps)
	next := time.Now()

	frames := 0
	mark := time.Now()
	reported := false

	if showFPS {
		defer func() {
			if reported {
				fmt.Fprintln(os.Stderr)
			}
		}()
	}

	for {
		select {
		case <-done:
			return
		default:
		}

		img, err := cam.Read()
		if err != nil {
			select {
			case <-done:
				return
			default:
			}

			stderr("%s\n", err.Error())

			select {
			case <-done:
				return
			case <-time.After(100 * time.Millisecond):
			}

			next = time.Now()

			continue
		}

		s.set(copyFrame(s.buffer(), img))

		if frames++; showFPS {
			if d := time.Since(mark); d >= time.Second {
				fmt.Fprintf(os.Stderr, "\rFPS: %5.1f", float64(frames)/d.Seconds())
				reported = true
				frames = 0
				mark = time.Now()
			}
		}

		if next = next.Add(interval); time.Until(next) <= 0 {
			next = time.Now()

			continue
		}

		select {
		case <-done:
			return
		case <-time.After(time.Until(next)):
		}
	}
}

func appIcon() image.Image {
	img, err := png.Decode(bytes.NewReader(iconPNG))
	if err != nil {
		return nil
	}

	return img
}

// banner prints the device and negotiated format on startup.
func banner(index int, info camera.Info) {
	fmt.Fprintf(os.Stderr, "%s %s\n", name, version)

	if n := deviceName(index); n != "" {
		fmt.Fprintf(os.Stderr, "  device  %s\n", n)
	}

	fmt.Fprintf(os.Stderr, "  format  %s %dx%d\n", info.Format, info.Width, info.Height)
}

func title(index int) string {
	if n := deviceName(index); n != "" {
		return fmt.Sprintf("%s - %s", name, n)
	}

	return name
}

func stderr(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}

// isTerminal reports whether f is a character device (a terminal).
func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}

	return stat.Mode()&os.ModeCharDevice != 0
}

// deviceName returns the name of the camera at index, or an empty string.
func deviceName(index int) string {
	devices, err := camera.Devices()
	if err != nil {
		return ""
	}

	for _, d := range devices {
		if d.Index == index {
			return d.Name
		}
	}

	return ""
}

// deviceIndex returns the index of the first camera whose name contains the query.
func deviceIndex(name string) (int, error) {
	devices, err := camera.Devices()
	if err != nil {
		return 0, err
	}

	want := strings.ToLower(name)
	for _, d := range devices {
		if strings.Contains(strings.ToLower(d.Name), want) {
			return d.Index, nil
		}
	}

	return 0, fmt.Errorf("camera: no device matching %q", name)
}
