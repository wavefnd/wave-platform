//go:build !cgo || !linux

package waveruntime

import (
	"context"
	"fmt"

	"github.com/wavefnd/wave-platform/internal/mediapolicy"
)

type NativeMediaPolicy struct{}

func OpenMediaPolicy(path string) (*NativeMediaPolicy, error) {
	return nil, fmt.Errorf("%w: native Wave modules require Linux with cgo", ErrModuleUnavailable)
}

func (policy *NativeMediaPolicy) Plan(context.Context, int, int, int64, mediapolicy.Format) (mediapolicy.Plan, error) {
	return mediapolicy.Plan{}, mediapolicy.ErrUnavailable
}

func (policy *NativeMediaPolicy) Close() error { return nil }
