// +build !cv2,!cv3

// Package camera.
package camera

import (
	"bytes"
	"fmt"
	"image"
	"syscall"
	"unsafe"

	"github.com/disintegration/imaging"
)

// Options.
type Options struct {
	Index  int
	Rotate int
	Width  float64
	Height float64
}

// Camera represents camera.
type Camera struct {
	opts      Options
	camera    syscall.Handle
	frame     *image.RGBA
	hdr       *videoHdr
	instance  syscall.Handle
	className string
}

// New returns new Camera for given camera index.
func New(opts Options) (camera *Camera, err error) {
	camera = &Camera{}
	camera.opts = opts
	camera.className = "capWindowClass"

	camera.instance, err = getModuleHandle()
	if err != nil {
		return
	}

	camera.frame = image.NewRGBA(image.Rect(0, 0, int(camera.opts.Width), int(camera.opts.Height)))

	go func(c *Camera) {
		fn := func(hwnd syscall.Handle, msg uint32, wparam, lparam uintptr) uintptr {
			switch msg {
			case wmClose:
				destroyWindow(hwnd)
			case wmDestroy:
				postQuitMessage(0)
			default:
				ret := defWindowProc(hwnd, msg, wparam, lparam)
				return ret
			}

			return 0
		}

		err = registerClass(c.className, c.instance, fn)
		if err != nil {
			return
		}

		hwnd, err := createWindow(0, c.className, "", wsOverlappedWindow, cwUseDefault, cwUseDefault, int64(c.opts.Width)+100, int64(c.opts.Height)+100, 0, 0, c.instance)
		if err != nil {
			return
		}

		c.camera, err = capCreateCaptureWindow("", wsChild, 0, 0, int64(c.opts.Width), int64(c.opts.Height), hwnd, 0)
		if err != nil {
			return
		}

		ret := sendMessage(c.camera, wmCapDriverConnect, uintptr(c.opts.Index), 0)
		if bool(int(ret) == 0) {
			err = fmt.Errorf("camera: can not open camera %d", c.opts.Index)
			return
		}

		var bi bitmapInfo
		size := sendMessage(c.camera, wmCapGetVideoformat, 0, 0)
		sendMessage(c.camera, wmCapGetVideoformat, size, uintptr(unsafe.Pointer(&bi)))

		bi.BmiHeader.BiWidth = int32(c.opts.Width)
		bi.BmiHeader.BiHeight = int32(c.opts.Height)

		ret = sendMessage(c.camera, wmCapSetVideoformat, size, uintptr(unsafe.Pointer(&bi)))
		if bool(int(ret) == 0) {
			err = fmt.Errorf("camera: can not set video format")
			return
		}

		sendMessage(c.camera, wmCapSetCallbackFrame, 0, syscall.NewCallback(c.callback))

		messageLoop(c.camera)
	}(camera)

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	ret := sendMessage(c.camera, wmCapGrabFrameNoStop, 0, 0)
	if bool(int(ret) == 0) {
		err = fmt.Errorf("camera: can not grab frame")
		return
	}

	data := (*[1 << 24]uint8)(unsafe.Pointer(c.hdr.LpData))[0:c.hdr.DwBytesUsed]
	r := bytes.NewReader(data)

	width := int(c.opts.Width)
	height := int(c.opts.Height)

	// Taken from https://github.com/hotei/bmp/blob/master/bmpRGBA.go#L12
	// There are 3 bytes per pixel, and each row is 4-byte aligned.
	b := make([]byte, (3*width+3)&^3)
	// BMP images are stored bottom-up rather than top-down.
	for y := height - 1; y >= 0; y-- {
		_, err = r.Read(b)
		if err != nil {
			err = fmt.Errorf("camera: can not retrieve frame: %v", err)
			return
		}

		p := c.frame.Pix[y*c.frame.Stride : y*c.frame.Stride+width*4]
		for i, j := 0, 0; i < len(p); i, j = i+4, j+3 {
			// BMP images are stored in BGR order rather than RGB order.
			p[i+0] = b[j+2]
			p[i+1] = b[j+1]
			p[i+2] = b[j+0]
			p[i+3] = 0xFF
		}
	}

	img = c.frame
	if c.opts.Rotate == 0 {
		return
	}

	switch c.opts.Rotate {
	case 90:
		img = imaging.Rotate90(img)
	case 180:
		img = imaging.Rotate180(img)
	case 270:
		img = imaging.Rotate270(img)
	}

	return
}

// GetProperty returns the specified camera property.
func (c *Camera) GetProperty(id int) float64 {
	return 0
}

// SetProperty sets a camera property.
func (c *Camera) SetProperty(id int, value float64) {
	return
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	sendMessage(c.camera, wmCapSetCallbackFrame, 0, 0)
	unregisterClass(c.className, c.instance)
	sendMessage(c.camera, wmCapDriverDisconnect, 0, 0)
	destroyWindow(c.camera)
	return
}

// callback function.
func (c *Camera) callback(hwvd syscall.Handle, hdr *videoHdr) uintptr {
	c.hdr = hdr
	return 0
}

var (
	user32   = syscall.NewLazyDLL("user32.dll")
	kernel32 = syscall.NewLazyDLL("kernel32.dll")
	avicap32 = syscall.NewLazyDLL("avicap32.dll")

	createWindowExW   = user32.NewProc("CreateWindowExW")
	destroyWindowW    = user32.NewProc("DestroyWindow")
	defWindowProcW    = user32.NewProc("DefWindowProcW")
	dispatchMessageW  = user32.NewProc("DispatchMessageW")
	translateMessageW = user32.NewProc("TranslateMessage")
	getMessageW       = user32.NewProc("GetMessageW")
	sendMessageW      = user32.NewProc("SendMessageW")
	postQuitMessageW  = user32.NewProc("PostQuitMessage")
	registerClassExW  = user32.NewProc("RegisterClassExW")
	unregisterClassW  = user32.NewProc("UnregisterClassW")

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
	wmCapGrabFrame        = wmCapStart + 60
	wmCapGrabFrameNoStop  = wmCapStart + 61
	wmCapStop             = wmCapStart + 68
	wmCapAbort            = wmCapStart + 69

	wsChild            = 0x40000000
	wsOverlappedWindow = 0x00CF0000

	cwUseDefault = 0x7fffffff
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
	ret, _, _ := defWindowProcW.Call(uintptr(hwnd), uintptr(msg), uintptr(wparam), uintptr(lparam))
	return uintptr(ret)
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-dispatchmessagew
func dispatchMessage(msg *msgW) {
	dispatchMessageW.Call(uintptr(unsafe.Pointer(msg)))
}

// https://docs.microsoft.com/en-us/windows/desktop/api/winuser/nf-winuser-translatemessage
func translateMessage(msg *msgW) {
	translateMessageW.Call(uintptr(unsafe.Pointer(msg)))
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

// messageLoop function
func messageLoop(hwnd syscall.Handle) {
	for {
		msg := &msgW{}
		ok, _ := getMessage(msg, 0, 0, 0)
		if ok {
			translateMessage(msg)
			dispatchMessage(msg)
		} else {
			break
		}
	}

	return
}
