package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/gen2brain/cam2ip/camera"
	"github.com/gen2brain/cam2ip/server"
)

const (
	name    = "cam2ip"
	version = "1.0"
)

func main() {
	srv := server.NewServer()

	flag.IntVar(&srv.Index, "index", 0, "Camera index")
	flag.IntVar(&srv.Delay, "delay", 10, "Delay between frames, in milliseconds")
	flag.Float64Var(&srv.FrameWidth, "frame-width", 640, "Frame width")
	flag.Float64Var(&srv.FrameHeight, "frame-height", 480, "Frame height")
	flag.BoolVar(&srv.NoWebGL, "nowebgl", false, "Disable WebGL drawing of images (html handler)")
	flag.StringVar(&srv.Bind, "bind-addr", ":56000", "Bind address")
	flag.StringVar(&srv.Htpasswd, "htpasswd-file", "", "Path to htpasswd file, if empty auth is disabled")
	flag.Parse()

	srv.Name = name
	srv.Version = version

	var err error

	if srv.Htpasswd != "" {
		if _, err = os.Stat(srv.Htpasswd); err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
			os.Exit(1)
		}
	}

	srv.Camera, err = camera.New(srv.Index)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}

	srv.Camera.SetProperty(camera.PropFrameWidth, srv.FrameWidth)
	srv.Camera.SetProperty(camera.PropFrameHeight, srv.FrameHeight)

	defer srv.Camera.Close()

	fmt.Fprintf(os.Stderr, "Listening on %s\n", srv.Bind)

	err = srv.ListenAndServe()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		os.Exit(1)
	}
}
