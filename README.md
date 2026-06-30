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

* `vfw` - on Windows, use `Video for Windows` to access camera instead of the default `Media Foundation`
* `libjpeg` - build with `libjpeg` ([go-libjpeg](https://github.com/pixiv/go-libjpeg)) instead of native `image/jpeg`

### Download

Download the latest binaries from the [releases](https://github.com/gen2brain/cam2ip/releases).

### Installation

    go install github.com/gen2brain/cam2ip/cmd/cam2ip@latest

This command will install `cam2ip` in `GOBIN`, you can point `GOBIN` to e.g. `/usr/local/bin` or `~/.local/bin`.

### Usage

```
Usage: cam2ip [<flags>]
  --index
    	Camera index [CAM2IP_INDEX] (default "0")
  --device
    	Camera name to use, matched as substring, overrides index [CAM2IP_DEVICE] (default "")
  --delay
    	Delay between frames, in milliseconds [CAM2IP_DELAY] (default "10")
  --width
    	Frame width [CAM2IP_WIDTH] (default "640")
  --height
    	Frame height [CAM2IP_HEIGHT] (default "480")
  --quality
    	Image quality [CAM2IP_QUALITY] (default "75")
  --rotate
    	Rotate image, valid values are 90, 180, 270 [CAM2IP_ROTATE] (default "0")
  --flip
    	Flip image, valid values are horizontal and vertical [CAM2IP_FLIP] (default "")
  --no-webgl
    	Disable WebGL drawing of image (html handler) [CAM2IP_NO_WEBGL] (default "false")
  --timestamp
    	Draws timestamp on image [CAM2IP_TIMESTAMP] (default "false")
  --time-format
    	Time format [CAM2IP_TIME_FORMAT] (default "2006-01-02 15:04:05")
  --bind-addr
    	Bind address [CAM2IP_BIND_ADDR] (default ":56000")
  --htpasswd-file
    	Path to htpasswd file, if empty auth is disabled [CAM2IP_HTPASSWD_FILE] (default "")
  --list-devices
    	List available cameras and exit (default "false")
  --version
    	Print version and exit (default "false")
```

### Handlers

  * `/html`: HTML handler, frames are pushed to canvas over websocket
  * `/jpeg`: Static JPEG handler
  * `/mjpeg`: Motion JPEG, supported natively in major web browsers
