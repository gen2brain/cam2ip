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
	flag.Float64Var(&srv.FrameWidth, "width", 640, "Frame width [CAM2IP_WIDTH]")
	flag.Float64Var(&srv.FrameHeight, "height", 480, "Frame height [CAM2IP_HEIGHT]")
	flag.IntVar(&srv.Rotate, "rotate", 0, "Rotate image, valid values are 90, 180, 270 [CAM2IP_ROTATE]")
	flag.BoolVar(&srv.NoWebGL, "no-webgl", false, "Disable WebGL drawing of images (html handler) [CAM2IP_NO_WEBGL]")
	flag.BoolVar(&srv.Timestamp, "timestamp", false, "Draws timestamp on images [CAM2IP_TIMESTAMP]")
	flag.StringVar(&srv.Bind, "bind-addr", ":56000", "Bind address [CAM2IP_BIND_ADDR]")
	flag.StringVar(&srv.Htpasswd, "htpasswd-file", "", "Path to htpasswd file, if empty auth is disabled [CAM2IP_HTPASSWD_FILE]")

	flag.Usage = func() {
		stderr("Usage: %s [<flags>]\n", name)
		order := []string{"index", "delay", "width", "height", "rotate", "no-webgl", "timestamp", "bind-addr", "htpasswd-file"}

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
		Index:     srv.Index,
		Rotate:    srv.Rotate,
		Width:     srv.FrameWidth,
		Height:    srv.FrameHeight,
		Timestamp: srv.Timestamp,
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
