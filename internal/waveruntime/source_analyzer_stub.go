//go:build !cgo || !linux

package waveruntime

import (
	"context"
	"fmt"

	"github.com/wavefnd/wave-platform/internal/sourceanalysis"
)

type NativeSourceAnalyzer struct{}

func OpenSourceAnalyzer(path string) (*NativeSourceAnalyzer, error) {
	return nil, fmt.Errorf("%w: native Wave modules require Linux with cgo", ErrModuleUnavailable)
}

func (analyzer *NativeSourceAnalyzer) Analyze(context.Context, []byte) (sourceanalysis.Analysis, error) {
	return sourceanalysis.Analysis{}, ErrModuleUnavailable
}

func (analyzer *NativeSourceAnalyzer) Close() error { return nil }
