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

* `opencv` - use `OpenCV` to access camera ([gocv](https://github.com/hybridgroup/gocv))
* `turbo` - build with `libjpeg-turbo` ([libjpeg-turbo](https://www.libjpeg-turbo.org/)) instead of native Go `image/jpeg`

### Download

Binaries are compiled with static OpenCV/libjpeg-turbo libraries, they should just work:

 - [Linux 64bit](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-64bit.tar.gz)
 - [Linux 64bit OpenCV](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-64bit-cv2.tar.gz)
 - [macOS 64bit OpenCV](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-darwin-cv2.zip)
 - [RPi 32bit](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-RPi.tar.gz)
 - [RPi 32bit OpenCV](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-RPi-cv2.tar.gz)
 - [RPi 32bit Static](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-RPi-nocgo.tar.gz)
 - [RPi3 32bit](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-RPi3.tar.gz)
 - [RPi3 32bit OpenCV](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-RPi3-cv2.tar.gz)
 - [Windows 32bit](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-32bit.zip)
 - [Windows 32bit OpenCV](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-32bit-cv2.zip)
 - [Windows 64bit](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-64bit.zip)
 - [Windows 64bit OpenCV](https://github.com/gen2brain/cam2ip/releases/download/1.6/cam2ip-1.6-64bit-cv2.zip)


### Installation

    go get -v github.com/gen2brain/cam2ip/cmd/cam2ip

This will install app in `$GOPATH/bin/cam2ip`.

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
