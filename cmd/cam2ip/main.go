package main

import (
	"flag"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"go.senan.xyz/flagconf"

	"github.com/gen2brain/cam2ip/camera"
	"github.com/gen2brain/cam2ip/server"
)

const name = "cam2ip"

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
	srv := server.NewServer()

	flag.IntVar(&srv.Index, "index", 0, "Camera index [CAM2IP_INDEX]")
	flag.StringVar(&srv.Device, "device", "", "Camera name to use, matched as substring, overrides index [CAM2IP_DEVICE]")
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

	var showVersion bool
	flag.BoolVar(&showVersion, "version", false, "Print version and exit")

	var listDevices bool
	flag.BoolVar(&listDevices, "list-devices", false, "List available cameras and exit")

	flag.Usage = func() {
		color := useColor(os.Stderr)

		stderr("%s %s [<flags>]\n", colorize(color, colorBold, "Usage:"), name)
		order := []string{"index", "device", "delay", "width", "height", "quality", "rotate", "flip", "no-webgl",
			"timestamp", "time-format", "bind-addr", "htpasswd-file", "list-devices", "version"}

		for _, name := range order {
			f := flag.Lookup(name)
			if f != nil {
				stderr("  %s\n    \t%v (default %q)\n", colorize(color, colorCyan, "--"+f.Name), f.Usage, f.DefValue)
			}
		}
	}

	flag.Parse()
	_ = flagconf.ParseEnv()

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
			fmt.Printf("%d: %s\n", d.Index, d.Name)
		}

		os.Exit(0)
	}

	if srv.Device != "" {
		index, err := deviceIndex(srv.Device)
		if err != nil {
			stderr("%s\n", err.Error())
			os.Exit(1)
		}

		srv.Index = index
	}

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

	stderr("%s %s listening on %s\n", name, version, srv.Bind)

	err = srv.ListenAndServe()
	if err != nil {
		stderr("%s\n", err.Error())
		os.Exit(1)
	}
}

func stderr(format string, a ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, a...)
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

const (
	colorBold = "\033[1m"
	colorCyan = "\033[36m"
)

func useColor(f *os.File) bool {
	if _, ok := os.LookupEnv("NO_COLOR"); ok {
		return false
	}

	stat, err := f.Stat()
	if err != nil {
		return false
	}

	return stat.Mode()&os.ModeCharDevice != 0
}

func colorize(on bool, code, s string) string {
	if !on {
		return s
	}

	return code + s + "\033[0m"
}
