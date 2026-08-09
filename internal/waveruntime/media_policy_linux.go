//go:build cgo && linux

package waveruntime

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdint.h>
#include <stdlib.h>

typedef int32_t (*wave_media_policy_abi_version_fn)(void);
typedef int32_t (*wave_media_policy_plan_fn)(
    int64_t, int64_t, int64_t, int32_t, int64_t *, int64_t *
);

static void *wave_media_policy_open(const char *path, int32_t expected_abi) {
    void *handle = dlopen(path, RTLD_NOW | RTLD_LOCAL);
    if (handle == NULL) {
        return NULL;
    }
    wave_media_policy_abi_version_fn abi =
        (wave_media_policy_abi_version_fn)dlsym(handle, "wave_media_policy_abi_version");
    wave_media_policy_plan_fn plan =
        (wave_media_policy_plan_fn)dlsym(handle, "wave_media_policy_plan");
    if (abi == NULL || plan == NULL || abi() != expected_abi) {
        dlclose(handle);
        return NULL;
    }
    return handle;
}

static int32_t wave_media_policy_call(
    void *handle,
    int64_t width,
    int64_t height,
    int64_t input_bytes,
    int32_t format,
    int64_t *out_width,
    int64_t *out_height
) {
    wave_media_policy_plan_fn plan =
        (wave_media_policy_plan_fn)dlsym(handle, "wave_media_policy_plan");
    if (plan == NULL) {
        return -99;
    }
    return plan(width, height, input_bytes, format, out_width, out_height);
}
*/
import "C"

import (
	"context"
	"fmt"
	"sync"
	"unsafe"

	"github.com/wavefnd/wave-platform/internal/mediapolicy"
)

type NativeMediaPolicy struct {
	mu     sync.RWMutex
	handle unsafe.Pointer
}

func OpenMediaPolicy(path string) (*NativeMediaPolicy, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	handle := C.wave_media_policy_open(cPath, C.int32_t(mediaPolicyABI))
	if handle == nil {
		return nil, fmt.Errorf("%w: open %s", ErrModuleUnavailable, path)
	}
	return &NativeMediaPolicy{handle: handle}, nil
}

func (policy *NativeMediaPolicy) Plan(ctx context.Context, width, height int, inputBytes int64, format mediapolicy.Format) (mediapolicy.Plan, error) {
	if err := ctx.Err(); err != nil {
		return mediapolicy.Plan{}, err
	}
	policy.mu.RLock()
	defer policy.mu.RUnlock()
	if policy.handle == nil {
		return mediapolicy.Plan{}, mediapolicy.ErrUnavailable
	}

	var outputWidth, outputHeight C.int64_t
	status := C.wave_media_policy_call(
		policy.handle,
		C.int64_t(width),
		C.int64_t(height),
		C.int64_t(inputBytes),
		C.int32_t(format),
		&outputWidth,
		&outputHeight,
	)
	if status != 0 {
		return mediapolicy.Plan{}, mediaPolicyError(int(status))
	}
	return mediapolicy.Plan{Width: int(outputWidth), Height: int(outputHeight)}, nil
}

func mediaPolicyError(status int) error {
	switch status {
	case -1:
		return mediapolicy.ErrInvalidDimensions
	case -2:
		return mediapolicy.ErrInputTooLarge
	case -3:
		return mediapolicy.ErrUnsupportedFormat
	case -4:
		return mediapolicy.ErrPixelLimit
	default:
		return fmt.Errorf("Wave media policy returned status %d", status)
	}
}

func (policy *NativeMediaPolicy) Close() error {
	policy.mu.Lock()
	defer policy.mu.Unlock()
	if policy.handle != nil {
		C.dlclose(policy.handle)
		policy.handle = nil
	}
	return nil
}
