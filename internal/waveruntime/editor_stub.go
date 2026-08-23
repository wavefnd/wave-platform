//go:build !cgo || !linux

package waveruntime

import (
	"context"
	"fmt"

	editor "github.com/wavefnd/wave-platform/internal/editor"
)

type NativeEditor struct{}

func OpenEditor(path string) (*NativeEditor, error) {
	return nil, fmt.Errorf("%w: native Wave modules require Linux with cgo", ErrModuleUnavailable)
}

func (engine *NativeEditor) Transform(context.Context, editor.Request) (editor.Result, error) {
	return editor.Result{}, ErrModuleUnavailable
}

func (engine *NativeEditor) Close() error { return nil }
