//go:build cgo && linux

package waveruntime

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

typedef int32_t (*wave_source_analyzer_abi_version_fn)(void);
typedef int32_t (*wave_source_analyzer_highlight_fn)(
    const uint8_t *, int64_t, uint8_t *, int64_t
);

static void *wave_source_analyzer_open(const char *path, int32_t expected_abi) {
    void *handle = dlopen(path, RTLD_NOW | RTLD_LOCAL);
    if (handle == NULL) {
        return NULL;
    }
    wave_source_analyzer_abi_version_fn abi =
        (wave_source_analyzer_abi_version_fn)dlsym(handle, "wave_source_analyzer_abi_version");
    wave_source_analyzer_highlight_fn highlight =
        (wave_source_analyzer_highlight_fn)dlsym(handle, "wave_source_analyzer_highlight");
    if (abi == NULL || highlight == NULL || abi() != expected_abi) {
        dlclose(handle);
        return NULL;
    }
    return handle;
}

static int32_t wave_source_analyzer_call(
    void *handle,
    const uint8_t *source,
    int64_t source_length,
    uint8_t *out_kinds,
    int64_t out_length
) {
    wave_source_analyzer_highlight_fn highlight =
        (wave_source_analyzer_highlight_fn)dlsym(handle, "wave_source_analyzer_highlight");
    if (highlight == NULL) {
        return -2;
    }
    return highlight(source, source_length, out_kinds, out_length);
}
*/
import "C"

import (
	"context"
	"fmt"
	"runtime"
	"sync"
	"unsafe"

	"github.com/wavefnd/wave-platform/internal/sourceanalysis"
)

const maxSourceAnalysisBytes = 1 << 20

type NativeSourceAnalyzer struct {
	mu     sync.RWMutex
	handle unsafe.Pointer
}

func OpenSourceAnalyzer(path string) (*NativeSourceAnalyzer, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	handle := C.wave_source_analyzer_open(cPath, C.int32_t(sourceAnalyzerABI))
	if handle == nil {
		return nil, fmt.Errorf("%w: open %s", ErrModuleUnavailable, path)
	}
	return &NativeSourceAnalyzer{handle: handle}, nil
}

func (analyzer *NativeSourceAnalyzer) Analyze(ctx context.Context, source []byte) (sourceanalysis.Analysis, error) {
	if err := ctx.Err(); err != nil {
		return sourceanalysis.Analysis{}, err
	}
	if len(source) > maxSourceAnalysisBytes {
		return sourceanalysis.Analysis{}, fmt.Errorf("source exceeds Wave analyzer limit")
	}

	analyzer.mu.RLock()
	defer analyzer.mu.RUnlock()
	if analyzer.handle == nil {
		return sourceanalysis.Analysis{}, ErrModuleUnavailable
	}

	bufferSize := len(source)
	if bufferSize == 0 {
		bufferSize = 1
	}
	input := C.malloc(C.size_t(bufferSize))
	if input == nil {
		return sourceanalysis.Analysis{}, fmt.Errorf("allocate Wave analyzer input")
	}
	defer C.free(input)
	if len(source) > 0 {
		C.memcpy(input, unsafe.Pointer(&source[0]), C.size_t(len(source)))
	}
	output := C.calloc(C.size_t(bufferSize), 1)
	if output == nil {
		return sourceanalysis.Analysis{}, fmt.Errorf("allocate Wave analyzer output")
	}
	defer C.free(output)

	status := C.wave_source_analyzer_call(
		analyzer.handle,
		(*C.uint8_t)(input),
		C.int64_t(len(source)),
		(*C.uint8_t)(output),
		C.int64_t(bufferSize),
	)
	runtime.KeepAlive(source)
	if status != 0 {
		return sourceanalysis.Analysis{}, fmt.Errorf("Wave source analyzer returned status %d", int(status))
	}

	kinds := C.GoBytes(output, C.int(len(source)))
	tokens := make([]sourceanalysis.Token, 0)
	for start := 0; start < len(kinds); {
		kind := kinds[start]
		end := start + 1
		for end < len(kinds) && kinds[end] == kind {
			end++
		}
		if name := tokenKindName(kind); name != "" {
			tokens = append(tokens, sourceanalysis.Token{Kind: name, Start: start, End: end})
		}
		start = end
	}

	return sourceanalysis.Analysis{Engine: "wave", ABI: sourceAnalyzerABI, Tokens: tokens}, nil
}

func tokenKindName(kind byte) string {
	switch kind {
	case 1:
		return "keyword"
	case 2:
		return "type"
	case 3:
		return "string"
	case 4:
		return "comment"
	case 5:
		return "number"
	default:
		return ""
	}
}

func (analyzer *NativeSourceAnalyzer) Close() error {
	analyzer.mu.Lock()
	defer analyzer.mu.Unlock()
	if analyzer.handle != nil {
		C.dlclose(analyzer.handle)
		analyzer.handle = nil
	}
	return nil
}
