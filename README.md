## cam2ip

Turn any webcam into ip camera.

Example (in web browser):

    http://localhost:56000/mjpeg
or

    http://localhost:56000/html

### Requirements

* [OpenCV 2.x](http://opencv.org/)


### Download

Binaries are compiled with static OpenCV library:

 - [Android 32bit](https://github.com/gen2brain/cam2ip/releases/download/1.0/cam2ip-1.0-android.tar.gz)
 - [Linux 64bit](https://github.com/gen2brain/cam2ip/releases/download/1.0/cam2ip-1.0-64bit.tar.gz)
 - [RPi 32bit](https://github.com/gen2brain/cam2ip/releases/download/1.0/cam2ip-1.0-RPi.tar.gz)
 - [Windows 32bit](https://github.com/gen2brain/cam2ip/releases/download/1.0/cam2ip-1.0.zip)


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
  -frame-height float
    	Frame height (default 480)
  -frame-width float
    	Frame width (default 640)
  -htpasswd-file string
    	Path to htpasswd file, if empty auth is disabled
  -index int
    	Camera index
```

### Handlers

  * `/html`: HTML handler, frames are pushed to canvas over websocket
  * `/jpeg`: Static JPEG handler
  * `/mjpeg`: Motion JPEG, supported natively in major web browsers
