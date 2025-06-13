package main

import (
	"flag"
	"fmt"
	"os"

	"go.senan.xyz/flagconf"

	"github.com/gen2brain/cam2ip/camera"
	"github.com/gen2brain/cam2ip/server"
)

const (
	name    = "cam2ip"
	version = "1.6"
)

func main() {
	srv := server.NewServer()

	flag.IntVar(&srv.Index, "index", 0, "Camera index [CAM2IP_INDEX]")
	flag.IntVar(&srv.Delay, "delay", 10, "Delay between frames, in milliseconds [CAM2IP_DELAY]")
	flag.Float64Var(&srv.Width, "width", 640, "Frame width [CAM2IP_WIDTH]")
	flag.Float64Var(&srv.Height, "height", 480, "Frame height [CAM2IP_HEIGHT]")
	flag.IntVar(&srv.Quality, "quality", 75, "Image quality [CAM2IP_QUALITY]")
	flag.IntVar(&srv.Rotate, "rotate", 0, "Rotate image, valid values are 90, 180, 270 [CAM2IP_ROTATE]")
	flag.StringVar(&srv.Flip, "flip", "", "Flip image, valid values are horizontal and vertical [CAM2IP_FLIP]")
	flag.BoolVar(&srv.NoWebGL, "no-webgl", false, "Disable WebGL drawing of image (html handler) [CAM2IP_NO_WEBGL]")
	flag.BoolVar(&srv.Timestamp, "timestamp", false, "Draws timestamp on image [CAM2IP_TIMESTAMP]")
	flag.StringVar(&srv.TimeFormat, "time-format", "2006-01-02 15:04:05", "Time format [CAM2IP_TIME_FORMAT]")
	flag.StringVar(&srv.Bind, "bind-addr", ":56000", "Bind address [CAM2IP_BIND_ADDR]")
	flag.StringVar(&srv.Htpasswd, "htpasswd-file", "", "Path to htpasswd file, if empty auth is disabled [CAM2IP_HTPASSWD_FILE]")

	flag.Usage = func() {
		stderr("Usage: %s [<flags>]\n", name)
		order := []string{"index", "delay", "width", "height", "quality", "rotate", "flip", "no-webgl",
			"timestamp", "time-format", "bind-addr", "htpasswd-file"}

		for _, name := range order {
			f := flag.Lookup(name)
			if f != nil {
				stderr("  --%s\n    \t%v (default %q)\n", f.Name, f.Usage, f.DefValue)
			}
		}
	}

	flag.Parse()
	_ = flagconf.ParseEnv()

	srv.Name = name
	srv.Version = version

	if srv.Htpasswd != "" {
		if _, err := os.Stat(srv.Htpasswd); err != nil {
			stderr("%s\n", err.Error())
			os.Exit(1)
		}
	}

	cam, err := camera.New(camera.Options{
		Index:      srv.Index,
		Rotate:     srv.Rotate,
		Flip:       srv.Flip,
		Width:      srv.Width,
		Height:     srv.Height,
		Timestamp:  srv.Timestamp,
		TimeFormat: srv.TimeFormat,
	})
	if err != nil {
		stderr("%s\n", err.Error())
		os.Exit(1)
	}

	srv.Reader = cam

	defer srv.Reader.Close()

	stderr("Listening on %s\n", srv.Bind)

	err = srv.ListenAndServe()
	if err != nil {
		stderr("%s\n", err.Error())
		os.Exit(1)
	}
}

func stderr(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
}
