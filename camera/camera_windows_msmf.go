//go:build windows && !vfw

// Package camera.
package camera

import (
	"fmt"
	"image"
	"runtime"
	"sync"
	"syscall"
	"unsafe"

	im "github.com/gen2brain/cam2ip/image"
)

type guid struct {
	Data1 uint32
	Data2 uint16
	Data3 uint16
	Data4 [8]byte
}

var (
	iidIMFMediaSource                    = guid{0x279a808d, 0xaec7, 0x40c8, [8]byte{0x9c, 0x6b, 0xa6, 0xb4, 0x92, 0xc7, 0x8a, 0x66}}
	mfDevsourceAttributeSourceType       = guid{0xc60ac5fe, 0x252a, 0x478f, [8]byte{0xa0, 0xef, 0xbc, 0x8f, 0xa5, 0xf7, 0xca, 0xd3}}
	mfDevsourceAttributeSourceTypeVidcap = guid{0x8ac3587a, 0x4ae7, 0x42d8, [8]byte{0x99, 0xe0, 0x0a, 0x60, 0x13, 0xee, 0xf9, 0x0f}}
	mfDevsourceAttributeFriendlyName     = guid{0x60d0e559, 0x52f8, 0x4fa2, [8]byte{0xbb, 0xce, 0xac, 0xdb, 0x34, 0xa8, 0xec, 0x01}}
	mfSourceReaderEnableVideoProcessing  = guid{0xfb394f3d, 0xccf1, 0x42ee, [8]byte{0xbb, 0xb3, 0xf9, 0xb8, 0x45, 0xd5, 0x68, 0x1d}}
	mfMTMajorType                        = guid{0x48eba18e, 0xf8c9, 0x4687, [8]byte{0xbf, 0x11, 0x0a, 0x74, 0xc9, 0xf9, 0x6a, 0x8f}}
	mfMTSubtype                          = guid{0xf7e34c9a, 0x42e8, 0x4714, [8]byte{0xb7, 0x4b, 0xcb, 0x29, 0xd7, 0x2c, 0x35, 0xe5}}
	mfMTFrameSize                        = guid{0x1652c33d, 0xd6b2, 0x4012, [8]byte{0xb8, 0x34, 0x72, 0x03, 0x08, 0x49, 0xa3, 0x7d}}
	mfMTDefaultStride                    = guid{0x644b4e48, 0x1e02, 0x4516, [8]byte{0xb0, 0xeb, 0xc0, 0x1c, 0xa9, 0xd4, 0x9a, 0xc6}}
	mfMediaTypeVideo                     = guid{0x73646976, 0x0000, 0x0010, [8]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}}
	mfVideoFormatNV12                    = guid{0x3231564e, 0x0000, 0x0010, [8]byte{0x80, 0x00, 0x00, 0xaa, 0x00, 0x38, 0x9b, 0x71}}
)

const (
	coinitMultithreaded = 0x0
	rpcEChangedMode     = 0x80010106

	mfVersion     = 0x00020070
	mfstartupLite = 1

	firstVideoStream = 0xFFFFFFFC
	allStreams       = 0xFFFFFFFE

	readerfEndOfStream            = 0x0002
	readerfCurrentMediaTypeChange = 0x0020
)

// IMFAttributes vtable indices, shared by IMFMediaType, IMFSample and IMFActivate.
const (
	idxGetUINT32 = 7
	idxGetUINT64 = 8
	idxSetUINT32 = 21
	idxSetUINT64 = 22
	idxSetGUID   = 24
)

var (
	mfplat      = syscall.NewLazyDLL("mfplat.dll")
	mfdll       = syscall.NewLazyDLL("mf.dll")
	mfreadwrite = syscall.NewLazyDLL("mfreadwrite.dll")
	ole32       = syscall.NewLazyDLL("ole32.dll")

	procCoInitializeEx                      = ole32.NewProc("CoInitializeEx")
	procCoUninitialize                      = ole32.NewProc("CoUninitialize")
	procCoTaskMemFree                       = ole32.NewProc("CoTaskMemFree")
	procMFStartup                           = mfplat.NewProc("MFStartup")
	procMFShutdown                          = mfplat.NewProc("MFShutdown")
	procMFCreateAttributes                  = mfplat.NewProc("MFCreateAttributes")
	procMFCreateMediaType                   = mfplat.NewProc("MFCreateMediaType")
	procMFEnumDeviceSources                 = mfdll.NewProc("MFEnumDeviceSources")
	procMFCreateSourceReaderFromMediaSource = mfreadwrite.NewProc("MFCreateSourceReaderFromMediaSource")
)

// Camera represents camera.
type Camera struct {
	opts Options

	source unsafe.Pointer
	reader unsafe.Pointer

	mu     sync.Mutex
	cond   *sync.Cond
	frame  []byte
	width  int
	height int
	stride int

	haveFrame bool
	closed    bool
	readErr   error

	ycbcr *image.YCbCr

	ready chan error
	done  chan struct{}
}

// New returns new Camera for given camera index.
func New(opts Options) (*Camera, error) {
	c := &Camera{opts: opts}
	c.cond = sync.NewCond(&c.mu)
	c.ready = make(chan error, 1)
	c.done = make(chan struct{})

	go c.run()

	if err := <-c.ready; err != nil {
		return nil, err
	}

	return c, nil
}

// run holds the COM apartment on a single OS thread for the camera's lifetime.
func (c *Camera) run() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	if err := c.setup(); err != nil {
		c.teardown()
		c.ready <- err
		close(c.done)

		return
	}

	c.ready <- nil

	for {
		c.mu.Lock()
		closed := c.closed
		c.mu.Unlock()

		if closed {
			break
		}

		if err := c.grab(); err != nil {
			c.fail(err)

			break
		}
	}

	c.teardown()
	close(c.done)
}

func (c *Camera) setup() error {
	for _, p := range []*syscall.LazyProc{
		procCoInitializeEx, procMFStartup, procMFCreateAttributes, procMFCreateMediaType,
		procMFEnumDeviceSources, procMFCreateSourceReaderFromMediaSource,
	} {
		if err := p.Find(); err != nil {
			return fmt.Errorf("camera: %s: %w", p.Name, err)
		}
	}

	if r, _, _ := procCoInitializeEx.Call(0, coinitMultithreaded); failed(r) && uint32(r) != rpcEChangedMode {
		return fmt.Errorf("camera: CoInitializeEx: %#x", r)
	}

	if r, _, _ := procMFStartup.Call(mfVersion, mfstartupLite); failed(r) {
		return fmt.Errorf("camera: MFStartup: %#x", r)
	}

	source, err := c.openSource()
	if err != nil {
		return err
	}
	c.source = source

	var readerAttr unsafe.Pointer
	if r, _, _ := procMFCreateAttributes.Call(uintptr(unsafe.Pointer(&readerAttr)), 1); failed(r) {
		return fmt.Errorf("camera: MFCreateAttributes: %#x", r)
	}
	attrSetUINT32(readerAttr, &mfSourceReaderEnableVideoProcessing, 1)

	var reader unsafe.Pointer
	r, _, _ := procMFCreateSourceReaderFromMediaSource.Call(uintptr(source), uintptr(readerAttr), uintptr(unsafe.Pointer(&reader)))
	release(readerAttr)
	if failed(r) {
		return fmt.Errorf("camera: MFCreateSourceReaderFromMediaSource: %#x", r)
	}
	c.reader = reader

	comCall(reader, 4, allStreams, 0)
	comCall(reader, 4, firstVideoStream, 1)

	return c.setMediaType()
}

// openSource enumerates video capture devices and activates the one at the configured index.
func (c *Camera) openSource() (unsafe.Pointer, error) {
	var attr unsafe.Pointer
	if r, _, _ := procMFCreateAttributes.Call(uintptr(unsafe.Pointer(&attr)), 1); failed(r) {
		return nil, fmt.Errorf("camera: MFCreateAttributes: %#x", r)
	}
	attrSetGUID(attr, &mfDevsourceAttributeSourceType, &mfDevsourceAttributeSourceTypeVidcap)

	var devices *unsafe.Pointer
	var count uint32
	r, _, _ := procMFEnumDeviceSources.Call(uintptr(attr), uintptr(unsafe.Pointer(&devices)), uintptr(unsafe.Pointer(&count)))
	release(attr)
	if failed(r) {
		return nil, fmt.Errorf("camera: MFEnumDeviceSources: %#x", r)
	}

	list := unsafe.Slice(devices, count)
	defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(devices)))

	if c.opts.Index < 0 || c.opts.Index >= int(count) {
		for _, a := range list {
			release(a)
		}

		return nil, fmt.Errorf("camera: no camera at index %d", c.opts.Index)
	}

	var source unsafe.Pointer
	r = comCall(list[c.opts.Index], 33, uintptr(unsafe.Pointer(&iidIMFMediaSource)), uintptr(unsafe.Pointer(&source)))

	for _, a := range list {
		release(a)
	}

	if failed(r) {
		return nil, fmt.Errorf("camera: ActivateObject: %#x", r)
	}

	return source, nil
}

// setMediaType requests NV12 at the configured size, falling back to the native size.
func (c *Camera) setMediaType() error {
	apply := func(withSize bool) uintptr {
		var mt unsafe.Pointer
		if r, _, _ := procMFCreateMediaType.Call(uintptr(unsafe.Pointer(&mt))); failed(r) {
			return r
		}

		attrSetGUID(mt, &mfMTMajorType, &mfMediaTypeVideo)
		attrSetGUID(mt, &mfMTSubtype, &mfVideoFormatNV12)
		if withSize {
			attrSetUINT64(mt, &mfMTFrameSize, uint64(c.opts.Width)<<32|uint64(c.opts.Height))
		}

		r := comCall(c.reader, 7, firstVideoStream, 0, uintptr(mt))
		release(mt)

		return r
	}

	if failed(apply(true)) {
		if r := apply(false); failed(r) {
			return fmt.Errorf("camera: SetCurrentMediaType: %#x", r)
		}
	}

	c.updateFormat()
	if c.width == 0 || c.height == 0 {
		c.width, c.height, c.stride = int(c.opts.Width), int(c.opts.Height), int(c.opts.Width)
	}

	c.ycbcr = image.NewYCbCr(image.Rect(0, 0, c.width, c.height), image.YCbCrSubsampleRatio420)

	return nil
}

// updateFormat reads the actual frame size and stride from the reader's current media type.
func (c *Camera) updateFormat() {
	var mt unsafe.Pointer
	if r := comCall(c.reader, 6, firstVideoStream, uintptr(unsafe.Pointer(&mt))); failed(r) {
		return
	}
	defer release(mt)

	var size uint64
	if r := comCall(mt, idxGetUINT64, uintptr(unsafe.Pointer(&mfMTFrameSize)), uintptr(unsafe.Pointer(&size))); !failed(r) {
		c.width = int(size >> 32)
		c.height = int(size & 0xffffffff)
	}

	var stride uint32
	if r := comCall(mt, idxGetUINT32, uintptr(unsafe.Pointer(&mfMTDefaultStride)), uintptr(unsafe.Pointer(&stride))); !failed(r) && stride != 0 {
		c.stride = abs(int(int32(stride)))
	} else {
		c.stride = c.width
	}
}

// grab reads one sample and stores the latest frame.
func (c *Camera) grab() error {
	var streamIndex, flags uint32
	var timestamp int64
	var sample unsafe.Pointer

	r := comCall(c.reader, 9, firstVideoStream, 0,
		uintptr(unsafe.Pointer(&streamIndex)),
		uintptr(unsafe.Pointer(&flags)),
		uintptr(unsafe.Pointer(&timestamp)),
		uintptr(unsafe.Pointer(&sample)))
	if failed(r) {
		return fmt.Errorf("camera: ReadSample: %#x", r)
	}

	if flags&readerfEndOfStream != 0 {
		return fmt.Errorf("camera: end of stream")
	}

	if flags&readerfCurrentMediaTypeChange != 0 {
		c.mu.Lock()
		c.updateFormat()
		c.ycbcr = image.NewYCbCr(image.Rect(0, 0, c.width, c.height), image.YCbCrSubsampleRatio420)
		c.mu.Unlock()
	}

	if sample == nil {
		return nil
	}
	defer release(sample)

	var buffer unsafe.Pointer
	if r := comCall(sample, 41, uintptr(unsafe.Pointer(&buffer))); failed(r) {
		return fmt.Errorf("camera: ConvertToContiguousBuffer: %#x", r)
	}
	defer release(buffer)

	var ptr *byte
	var maxLen, curLen uint32
	if r := comCall(buffer, 3, uintptr(unsafe.Pointer(&ptr)), uintptr(unsafe.Pointer(&maxLen)), uintptr(unsafe.Pointer(&curLen))); failed(r) {
		return fmt.Errorf("camera: Lock: %#x", r)
	}

	src := unsafe.Slice(ptr, curLen)

	c.mu.Lock()
	if len(c.frame) != int(curLen) {
		c.frame = make([]byte, curLen)
	}
	copy(c.frame, src)
	c.haveFrame = true
	c.cond.Signal()
	c.mu.Unlock()

	comCall(buffer, 4)

	return nil
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	c.mu.Lock()

	for !c.haveFrame && !c.closed {
		c.cond.Wait()
	}

	if c.closed {
		err = c.readErr
		c.mu.Unlock()

		if err == nil {
			err = fmt.Errorf("camera: closed")
		}

		return nil, err
	}

	nv12ToYCbCr420(c.frame, c.stride, c.width, c.height, c.ycbcr)
	img = c.ycbcr

	c.mu.Unlock()

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

	<-c.done

	return nil
}

func (c *Camera) fail(err error) {
	c.mu.Lock()
	c.closed = true
	c.readErr = err
	c.cond.Broadcast()
	c.mu.Unlock()
}

func (c *Camera) teardown() {
	if c.reader != nil {
		release(c.reader)
		c.reader = nil
	}

	if c.source != nil {
		comCall(c.source, 12)
		release(c.source)
		c.source = nil
	}

	procMFShutdown.Call()
	procCoUninitialize.Call()
}

// nv12ToYCbCr420 converts a stride-padded NV12 buffer to an image.YCbCr.
func nv12ToYCbCr420(data []byte, stride, width, height int, dst *image.YCbCr) {
	if stride < width || len(data) < stride*height+stride*(height/2) {
		return
	}

	for r := 0; r < height; r++ {
		copy(dst.Y[r*dst.YStride:r*dst.YStride+width], data[r*stride:r*stride+width])
	}

	uv := data[stride*height:]
	cw, ch := width/2, height/2

	for r := 0; r < ch; r++ {
		row := uv[r*stride:]
		for col := 0; col < cw; col++ {
			dst.Cb[r*dst.CStride+col] = row[col*2+0]
			dst.Cr[r*dst.CStride+col] = row[col*2+1]
		}
	}
}

// Info returns the negotiated capture format.
func (c *Camera) Info() Info {
	return Info{Format: "NV12", Width: c.width, Height: c.height}
}

// Devices returns the available capture devices.
func Devices() ([]DeviceInfo, error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	for _, p := range []*syscall.LazyProc{procCoInitializeEx, procMFStartup, procMFCreateAttributes, procMFEnumDeviceSources} {
		if err := p.Find(); err != nil {
			return nil, fmt.Errorf("camera: %s: %w", p.Name, err)
		}
	}

	if r, _, _ := procCoInitializeEx.Call(0, coinitMultithreaded); failed(r) && uint32(r) != rpcEChangedMode {
		return nil, fmt.Errorf("camera: CoInitializeEx: %#x", r)
	}

	if r, _, _ := procMFStartup.Call(mfVersion, mfstartupLite); failed(r) {
		return nil, fmt.Errorf("camera: MFStartup: %#x", r)
	}
	defer procMFShutdown.Call()
	defer procCoUninitialize.Call()

	var attr unsafe.Pointer
	if r, _, _ := procMFCreateAttributes.Call(uintptr(unsafe.Pointer(&attr)), 1); failed(r) {
		return nil, fmt.Errorf("camera: MFCreateAttributes: %#x", r)
	}
	attrSetGUID(attr, &mfDevsourceAttributeSourceType, &mfDevsourceAttributeSourceTypeVidcap)

	var list *unsafe.Pointer
	var count uint32
	r, _, _ := procMFEnumDeviceSources.Call(uintptr(attr), uintptr(unsafe.Pointer(&list)), uintptr(unsafe.Pointer(&count)))
	release(attr)
	if failed(r) {
		return nil, fmt.Errorf("camera: MFEnumDeviceSources: %#x", r)
	}
	defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(list)))

	activates := unsafe.Slice(list, count)
	devices := make([]DeviceInfo, 0, count)

	for i, a := range activates {
		devices = append(devices, DeviceInfo{Index: i, Name: friendlyName(a)})
		release(a)
	}

	return devices, nil
}

// friendlyName reads MF_DEVSOURCE_ATTRIBUTE_FRIENDLY_NAME from an activate object.
func friendlyName(activate unsafe.Pointer) string {
	var str *uint16
	var length uint32

	if r := comCall(activate, 13, uintptr(unsafe.Pointer(&mfDevsourceAttributeFriendlyName)), uintptr(unsafe.Pointer(&str)), uintptr(unsafe.Pointer(&length))); failed(r) {
		return ""
	}
	defer procCoTaskMemFree.Call(uintptr(unsafe.Pointer(str)))

	if str == nil {
		return ""
	}

	return syscall.UTF16ToString(unsafe.Slice(str, length))
}

// comCall invokes the method at the given vtable index on a COM object.
func comCall(obj unsafe.Pointer, index int, args ...uintptr) uintptr {
	vtbl := *(*unsafe.Pointer)(obj)
	method := *(*uintptr)(unsafe.Add(vtbl, uintptr(index)*unsafe.Sizeof(uintptr(0))))

	r, _, _ := syscall.SyscallN(method, append([]uintptr{uintptr(obj)}, args...)...)

	return r
}

func attrSetGUID(attr unsafe.Pointer, key, value *guid) {
	comCall(attr, idxSetGUID, uintptr(unsafe.Pointer(key)), uintptr(unsafe.Pointer(value)))
}

func attrSetUINT32(attr unsafe.Pointer, key *guid, value uint32) {
	comCall(attr, idxSetUINT32, uintptr(unsafe.Pointer(key)), uintptr(value))
}

func attrSetUINT64(attr unsafe.Pointer, key *guid, value uint64) {
	comCall(attr, idxSetUINT64, uintptr(unsafe.Pointer(key)), uintptr(value))
}

func release(obj unsafe.Pointer) {
	if obj != nil {
		comCall(obj, 2)
	}
}

func failed(hr uintptr) bool {
	return int32(hr) < 0
}

func abs(n int) int {
	if n < 0 {
		return -n
	}

	return n
}
