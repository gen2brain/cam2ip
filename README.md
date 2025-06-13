## cam2ip

Turn any webcam into an IP camera.

Example (in web browser):

    http://localhost:56000/html

or

    http://localhost:56000/mjpeg

You can also use apps like `ffplay` or `vlc`:

    ffplay -i http://localhost:56000/mjpeg

### Requirements

* On Linux/RPi native Go [V4L](https://github.com/korandiz/v4l) implementation is used to capture images.
* On Windows [Video for Windows (VfW)](https://en.wikipedia.org/wiki/Video_for_Windows) framework is used over win32 API.

### Build tags

* `opencv` - use `OpenCV` library to access camera ([gocv](https://github.com/hybridgroup/gocv))
* `libjpeg` - build with `libjpeg` ([go-libjpeg](https://github.com/pixiv/go-libjpeg)) instead of native Go `image/jpeg`

### Download

Download the latest binaries from the [releases](https://github.com/gen2brain/cam2ip/releases).

### Installation

    go install github.com/gen2brain/cam2ip/cmd/cam2ip@latest

This command will install `cam2ip` in `GOBIN`, you can point `GOBIN` to e.g. `/usr/local/bin` or `~/.local/bin`.

### Run in Docker container

    docker run --device=/dev/video0:/dev/video0 -p56000:56000 -it gen2brain/cam2ip # on RPi use gen2brain/cam2ip:arm

### Usage

```
Usage of cam2ip:
  -bind-addr string
	Bind address [CAM2IP_BIND_ADDR] (default ":56000")
  -delay int
	Delay between frames, in milliseconds [CAM2IP_DELAY] (default 10)
  -height float
	Frame height [CAM2IP_HEIGHT] (default 480)
  -htpasswd-file string
	Path to htpasswd file, if empty auth is disabled [CAM2IP_HTPASSWD_FILE]
  -index int
	Camera index [CAM2IP_INDEX]
  -nowebgl
	Disable WebGL drawing of images (html handler) [CAM2IP_NOWEBGL]
  -rotate int
	Rotate image, valid values are 90, 180, 270 [CAM2IP_ROTATE]
  -timestamp
	Draws timestamp on images [CAM2IP_TIMESTAMP]
  -width float
	Frame width [CAM2IP_WIDTH] (default 640)
```

### Handlers

  * `/html`: HTML handler, frames are pushed to canvas over websocket
  * `/jpeg`: Static JPEG handler
  * `/mjpeg`: Motion JPEG, supported natively in major web browsers
