//go:build !opencv

// Package camera.
package camera

import (
	"bytes"
	"fmt"
	"image"
	"runtime"
	"syscall"
	"unsafe"

	im "github.com/gen2brain/cam2ip/image"
)

// Camera represents camera.
type Camera struct {
	opts      Options
	camera    syscall.Handle
	rgba      *image.RGBA
	ycbcr     *image.YCbCr
	hdr       *videoHdr
	instance  syscall.Handle
	className string
	format    uint32
	bpp       int
}

// New returns new Camera for given camera index.
func New(opts Options) (camera *Camera, err error) {
	camera = &Camera{}
	camera.opts = opts
	camera.className = "capWindowClass"

	ready := make(chan error, 1)

	go camera.run(ready)

	if err = <-ready; err != nil {
		camera = nil

		return
	}

	return
}

// run sets up the capture window on a locked OS thread, reports the result over ready, then pumps messages.
func (c *Camera) run(ready chan<- error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	fn := func(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
		switch msg {
		case wmClose:
			_ = destroyWindow(hwnd)
		case wmDestroy:
			postQuitMessage(0)
		default:
			return defWindowProc(hwnd, msg, wparam, lparam)
		}

		return 0
	}

	instance, err := getModuleHandle()
	if err != nil {
		ready <- err

		return
	}
	c.instance = instance

	if err := registerClass(c.className, c.instance, fn); err != nil {
		ready <- err

		return
	}

	hwnd, err := createWindow(0, c.className, "", wsOverlappedWindow, cwUseDefault, cwUseDefault,
		int64(c.opts.Width)+100, int64(c.opts.Height)+100, 0, 0, c.instance)
	if err != nil {
		ready <- err

		return
	}

	c.camera, err = capCreateCaptureWindow("", wsChild, 0, 0, int64(c.opts.Width), int64(c.opts.Height), hwnd, 0)
	if err != nil {
		ready <- err

		return
	}

	ret := sendMessage(c.camera, wmCapDriverConnect, uintptr(c.opts.Index), 0)
	if int(ret) == 0 {
		ready <- fmt.Errorf("camera: can not open camera %d", c.opts.Index)

		return
	}

	sendMessage(c.camera, wmCapSetPreview, 0, 0)
	sendMessage(c.camera, wmCapSetOverlay, 0, 0)

	var bi bitmapInfo
	size := sendMessage(c.camera, wmCapGetVideoformat, 0, 0)
	sendMessage(c.camera, wmCapGetVideoformat, size, uintptr(unsafe.Pointer(&bi)))

	headerSize := uintptr(unsafe.Sizeof(bi.BmiHeader))

	// Request an uncompressed format; the WDM mapper converts, and MJPG via grab is unreliable.
	for _, f := range []struct {
		compression uint32
		bitCount    uint16
	}{
		{biRGB, 24},
		{biRGB, 32},
		{yuy2FourCC, 16},
	} {
		stride := ((int(c.opts.Width)*int(f.bitCount) + 31) / 32) * 4

		bi.BmiHeader.BiSize = uint32(headerSize)
		bi.BmiHeader.BiPlanes = 1
		bi.BmiHeader.BiWidth = int32(c.opts.Width)
		bi.BmiHeader.BiHeight = int32(c.opts.Height)
		bi.BmiHeader.BiCompression = f.compression
		bi.BmiHeader.BiBitCount = f.bitCount
		bi.BmiHeader.BiSizeImage = uint32(stride * int(c.opts.Height))

		if int(sendMessage(c.camera, wmCapSetVideoformat, headerSize, uintptr(unsafe.Pointer(&bi)))) != 0 {
			break
		}
	}

	// Read back the format the driver actually accepted.
	sendMessage(c.camera, wmCapGetVideoformat, headerSize, uintptr(unsafe.Pointer(&bi)))
	c.format = bi.BmiHeader.BiCompression

	sendMessage(c.camera, wmCapSetCallbackFrame, 0, syscall.NewCallback(c.callback))

	width := int(bi.BmiHeader.BiWidth)
	height := int(bi.BmiHeader.BiHeight)
	if height < 0 {
		height = -height
	}

	rect := image.Rect(0, 0, width, height)

	switch {
	case c.format == 0:
		if bi.BmiHeader.BiBitCount != 24 && bi.BmiHeader.BiBitCount != 32 {
			ready <- fmt.Errorf("camera: unsupported RGB bitcount %d", bi.BmiHeader.BiBitCount)

			return
		}

		c.bpp = int(bi.BmiHeader.BiBitCount) / 8
		c.rgba = image.NewRGBA(rect)
	case is422Format(c.format):
		c.ycbcr = image.NewYCbCr(rect, image.YCbCrSubsampleRatio422)
	case c.format == mjpgFourCC:
	default:
		ready <- fmt.Errorf("camera: unsupported format %d", c.format)

		return
	}

	ready <- nil

	for {
		var msg msgW
		ok, _ := getMessage(&msg, 0, 0, 0)
		if !ok {
			break
		}

		dispatchMessage(&msg)
	}
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	ret := sendMessage(c.camera, wmCapGrabFrame, 0, 0)
	if int(ret) == 0 {
		err = fmt.Errorf("camera: can not grab frame")

		return
	}

	if c.hdr == nil {
		err = fmt.Errorf("camera: no frame available")

		return
	}

	data := unsafe.Slice((*byte)(unsafe.Pointer(c.hdr.LpData)), c.hdr.DwBufferLength)

	switch {
	case c.format == 0:
		e := bmpToRgba(data, c.rgba, c.bpp)
		if e != nil {
			err = fmt.Errorf("camera: format %d: can not retrieve frame: %w", c.format, e)

			return
		}

		img = c.rgba
	case is422Format(c.format):
		y0, y1, cb, cr, _ := packed422Offsets(c.format)

		e := packedYUV422ToYCbCr(data, c.ycbcr, y0, y1, cb, cr)
		if e != nil {
			err = fmt.Errorf("camera: format %d: can not retrieve frame: %w", c.format, e)

			return
		}

		img = c.ycbcr
	case c.format == mjpgFourCC:
		i, e := im.NewDecoder(bytes.NewReader(data)).Decode()
		if e != nil {
			err = fmt.Errorf("camera: format %d: can not retrieve frame: %w", c.format, e)

			return
		}

		img = i
	}

	if c.opts.Rotate != 0 {
		img = im.Rotate(img, c.opts.Rotate)
	}

	if c.opts.Flip != "" {
		img = im.Flip(img, c.opts.Flip)
	}

	if c.opts.Timestamp {
		img = im.Timestamp(img, c.opts.TimeFormat)
	}

	return
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	sendMessage(c.camera, wmCapSetCallbackFrame, 0, 0)
	sendMessage(c.camera, wmCapDriverDisconnect, 0, 0)

	err = destroyWindow(c.camera)

	unregisterClass(c.className, c.instance)

	return err
}

// callback function.
func (c *Camera) callback(hwnd syscall.Handle, hdr *videoHdr) uintptr {
	if hdr != nil {
		c.hdr = hdr
	}

	return 0
}

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	avicap32 = syscall.NewLazyDLL("avicap32.dll")

	createWindowExW  = user32.NewProc("CreateWindowExW")
	destroyWindowW   = user32.NewProc("DestroyWindow")
	defWindowProcW   = user32.NewProc("DefWindowProcW")
	dispatchMessageW = user32.NewProc("DispatchMessageW")
	getMessageW      = user32.NewProc("GetMessageW")
	sendMessageW     = user32.NewProc("SendMessageW")
	postQuitMessageW = user32.NewProc("PostQuitMessage")
	registerClassExW = user32.NewProc("RegisterClassExW")
	unregisterClassW = user32.NewProc("UnregisterClassW")

	getModuleHandleW        = kernel32.NewProc("GetModuleHandleW")
	capCreateCaptureWindowW = avicap32.NewProc("capCreateCaptureWindowW")
)

const (
	wmDestroy = 0x0002
	wmClose   = 0x0010
	wmUser    = 0x0400

	wmCapStart            = wmUser
	wmCapSetCallbackFrame = wmCapStart + 5
	wmCapDriverConnect    = wmCapStart + 10
	wmCapDriverDisconnect = wmCapStart + 11
	wmCapGetVideoformat   = wmCapStart + 44
	wmCapSetVideoformat   = wmCapStart + 45
	wmCapSetPreview       = wmCapStart + 50
	wmCapSetOverlay       = wmCapStart + 51
	wmCapGrabFrame        = wmCapStart + 60
	wmCapGrabFrameNoStop  = wmCapStart + 61
	wmCapStop             = wmCapStart + 68
	wmCapAbort            = wmCapStart + 69

	wsChild            = 0x40000000
	wsOverlappedWindow = 0x00CF0000

	cwUseDefault = 0x7fffffff

	biRGB = 0
)

// wndClassExW https://msdn.microsoft.com/en-us/library/windows/desktop/ms633577.aspx
type wndClassExW struct {
	size       uint32
	style      uint32
	wndProc    uintptr
	clsExtra   int32
	wndExtra   int32
	instance   syscall.Handle
	icon       syscall.Handle
	cursor     syscall.Handle
	background syscall.Handle
	menuName   *uint16
	className  *uint16
	iconSm     syscall.Handle
}

// msgW https://msdn.microsoft.com/en-us/library/windows/desktop/ms644958.aspx
type msgW struct {
	hwnd    syscall.Handle
	message uint32
	wParam  uintptr
	lParam  uintptr
	time    uint32
	pt      pointW
}

// https://msdn.microsoft.com/en-us/ecb0f0e1-90c2-48ab-a069-552262b49c7c
type pointW struct {
	x, y int32
}

// http://msdn.microsoft.com/en-us/library/windows/desktop/dd183376.aspx
type bitmapInfoHeader struct {
	BiSize          uint32
	BiWidth         int32
	BiHeight        int32
	BiPlanes        uint16
	BiBitCount      uint16
	BiCompression   uint32
	BiSizeImage     uint32
	BiXPelsPerMeter int32
	BiYPelsPerMeter int32
	BiClrUsed       uint32
	BiClrImportant  uint32
}

// http://msdn.microsoft.com/en-us/library/windows/desktop/dd183375.aspx
type bitmapInfo struct {
	BmiHeader bitmapInfoHeader
	BmiColors *rgbQuad
}

// http://msdn.microsoft.com/en-us/library/windows/desktop/dd162938.aspx
type rgbQuad struct {
	RgbBlue     byte
	RgbGreen    byte
	RgbRed      byte
	RgbReserved byte
}

// https://docs.microsoft.com/en-us/windows/desktop/api/vfw/ns-vfw-videohdr_tag
type videoHdr struct {
	LpData         *uint8
	DwBufferLength uint32
	DwBytesUsed    uint32
	DwTimeCaptured uint32
	DwUser         uint64
	DwFlags        uint32
	DwReserved     [4]uint64
}

// https://docs.microsoft.com/en-us/windows/desktop/api/libloaderapi/nf-libloaderapi-getmodulehandlea
func getModuleHandle() (syscall.Handle, error) {
	ret, _, err := getModuleHandleW.Call(uintptr(0))
	if ret == 0 {
		return 0, err
	}

	return syscall.Handle(ret), nil
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-createwindowexw
func createWindow(exStyle uint64, className, windowName string, style uint64, x, y, width, height int64,
	parent, menu, instance syscall.Handle) (syscall.Handle, error) {
	ret, _, err := createWindowExW.Call(uintptr(exStyle), uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className))),
		uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(windowName))), uintptr(style), uintptr(x), uintptr(y),
		uintptr(width), uintptr(height), uintptr(parent), uintptr(menu), uintptr(instance), uintptr(0))

	if ret == 0 {
		return 0, err
	}

	return syscall.Handle(ret), nil
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-destroywindow
func destroyWindow(hwnd syscall.Handle) error {
	ret, _, err := destroyWindowW.Call(uintptr(hwnd))
	if ret == 0 {
		return err
	}

	return nil
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-defwindowprocw
func defWindowProc(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
	ret, _, _ := defWindowProcW.Call(uintptr(hwnd), uintptr(msg), wparam, lparam)

	return ret
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-dispatchmessagew
func dispatchMessage(msg *msgW) {
	dispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-getmessagew
func getMessage(msg *msgW, hwnd syscall.Handle, msgFilterMin, msgFilterMax uint32) (bool, error) {
	ret, _, err := getMessageW.Call(uintptr(unsafe.Pointer(msg)), uintptr(hwnd), uintptr(msgFilterMin), uintptr(msgFilterMax))
	if int32(ret) == -1 {
		return false, err
	}

	return int32(ret) != 0, nil
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-sendmessage
func sendMessage(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
	ret, _, _ := sendMessageW.Call(uintptr(hwnd), uintptr(msg), wparam, lparam, 0, 0)

	return ret
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-postquitmessage
func postQuitMessage(exitCode int32) {
	postQuitMessageW.Call(uintptr(exitCode))
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-registerclassexw
func registerClass(className string, instance syscall.Handle, fn interface{}) error {
	var wcx wndClassExW
	wcx.size = uint32(unsafe.Sizeof(wcx))
	wcx.wndProc = syscall.NewCallback(fn)
	wcx.instance = instance
	wcx.className = syscall.StringToUTF16Ptr(className)

	ret, _, err := registerClassExW.Call(uintptr(unsafe.Pointer(&wcx)))
	if ret == 0 {
		return err
	}

	return nil
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-unregisterclassw
func unregisterClass(className string, instance syscall.Handle) bool {
	ret, _, _ := unregisterClassW.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(className))), uintptr(instance))

	return ret != 0
}

// https://docs.microsoft.com/en-us/windows/desktop/api/vfw/nf-vfw-capcreatecapturewindoww
func capCreateCaptureWindow(lpszWindowName string, dwStyle, x, y, width, height int64, parent syscall.Handle, id int64) (syscall.Handle, error) {
	ret, _, err := capCreateCaptureWindowW.Call(uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(lpszWindowName))),
		uintptr(dwStyle), uintptr(x), uintptr(y), uintptr(width), uintptr(height), uintptr(parent), uintptr(id))
	if ret == 0 {
		return 0, err
	}

	return syscall.Handle(ret), nil
}
