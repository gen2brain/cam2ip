//go:build android

// Package camera.
package camera

/*
#include <android/log.h>

#include <media/NdkImageReader.h>

#include <camera/NdkCameraDevice.h>
#include <camera/NdkCameraManager.h>

#define TAG "camera"
#define LOGE(...) __android_log_print(ANDROID_LOG_ERROR, TAG, __VA_ARGS__)
#define LOGW(...) __android_log_print(ANDROID_LOG_WARN, TAG, __VA_ARGS__)
#define LOGI(...) __android_log_print(ANDROID_LOG_INFO, TAG, __VA_ARGS__)
#define LOGD(...) __android_log_print(ANDROID_LOG_DEBUG, TAG, __VA_ARGS__)

AImage *image;
AImageReader *imageReader;

ANativeWindow *nativeWindow;

ACameraDevice *cameraDevice;
ACameraManager *cameraManager;
ACameraOutputTarget *cameraOutputTarget;
ACameraCaptureSession *cameraCaptureSession;

ACaptureRequest *captureRequest;
ACaptureSessionOutput *captureSessionOutput;
ACaptureSessionOutputContainer *captureSessionOutputContainer;

void device_on_disconnected(void *context, ACameraDevice *device) {
    LOGI("camera %s is disconnected.\n", ACameraDevice_getId(device));
}

void device_on_error(void *context, ACameraDevice *device, int error) {
    LOGE("error %d on camera %s.\n", error, ACameraDevice_getId(device));
}

ACameraDevice_stateCallbacks deviceStateCallbacks = {
	.context = NULL,
	.onDisconnected = device_on_disconnected,
	.onError = device_on_error,
};

void session_on_ready(void *context, ACameraCaptureSession *session) {
    LOGI("session is ready. %p\n", session);
}

void session_on_active(void *context, ACameraCaptureSession *session) {
    LOGI("session is activated. %p\n", session);
}

void session_on_closed(void *context, ACameraCaptureSession *session) {
    LOGI("session is closed. %p\n", session);
}

ACameraCaptureSession_stateCallbacks captureSessionStateCallbacks = {
        .context = NULL,
        .onActive = session_on_active,
        .onReady = session_on_ready,
        .onClosed = session_on_closed,
};

void image_callback(void *context, AImageReader *reader) {
    LOGD("image_callback");

    media_status_t status = AImageReader_acquireLatestImage(reader, &image);
    if(status != AMEDIA_OK) {
		LOGE("failed to acquire next image (reason: %d).\n", status);
    }
}

AImageReader_ImageListener imageListener = {
	.context = NULL,
	.onImageAvailable = image_callback,
};

int openCamera(int index, int width, int height) {
    ACameraIdList *cameraIdList;
    const char *selectedCameraId;

    camera_status_t status = ACAMERA_OK;

    cameraManager = ACameraManager_create();

    status = ACameraManager_getCameraIdList(cameraManager, &cameraIdList);
    if(status != ACAMERA_OK) {
		LOGE("failed to get camera id list (reason: %d).\n", status);
		return status;
    }

    if(cameraIdList->numCameras < 1) {
		LOGE("no camera device detected.\n");
    }

    if(cameraIdList->numCameras < index+1) {
		LOGE("no camera at index %d.\n", index);
    }

    selectedCameraId = cameraIdList->cameraIds[index];
    LOGI("open camera (id: %s, num of cameras: %d).\n", selectedCameraId, cameraIdList->numCameras);

    status = ACameraManager_openCamera(cameraManager, selectedCameraId, &deviceStateCallbacks, &cameraDevice);
    if(status != ACAMERA_OK) {
		LOGE("failed to open camera device (id: %s)\n", selectedCameraId);
		return status;
    }

    status = ACameraDevice_createCaptureRequest(cameraDevice, TEMPLATE_STILL_CAPTURE, &captureRequest);
    if(status != ACAMERA_OK) {
		LOGE("failed to create snapshot capture request (id: %s)\n", selectedCameraId);
		return status;
    }

    status = ACaptureSessionOutputContainer_create(&captureSessionOutputContainer);
    if(status != ACAMERA_OK) {
		LOGE("failed to create session output container (id: %s)\n", selectedCameraId);
		return status;
    }

    media_status_t mstatus = AImageReader_new(width, height, AIMAGE_FORMAT_YUV_420_888, 2, &imageReader);
    if(mstatus != AMEDIA_OK) {
		LOGE("failed to create image reader (reason: %d).\n", mstatus);
		return mstatus;
    }

    mstatus = AImageReader_setImageListener(imageReader, &imageListener);
    if(mstatus != AMEDIA_OK) {
		LOGE("failed to set image listener (reason: %d).\n", mstatus);
		return mstatus;
    }

	AImageReader_getWindow(imageReader, &nativeWindow);
    ANativeWindow_acquire(nativeWindow);

    ACameraOutputTarget_create(nativeWindow, &cameraOutputTarget);
    ACaptureRequest_addTarget(captureRequest, cameraOutputTarget);

    ACaptureSessionOutput_create(nativeWindow, &captureSessionOutput);
	ACaptureSessionOutputContainer_add(captureSessionOutputContainer, captureSessionOutput);

    status = ACameraDevice_createCaptureSession(cameraDevice, captureSessionOutputContainer, &captureSessionStateCallbacks, &cameraCaptureSession);
    if(status != ACAMERA_OK) {
		LOGE("failed to create capture session (reason: %d).\n", status);
		return status;
    }

    ACameraManager_deleteCameraIdList(cameraIdList);
    ACameraManager_delete(cameraManager);

    return ACAMERA_OK;
}

int captureCamera() {
    camera_status_t status = ACameraCaptureSession_capture(cameraCaptureSession, NULL, 1, &captureRequest, NULL);
    if(status != ACAMERA_OK) {
		LOGE("failed to capture image (reason: %d).\n", status);
    }

    return status;
}

int closeCamera() {
    camera_status_t status = ACAMERA_OK;

    if(captureRequest != NULL) {
        ACaptureRequest_free(captureRequest);
        captureRequest = NULL;
    }

    if(cameraOutputTarget != NULL) {
        ACameraOutputTarget_free(cameraOutputTarget);
        cameraOutputTarget = NULL;
    }

    if(cameraDevice != NULL) {
        status = ACameraDevice_close(cameraDevice);

		if(status != ACAMERA_OK) {
			LOGE("failed to close camera device.\n");
			return status;
		}

		cameraDevice = NULL;
    }

    if(captureSessionOutput != NULL) {
        ACaptureSessionOutput_free(captureSessionOutput);
        captureSessionOutput = NULL;
    }

    if(captureSessionOutputContainer != NULL) {
        ACaptureSessionOutputContainer_free(captureSessionOutputContainer);
        captureSessionOutputContainer = NULL;
    }

    if(imageReader != NULL) {
		AImageReader_delete(imageReader);
		imageReader = NULL;
    }

    if(image != NULL) {
		AImage_delete(image);
		image = NULL;
	}

    LOGI("camera closed.\n");
    return ACAMERA_OK;
}

int openCamera(int index, int width, int height);
int captureCamera();
int closeCamera();

#cgo android CFLAGS: -D__ANDROID_API__=24
#cgo android LDFLAGS: -lcamera2ndk -lmediandk -llog -landroid
*/
import "C"

import (
	"fmt"
	"image"
	"unsafe"
)

// Camera represents camera.
type Camera struct {
	opts Options
	img  *image.YCbCr
}

// New returns new Camera for given camera index.
func New(opts Options) (camera *Camera, err error) {
	camera = &Camera{}
	camera.opts = opts

	camera.img = image.NewYCbCr(image.Rect(0, 0, int(opts.Width), int(opts.Height)), image.YCbCrSubsampleRatio420)

	ret := C.openCamera(C.int(opts.Index), C.int(opts.Width), C.int(opts.Height))
	if int(ret) != 0 {
		err = fmt.Errorf("camera: can not open camera %d: error %d", opts.Index, int(ret))

		return
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	ret := C.captureCamera()
	if int(ret) != 0 {
		err = fmt.Errorf("camera: can not grab frame: error %d", int(ret))

		return
	}

	if C.image == nil {
		err = fmt.Errorf("camera: can not retrieve frame")

		return
	}

	var yStride C.int
	var yLen, cbLen, crLen C.int
	var yPtr, cbPtr, crPtr *C.uint8_t

	C.AImage_getPlaneRowStride(C.image, 0, &yStride)
	C.AImage_getPlaneData(C.image, 0, &yPtr, &yLen)
	C.AImage_getPlaneData(C.image, 1, &cbPtr, &cbLen)
	C.AImage_getPlaneData(C.image, 2, &crPtr, &crLen)

	c.img.YStride = int(yStride)
	c.img.CStride = int(yStride) / 2

	c.img.Y = C.GoBytes(unsafe.Pointer(yPtr), yLen)
	c.img.Cb = C.GoBytes(unsafe.Pointer(cbPtr), cbLen)
	c.img.Cr = C.GoBytes(unsafe.Pointer(crPtr), crLen)

	img = c.img

	return
}

// GetProperty returns the specified camera property.
func (c *Camera) GetProperty(id int) float64 {
	return 0
}

// SetProperty sets a camera property.
func (c *Camera) SetProperty(id int, value float64) {
}

// Close closes camera.
func (c *Camera) Close() (err error) {
	ret := C.closeCamera()
	if int(ret) != 0 {
		err = fmt.Errorf("camera: can not close camera %d: error %d", c.opts.Index, int(ret))

		return
	}

	return
}
