//go:build darwin

// Package camera.
package camera

import (
	"fmt"
	"image"
	"sync"
	"unsafe"

	"github.com/ebitengine/purego"
	"github.com/ebitengine/purego/objc"

	im "github.com/gen2brain/cam2ip/image"
)

// Camera represents camera.
type Camera struct {
	opts     Options
	session  objc.ID
	delegate objc.ID

	mu        sync.Mutex
	cond      *sync.Cond
	frame     []byte
	width     int
	height    int
	haveFrame bool
	closed    bool

	rgba *image.RGBA
}

// New returns new Camera for given camera index.
func New(opts Options) (c *Camera, err error) {
	if err = loadFrameworks(); err != nil {
		return nil, err
	}

	c = &Camera{opts: opts}
	c.cond = sync.NewCond(&c.mu)

	pool := objc.ID(objc.GetClass("NSAutoreleasePool")).Send(selAlloc).Send(selInit)
	defer pool.Send(selDrain)

	avCaptureDevice := objc.ID(objc.GetClass("AVCaptureDevice"))

	status := objc.Send[int](avCaptureDevice, selAuthStatusForMediaType, avMediaTypeVideo)
	if status == authNotDetermined {
		status = requestAccess(avCaptureDevice)
	}

	if status != authAuthorized {
		return nil, fmt.Errorf("camera: camera access denied, grant it in System Settings > Privacy & Security > Camera")
	}

	devices := avCaptureDevice.Send(selDevicesWithMediaType, avMediaTypeVideo)
	count := int(objc.Send[uint64](devices, selCount))
	if opts.Index < 0 || opts.Index >= count {
		return nil, fmt.Errorf("camera: no camera at index %d", opts.Index)
	}

	device := devices.Send(selObjectAtIndex, uint64(opts.Index))

	var nserr objc.ID
	input := objc.ID(objc.GetClass("AVCaptureDeviceInput")).Send(selDeviceInputWithDevice, device, &nserr)
	if input == 0 {
		return nil, fmt.Errorf("camera: cannot create device input")
	}

	output := objc.ID(objc.GetClass("AVCaptureVideoDataOutput")).Send(selAlloc).Send(selInit)
	output.Send(selSetVideoSettings, c.videoSettings())

	c.delegate = objc.ID(delegateClass).Send(selNew)

	registryMu.Lock()
	registry[c.delegate] = c
	registryMu.Unlock()

	queue := dispatchQueueCreate(&queueLabel[0], 0)
	output.Send(selSetSampleBufferQueue, c.delegate, queue)

	c.session = objc.ID(objc.GetClass("AVCaptureSession")).Send(selAlloc).Send(selInit)

	if objc.Send[bool](c.session, selCanAddInput, input) {
		c.session.Send(selAddInput, input)
	} else {
		return nil, fmt.Errorf("camera: cannot add input")
	}

	if objc.Send[bool](c.session, selCanAddOutput, output) {
		c.session.Send(selAddOutput, output)
	} else {
		return nil, fmt.Errorf("camera: cannot add output")
	}

	c.session.Send(selStartRunning)

	return c, nil
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	c.mu.Lock()

	for !c.haveFrame && !c.closed {
		c.cond.Wait()
	}

	if c.closed {
		c.mu.Unlock()

		return nil, fmt.Errorf("camera: closed")
	}

	if c.rgba == nil || c.rgba.Bounds().Dx() != c.width || c.rgba.Bounds().Dy() != c.height {
		c.rgba = image.NewRGBA(image.Rect(0, 0, c.width, c.height))
	}

	bgraToRgba(c.frame, c.rgba)

	c.mu.Unlock()

	img = c.rgba

	if c.opts.Rotate != 0 {
		img = im.Rotate(img, c.opts.Rotate)
	}

	if c.opts.Flip != "" {
		img = im.Flip(img, c.opts.Flip)
	}

	if c.opts.Timestamp {
		img = im.Timestamp(img, c.opts.TimeFormat)
	}

	return img, nil
}

// Close closes camera.
func (c *Camera) Close() error {
	c.mu.Lock()
	c.closed = true
	c.cond.Broadcast()
	c.mu.Unlock()

	if c.session != 0 {
		c.session.Send(selStopRunning)
	}

	registryMu.Lock()
	delete(registry, c.delegate)
	registryMu.Unlock()

	return nil
}

// videoSettings builds the NSDictionary requesting BGRA at the configured size.
func (c *Camera) videoSettings() objc.ID {
	dict := objc.ID(objc.GetClass("NSMutableDictionary")).Send(selNew)

	number := objc.ID(objc.GetClass("NSNumber"))

	dict.Send(selSetObjectForKey, number.Send(selNumberWithInt, int32(pixelFormat32BGRA)), keyPixelFormatType)
	dict.Send(selSetObjectForKey, number.Send(selNumberWithInt, int32(c.opts.Width)), keyWidth)
	dict.Send(selSetObjectForKey, number.Send(selNumberWithInt, int32(c.opts.Height)), keyHeight)

	return dict
}

// onSampleBuffer copies the latest frame out of the pixel buffer; runs on the capture queue.
func (c *Camera) onSampleBuffer(sampleBuffer uintptr) {
	pixels := cmSampleBufferGetImageBuffer(sampleBuffer)
	if pixels == 0 {
		return
	}

	cvPixelBufferLockBaseAddress(pixels, lockReadOnly)
	defer cvPixelBufferUnlockBaseAddress(pixels, lockReadOnly)

	if cvPixelBufferGetPixelFormatType(pixels) != pixelFormat32BGRA {
		return
	}

	base := cvPixelBufferGetBaseAddress(pixels)
	if base == nil {
		return
	}

	width := int(cvPixelBufferGetWidth(pixels))
	height := int(cvPixelBufferGetHeight(pixels))
	rowBytes := int(cvPixelBufferGetBytesPerRow(pixels))

	src := unsafe.Slice((*byte)(base), rowBytes*height)

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return
	}

	stride := width * 4
	if len(c.frame) != stride*height {
		c.frame = make([]byte, stride*height)
	}

	for y := 0; y < height; y++ {
		copy(c.frame[y*stride:y*stride+stride], src[y*rowBytes:y*rowBytes+stride])
	}

	c.width = width
	c.height = height
	c.haveFrame = true

	c.cond.Signal()
}

func loadFrameworks() error {
	loadOnce.Do(func() {
		open := func(path string) uintptr {
			if loadErr != nil {
				return 0
			}

			handle, err := purego.Dlopen(path, purego.RTLD_GLOBAL|purego.RTLD_LAZY)
			if err != nil {
				loadErr = fmt.Errorf("camera: cannot load %s: %w", path, err)
			}

			return handle
		}

		open("/System/Library/Frameworks/Foundation.framework/Foundation")
		coreMedia := open("/System/Library/Frameworks/CoreMedia.framework/CoreMedia")
		coreVideo := open("/System/Library/Frameworks/CoreVideo.framework/CoreVideo")
		open("/System/Library/Frameworks/AVFoundation.framework/AVFoundation")
		libSystem := open("/usr/lib/libSystem.B.dylib")
		if loadErr != nil {
			return
		}

		purego.RegisterLibFunc(&cmSampleBufferGetImageBuffer, coreMedia, "CMSampleBufferGetImageBuffer")
		purego.RegisterLibFunc(&cvPixelBufferLockBaseAddress, coreVideo, "CVPixelBufferLockBaseAddress")
		purego.RegisterLibFunc(&cvPixelBufferUnlockBaseAddress, coreVideo, "CVPixelBufferUnlockBaseAddress")
		purego.RegisterLibFunc(&cvPixelBufferGetBaseAddress, coreVideo, "CVPixelBufferGetBaseAddress")
		purego.RegisterLibFunc(&cvPixelBufferGetWidth, coreVideo, "CVPixelBufferGetWidth")
		purego.RegisterLibFunc(&cvPixelBufferGetHeight, coreVideo, "CVPixelBufferGetHeight")
		purego.RegisterLibFunc(&cvPixelBufferGetBytesPerRow, coreVideo, "CVPixelBufferGetBytesPerRow")
		purego.RegisterLibFunc(&cvPixelBufferGetPixelFormatType, coreVideo, "CVPixelBufferGetPixelFormatType")
		purego.RegisterLibFunc(&dispatchQueueCreate, libSystem, "dispatch_queue_create")

		// String values of the AVFoundation/CoreVideo constants; matched by content.
		avMediaTypeVideo = nsString("vide")
		keyPixelFormatType = nsString("PixelFormatType")
		keyWidth = nsString("Width")
		keyHeight = nsString("Height")

		loadErr = registerDelegate()
	})

	return loadErr
}

// nsString returns an NSString with the given contents.
func nsString(s string) objc.ID {
	return objc.ID(objc.GetClass("NSString")).Send(selAlloc).Send(selInitWithUTF8, s+"\x00")
}

// requestAccess prompts for camera access and blocks until the user responds.
func requestAccess(device objc.ID) int {
	done := make(chan bool, 1)

	block := objc.NewBlock(func(_ objc.Block, granted bool) {
		done <- granted
	})

	device.Send(selRequestAccess, avMediaTypeVideo, block)

	if <-done {
		return authAuthorized
	}

	return authDenied
}

func registerDelegate() error {
	class, err := objc.RegisterClass("cam2ipCaptureDelegate", objc.GetClass("NSObject"), nil, nil, []objc.MethodDef{
		{
			Cmd: objc.RegisterName("captureOutput:didOutputSampleBuffer:fromConnection:"),
			Fn: func(self objc.ID, _ objc.SEL, _ objc.ID, sampleBuffer uintptr, _ objc.ID) {
				registryMu.Lock()
				c := registry[self]
				registryMu.Unlock()

				if c != nil {
					c.onSampleBuffer(sampleBuffer)
				}
			},
		},
	})
	if err != nil {
		return fmt.Errorf("camera: cannot register delegate: %w", err)
	}

	delegateClass = class

	return nil
}

// kCVPixelFormatType_32BGRA, the format requested from the capture output.
const pixelFormat32BGRA = 0x42475241

// kCVPixelBufferLock_ReadOnly.
const lockReadOnly = 0x00000001

// AVAuthorizationStatus values.
const (
	authNotDetermined = 0
	authRestricted    = 1
	authDenied        = 2
	authAuthorized    = 3
)

var (
	cmSampleBufferGetImageBuffer    func(uintptr) uintptr
	cvPixelBufferLockBaseAddress    func(uintptr, uint64) int32
	cvPixelBufferUnlockBaseAddress  func(uintptr, uint64) int32
	cvPixelBufferGetBaseAddress     func(uintptr) unsafe.Pointer
	cvPixelBufferGetWidth           func(uintptr) uint64
	cvPixelBufferGetHeight          func(uintptr) uint64
	cvPixelBufferGetBytesPerRow     func(uintptr) uint64
	cvPixelBufferGetPixelFormatType func(uintptr) uint32
	dispatchQueueCreate             func(*byte, uintptr) uintptr
)

var (
	avMediaTypeVideo objc.ID

	keyPixelFormatType objc.ID
	keyWidth           objc.ID
	keyHeight          objc.ID
)

var (
	selAlloc                  = objc.RegisterName("alloc")
	selInit                   = objc.RegisterName("init")
	selInitWithUTF8           = objc.RegisterName("initWithUTF8String:")
	selNew                    = objc.RegisterName("new")
	selDrain                  = objc.RegisterName("drain")
	selCount                  = objc.RegisterName("count")
	selObjectAtIndex          = objc.RegisterName("objectAtIndex:")
	selNumberWithInt          = objc.RegisterName("numberWithInt:")
	selSetObjectForKey        = objc.RegisterName("setObject:forKey:")
	selDevicesWithMediaType   = objc.RegisterName("devicesWithMediaType:")
	selAuthStatusForMediaType = objc.RegisterName("authorizationStatusForMediaType:")
	selRequestAccess          = objc.RegisterName("requestAccessForMediaType:completionHandler:")
	selDeviceInputWithDevice  = objc.RegisterName("deviceInputWithDevice:error:")
	selCanAddInput            = objc.RegisterName("canAddInput:")
	selAddInput               = objc.RegisterName("addInput:")
	selCanAddOutput           = objc.RegisterName("canAddOutput:")
	selAddOutput              = objc.RegisterName("addOutput:")
	selSetVideoSettings       = objc.RegisterName("setVideoSettings:")
	selSetSampleBufferQueue   = objc.RegisterName("setSampleBufferDelegate:queue:")
	selStartRunning           = objc.RegisterName("startRunning")
	selStopRunning            = objc.RegisterName("stopRunning")
)

var (
	loadOnce sync.Once
	loadErr  error

	delegateClass objc.Class

	registryMu sync.Mutex
	registry   = map[objc.ID]*Camera{}

	queueLabel = []byte("cam2ip\x00")
)

// bgraToRgba converts a packed BGRA byte slice to an image.RGBA.
func bgraToRgba(src []byte, dst *image.RGBA) {
	for i := 0; i+3 < len(src) && i+3 < len(dst.Pix); i += 4 {
		dst.Pix[i+0] = src[i+2]
		dst.Pix[i+1] = src[i+1]
		dst.Pix[i+2] = src[i+0]
		dst.Pix[i+3] = 0xFF
	}
}
