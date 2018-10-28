## cam2ip

Turn any webcam into an IP camera.

Example (in web browser):

    http://localhost:56000/mjpeg
or

    http://localhost:56000/html

### Requirements

* [OpenCV](http://opencv.org/)
* [libjpeg-turbo](https://www.libjpeg-turbo.org/)

### Build tags

* `cv3` - build with OpenCV 3.x [gocv](https://github.com/hybridgroup/gocv), default is version 2.x via [go-opencv](https://github.com/lazywei/go-opencv)
* `native` - build with native Go [V4L](https://github.com/korandiz/v4l) implementation on Linux/RPi instead of `OpenCV`
* `jpeg` - build with native Go `image/jpeg` instead of `libjpeg-turbo`

### Download

Binaries are compiled with static OpenCV/libjpeg-turbo libraries, they should just work:

 - [Linux 64bit](https://github.com/gen2brain/cam2ip/releases/download/1.4/cam2ip-1.4-64bit.tar.gz)
 - [RPi 32bit](https://github.com/gen2brain/cam2ip/releases/download/1.4/cam2ip-1.4-RPi.tar.gz)
 - [RPi3 32bit](https://github.com/gen2brain/cam2ip/releases/download/1.4/cam2ip-1.4-RPi3.tar.gz)
 - [Windows 32bit](https://github.com/gen2brain/cam2ip/releases/download/1.4/cam2ip-1.4.zip)
 - [Windows 64bit](https://github.com/gen2brain/cam2ip/releases/download/1.4/cam2ip-1.4-64bit.zip)


### Installation

    go get -v github.com/gen2brain/cam2ip

This will install app in `$GOPATH/bin/cam2ip`.

### Usage

```
Usage of ./cam2ip:
  -bind-addr string
        Bind address (default ":56000")
  -delay int
        Delay between frames, in milliseconds (default 10)
  -height float
        Frame height (default 480)
  -width float
        Frame width (default 640)
  -htpasswd-file string
        Path to htpasswd file, if empty auth is disabled
  -index int
        Camera index
  -nowebgl
        Disable WebGL drawing of images (html handler)
  -rotate int
    	Rotate image, valid values are 90, 180, 270
  -video-file string
    	Use video file instead of camera
```

### Handlers

  * `/html`: HTML handler, frames are pushed to canvas over websocket
  * `/jpeg`: Static JPEG handler
  * `/mjpeg`: Motion JPEG, supported natively in major web browsers
