// +build android

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

AImageReader *imageReader;

ACameraDevice *cameraDevice;
ACameraManager *cameraManager;
ACameraOutputTarget *cameraOutputTarget;
ACameraCaptureSession *cameraCaptureSession;

ACaptureRequest *captureRequest;
ACaptureSessionOutput *captureSessionOutput;
ACaptureSessionOutputContainer *captureSessionOutputContainer;

ACameraDevice_StateCallbacks deviceStateCallbacks;
ACameraCaptureSession_stateCallbacks captureSessionStateCallbacks;

uint8_t *image_data;
int image_len = 0;

void camera_device_on_disconnected(void *context, ACameraDevice *device) {
    LOGI("camera %s is diconnected.\n", ACameraDevice_getId(device));
}

void camera_device_on_error(void *context, ACameraDevice *device, int error) {
    LOGE("error %d on camera %s.\n", error, ACameraDevice_getId(device));
}

void capture_session_on_ready(void *context, ACameraCaptureSession *session) {
    LOGI("session is ready. %p\n", session);
}

void capture_session_on_active(void *context, ACameraCaptureSession *session) {
    LOGI("session is activated. %p\n", session);
}

void capture_session_on_closed(void *context, ACameraCaptureSession *session) {
    LOGI("session is closed. %p\n", session);
}

void imageCallback(void* context, AImageReader* reader) {
    LOGD("imageCallback");

    AImage *image;
    media_status_t status = AImageReader_acquireNextImage(reader, &image);
    if(status != ACAMERA_OK) {
		LOGE("failed to acquire next image (reason: %d).\n", status);
		return;
    }

    AImage_getPlaneData(image, 0, &image_data, &image_len);
    AImage_delete(image);
}

int openCamera(int index, int width, int height) {
    ACameraIdList *cameraIdList;
    const char *selectedCameraId;

    camera_status_t status = ACAMERA_OK;

    ACameraManager *cameraManager = ACameraManager_create();

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

    deviceStateCallbacks.onDisconnected = camera_device_on_disconnected;
    deviceStateCallbacks.onError = camera_device_on_error;

    status = ACameraManager_openCamera(cameraManager, selectedCameraId, &deviceStateCallbacks, &cameraDevice);
    if(status != ACAMERA_OK) {
		LOGE("failed to open camera device (id: %s)\n", selectedCameraId);
		return status;
    }

    status = ACameraDevice_createCaptureRequest(cameraDevice, TEMPLATE_VIDEO_SNAPSHOT, &captureRequest);
    if(status != ACAMERA_OK) {
		LOGE("failed to create snapshot capture request (id: %s)\n", selectedCameraId);
		return status;
    }

    ACaptureSessionOutputContainer_create(&captureSessionOutputContainer);

    captureSessionStateCallbacks.onReady = capture_session_on_ready;
    captureSessionStateCallbacks.onActive = capture_session_on_active;
    captureSessionStateCallbacks.onClosed = capture_session_on_closed;

    media_status_t mstatus = AImageReader_new(width, height, AIMAGE_FORMAT_RGBA_8888, 1, &imageReader);
    if(mstatus != ACAMERA_OK) {
		LOGE("failed to create image reader (reason: %d).\n", status);
		return mstatus;
    }

    AImageReader_ImageListener listener = {
		.context = NULL,
		.onImageAvailable = imageCallback,
    };

    AImageReader_setImageListener(imageReader, &listener);

    ANativeWindow *nativeWindow;
	AImageReader_getWindow(imageReader, &nativeWindow);
    ANativeWindow_acquire(nativeWindow);

    ACameraOutputTarget_create(nativeWindow, &cameraOutputTarget);
    ACaptureRequest_addTarget(captureRequest, cameraOutputTarget);
    ACaptureSessionOutput_create(nativeWindow, &captureSessionOutput);

    ACameraDevice_createCaptureSession(cameraDevice, captureSessionOutputContainer, &captureSessionStateCallbacks, &cameraCaptureSession);

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

void closeCamera() {
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

    LOGI("camera closed.\n");
}

int openCamera(int index, int width, int height);
int captureCamera();
void closeCamera();

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
	img  *image.RGBA
}

// New returns new Camera for given camera index.
func New(opts Options) (camera *Camera, err error) {
	camera = &Camera{}
	camera.opts = opts

	camera.img = image.NewRGBA(image.Rect(0, 0, int(opts.Width), int(opts.Height)))

	ret := C.openCamera(C.int(opts.Index), C.int(opts.Width), C.int(opts.Height))
	if bool(int(ret) != 0) {
		err = fmt.Errorf("camera: can not open camera %d: error %d", opts.Index, int(ret))
		return
	}

	return
}

// Read reads next frame from camera and returns image.
func (c *Camera) Read() (img image.Image, err error) {
	ret := C.captureCamera()
	if bool(int(ret) != 0) {
		err = fmt.Errorf("camera: can not grab frame: error %d", int(ret))
		return
	}

	if C.image_data != nil {
		c.img.Pix = (*[1 << 24]uint8)(unsafe.Pointer(&C.image_data))[0:int(C.image_len)]
		img = c.img
	}

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
	C.closeCamera()

	return
}
