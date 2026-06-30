//go:build android

// Package camera.
package camera

/*
#include <stdlib.h>

#include <android/log.h>

#include <media/NdkImageReader.h>

#include <camera/NdkCameraDevice.h>
#include <camera/NdkCameraManager.h>
#include <camera/NdkCameraMetadata.h>

#define TAG "cam2ip"
#define LOGE(...) __android_log_print(ANDROID_LOG_ERROR, TAG, __VA_ARGS__)
#define LOGI(...) __android_log_print(ANDROID_LOG_INFO, TAG, __VA_ARGS__)

AImage *image;
static AImageReader *imageReader;
static ANativeWindow *nativeWindow;

static ACameraDevice *cameraDevice;
static ACameraManager *cameraManager;
static ACameraOutputTarget *cameraOutputTarget;
static ACameraCaptureSession *cameraCaptureSession;

static ACaptureRequest *captureRequest;
static ACaptureSessionOutput *captureSessionOutput;
static ACaptureSessionOutputContainer *captureSessionOutputContainer;

static void device_on_disconnected(void *context, ACameraDevice *device) {
    LOGI("camera %s disconnected", ACameraDevice_getId(device));
}

static void device_on_error(void *context, ACameraDevice *device, int error) {
    LOGE("camera %s error %d", ACameraDevice_getId(device), error);
}

static ACameraDevice_stateCallbacks deviceStateCallbacks = {
    .onDisconnected = device_on_disconnected,
    .onError = device_on_error,
};

static ACameraCaptureSession_stateCallbacks captureSessionStateCallbacks = {0};

int openCamera(int index, int width, int height) {
    ACameraIdList *cameraIdList = NULL;

    cameraManager = ACameraManager_create();

    camera_status_t status = ACameraManager_getCameraIdList(cameraManager, &cameraIdList);
    if (status != ACAMERA_OK) {
        return status;
    }

    if (index < 0 || index >= cameraIdList->numCameras) {
        ACameraManager_deleteCameraIdList(cameraIdList);
        return -1;
    }

    const char *cameraId = cameraIdList->cameraIds[index];
    LOGI("open camera %s (%d available)", cameraId, cameraIdList->numCameras);

    status = ACameraManager_openCamera(cameraManager, cameraId, &deviceStateCallbacks, &cameraDevice);
    ACameraManager_deleteCameraIdList(cameraIdList);
    if (status != ACAMERA_OK) {
        return status;
    }

    media_status_t mstatus = AImageReader_new(width, height, AIMAGE_FORMAT_YUV_420_888, 4, &imageReader);
    if (mstatus != AMEDIA_OK) {
        return mstatus;
    }

    AImageReader_getWindow(imageReader, &nativeWindow);
    ANativeWindow_acquire(nativeWindow);

    status = ACameraDevice_createCaptureRequest(cameraDevice, TEMPLATE_PREVIEW, &captureRequest);
    if (status != ACAMERA_OK) {
        return status;
    }

    // Auto exposure and a frame rate range so the sensor streams continuously.
    uint8_t controlMode = ACAMERA_CONTROL_MODE_AUTO;
    ACaptureRequest_setEntry_u8(captureRequest, ACAMERA_CONTROL_MODE, 1, &controlMode);
    uint8_t aeMode = ACAMERA_CONTROL_AE_MODE_ON;
    ACaptureRequest_setEntry_u8(captureRequest, ACAMERA_CONTROL_AE_MODE, 1, &aeMode);
    int32_t fpsRange[2] = {15, 30};
    ACaptureRequest_setEntry_i32(captureRequest, ACAMERA_CONTROL_AE_TARGET_FPS_RANGE, 2, fpsRange);

    ACameraOutputTarget_create(nativeWindow, &cameraOutputTarget);
    ACaptureRequest_addTarget(captureRequest, cameraOutputTarget);

    ACaptureSessionOutputContainer_create(&captureSessionOutputContainer);
    ACaptureSessionOutput_create(nativeWindow, &captureSessionOutput);
    ACaptureSessionOutputContainer_add(captureSessionOutputContainer, captureSessionOutput);

    status = ACameraDevice_createCaptureSession(cameraDevice, captureSessionOutputContainer, &captureSessionStateCallbacks, &cameraCaptureSession);
    if (status != ACAMERA_OK) {
        return status;
    }

    status = ACameraCaptureSession_setRepeatingRequest(cameraCaptureSession, NULL, 1, &captureRequest, NULL);
    if (status != ACAMERA_OK) {
        return status;
    }

    return ACAMERA_OK;
}

// acquireImage releases the previous frame and acquires the most recent one.
int acquireImage() {
    if (image != NULL) {
        AImage_delete(image);
        image = NULL;
    }

    return AImageReader_acquireLatestImage(imageReader, &image);
}

int closeCamera() {
    if (cameraCaptureSession != NULL) {
        ACameraCaptureSession_stopRepeating(cameraCaptureSession);
        ACameraCaptureSession_close(cameraCaptureSession);
        cameraCaptureSession = NULL;
    }

    if (captureRequest != NULL) {
        ACaptureRequest_free(captureRequest);
        captureRequest = NULL;
    }

    if (cameraOutputTarget != NULL) {
        ACameraOutputTarget_free(cameraOutputTarget);
        cameraOutputTarget = NULL;
    }

    if (cameraDevice != NULL) {
        ACameraDevice_close(cameraDevice);
        cameraDevice = NULL;
    }

    if (captureSessionOutput != NULL) {
        ACaptureSessionOutput_free(captureSessionOutput);
        captureSessionOutput = NULL;
    }

    if (captureSessionOutputContainer != NULL) {
        ACaptureSessionOutputContainer_free(captureSessionOutputContainer);
        captureSessionOutputContainer = NULL;
    }

    if (nativeWindow != NULL) {
        ANativeWindow_release(nativeWindow);
        nativeWindow = NULL;
    }

    if (imageReader != NULL) {
        AImageReader_delete(imageReader);
        imageReader = NULL;
    }

    if (image != NULL) {
        AImage_delete(image);
        image = NULL;
    }

    if (cameraManager != NULL) {
        ACameraManager_delete(cameraManager);
        cameraManager = NULL;
    }

    return ACAMERA_OK;
}

static ACameraManager *enumManager;
static ACameraIdList *enumList;

int camerasOpen() {
    enumManager = ACameraManager_create();
    if (ACameraManager_getCameraIdList(enumManager, &enumList) != ACAMERA_OK) {
        return -1;
    }

    return enumList->numCameras;
}

const char *cameraIdAt(int i) {
    return enumList->cameraIds[i];
}

int cameraFacingAt(int i) {
    ACameraMetadata *meta;
    if (ACameraManager_getCameraCharacteristics(enumManager, enumList->cameraIds[i], &meta) != ACAMERA_OK) {
        return -1;
    }

    int facing = -1;
    ACameraMetadata_const_entry entry;
    if (ACameraMetadata_getConstEntry(meta, ACAMERA_LENS_FACING, &entry) == ACAMERA_OK) {
        facing = entry.data.u8[0];
    }

    ACameraMetadata_free(meta);

    return facing;
}

void camerasClose() {
    if (enumList != NULL) {
        ACameraManager_deleteCameraIdList(enumList);
        enumList = NULL;
    }

    if (enumManager != NULL) {
        ACameraManager_delete(enumManager);
        enumManager = NULL;
    }
}

#cgo android LDFLAGS: -lcamera2ndk -lmediandk -llog -landroid
*/
import "C"

import (
	"fmt"
	"image"
	"time"
	"unsafe"

	im "github.com/gen2brain/cam2ip/image"
)

// Camera represents camera.
type Camera struct {
	opts   Options
	width  int
	height int
	ycbcr  *image.YCbCr
}

// New returns new Camera for given camera index.
func New(opts Options) (*Camera, error) {
	c := &Camera{opts: opts}
	c.width = int(opts.Width)
	c.height = int(opts.Height)
	c.ycbcr = image.NewYCbCr(image.Rect(0, 0, c.width, c.height), image.YCbCrSubsampleRatio420)

	ret := C.openCamera(C.int(opts.Index), C.int(c.width), C.int(c.height))
	if int(ret) != 0 {
		return nil, fmt.Errorf("camera: can not open camera %d: error %d", opts.Index, int(ret))
	}

	return c, nil
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	for i := 0; ; i++ {
		if int(C.acquireImage()) == 0 && C.image != nil {
			break
		}

		if i == 99 {
			return nil, fmt.Errorf("camera: can not retrieve frame")
		}

		time.Sleep(10 * time.Millisecond)
	}

	c.convert()

	img = c.ycbcr

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

// convert copies the YUV_420_888 planes into the YCbCr image, handling row and pixel strides (planar or semi-planar).
func (c *Camera) convert() {
	var yStride, uStride, vStride C.int32_t
	var uPixel, vPixel C.int32_t
	var yPtr, uPtr, vPtr *C.uint8_t
	var yLen, uLen, vLen C.int

	C.AImage_getPlaneRowStride(C.image, 0, &yStride)
	C.AImage_getPlaneRowStride(C.image, 1, &uStride)
	C.AImage_getPlaneRowStride(C.image, 2, &vStride)
	C.AImage_getPlanePixelStride(C.image, 1, &uPixel)
	C.AImage_getPlanePixelStride(C.image, 2, &vPixel)
	C.AImage_getPlaneData(C.image, 0, &yPtr, &yLen)
	C.AImage_getPlaneData(C.image, 1, &uPtr, &uLen)
	C.AImage_getPlaneData(C.image, 2, &vPtr, &vLen)

	ySrc := unsafe.Slice((*byte)(unsafe.Pointer(yPtr)), int(yLen))
	uSrc := unsafe.Slice((*byte)(unsafe.Pointer(uPtr)), int(uLen))
	vSrc := unsafe.Slice((*byte)(unsafe.Pointer(vPtr)), int(vLen))

	ys := int(yStride)
	for r := 0; r < c.height; r++ {
		copy(c.ycbcr.Y[r*c.ycbcr.YStride:r*c.ycbcr.YStride+c.width], ySrc[r*ys:r*ys+c.width])
	}

	cw, ch := c.width/2, c.height/2
	us, up := int(uStride), int(uPixel)
	vs, vp := int(vStride), int(vPixel)

	for r := 0; r < ch; r++ {
		for col := 0; col < cw; col++ {
			c.ycbcr.Cb[r*c.ycbcr.CStride+col] = uSrc[r*us+col*up]
			c.ycbcr.Cr[r*c.ycbcr.CStride+col] = vSrc[r*vs+col*vp]
		}
	}
}

// Close closes camera.
func (c *Camera) Close() error {
	if int(C.closeCamera()) != 0 {
		return fmt.Errorf("camera: can not close camera %d", c.opts.Index)
	}

	return nil
}

// Info returns the negotiated capture format.
func (c *Camera) Info() Info {
	return Info{Format: "YUV420", Width: c.width, Height: c.height}
}

// Devices returns the available capture devices.
func Devices() ([]DeviceInfo, error) {
	count := int(C.camerasOpen())
	if count < 0 {
		C.camerasClose()

		return nil, fmt.Errorf("camera: can not list cameras")
	}
	defer C.camerasClose()

	devices := make([]DeviceInfo, 0, count)
	for i := 0; i < count; i++ {
		id := C.GoString(C.cameraIdAt(C.int(i)))

		name := id
		switch int(C.cameraFacingAt(C.int(i))) {
		case 0:
			name = id + " (front)"
		case 1:
			name = id + " (back)"
		case 2:
			name = id + " (external)"
		}

		devices = append(devices, DeviceInfo{Index: i, Name: name})
	}

	return devices, nil
}
