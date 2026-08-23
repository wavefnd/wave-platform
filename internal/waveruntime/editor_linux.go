//go:build cgo && linux

package waveruntime

/*
#cgo LDFLAGS: -ldl
#include <dlfcn.h>
#include <stdint.h>
#include <stdlib.h>
#include <string.h>

typedef int32_t (*wave_editor_abi_version_fn)(void);
typedef int64_t (*wave_editor_transform_fn)(
    const uint8_t *, int64_t, int64_t, int64_t, int32_t,
    uint8_t *, int64_t, int64_t *, int64_t *
);
typedef int32_t (*wave_editor_analyze_fn)(const uint8_t *, int64_t, int64_t *, int64_t *);

static void *wave_editor_open(const char *path, int32_t expected_abi) {
    void *handle = dlopen(path, RTLD_NOW | RTLD_LOCAL);
    if (handle == NULL) return NULL;
    wave_editor_abi_version_fn abi = (wave_editor_abi_version_fn)dlsym(handle, "wave_editor_abi_version");
    wave_editor_transform_fn transform = (wave_editor_transform_fn)dlsym(handle, "wave_editor_transform");
    wave_editor_analyze_fn analyze = (wave_editor_analyze_fn)dlsym(handle, "wave_editor_analyze");
    if (abi == NULL || transform == NULL || analyze == NULL || abi() != expected_abi) {
        dlclose(handle);
        return NULL;
    }
    return handle;
}

static int64_t wave_editor_call_transform(
    void *handle, const uint8_t *source, int64_t source_length,
    int64_t selection_start, int64_t selection_end, int32_t command,
    uint8_t *output, int64_t output_capacity,
    int64_t *out_selection_start, int64_t *out_selection_end
) {
    wave_editor_transform_fn transform = (wave_editor_transform_fn)dlsym(handle, "wave_editor_transform");
    if (transform == NULL) return -3;
    return transform(source, source_length, selection_start, selection_end, command,
        output, output_capacity, out_selection_start, out_selection_end);
}

static int32_t wave_editor_call_analyze(
    void *handle, const uint8_t *source, int64_t source_length, int64_t *out_lines, int64_t *out_words
) {
    wave_editor_analyze_fn analyze = (wave_editor_analyze_fn)dlsym(handle, "wave_editor_analyze");
    if (analyze == NULL) return -2;
    return analyze(source, source_length, out_lines, out_words);
}
*/
import "C"

import (
	"context"
	"fmt"
	"sync"
	"unicode/utf8"
	"unsafe"

	editor "github.com/wavefnd/wave-platform/internal/editor"
)

const maxEditorBytes = 1 << 20

type NativeEditor struct {
	mu     sync.RWMutex
	handle unsafe.Pointer
}

func OpenEditor(path string) (*NativeEditor, error) {
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	handle := C.wave_editor_open(cPath, C.int32_t(editorABI))
	if handle == nil {
		return nil, fmt.Errorf("%w: open %s", ErrModuleUnavailable, path)
	}
	return &NativeEditor{handle: handle}, nil
}

func (engine *NativeEditor) Transform(ctx context.Context, request editor.Request) (editor.Result, error) {
	if err := ctx.Err(); err != nil {
		return editor.Result{}, err
	}
	if err := editor.Validate(request); err != nil {
		return editor.Result{}, err
	}
	inputBytes := []byte(request.Content)
	if len(inputBytes) > maxEditorBytes {
		return editor.Result{}, fmt.Errorf("%w: document exceeds native byte limit", editor.ErrInvalidRequest)
	}
	start, ok := runeToByteOffset(request.Content, request.SelectionStart)
	if !ok {
		return editor.Result{}, fmt.Errorf("%w: invalid selection start", editor.ErrInvalidRequest)
	}
	end, ok := runeToByteOffset(request.Content, request.SelectionEnd)
	if !ok {
		return editor.Result{}, fmt.Errorf("%w: invalid selection end", editor.ErrInvalidRequest)
	}
	command, ok := editorCommandID(request.Command)
	if !ok {
		return editor.Result{}, fmt.Errorf("%w: unsupported command", editor.ErrInvalidRequest)
	}

	engine.mu.RLock()
	defer engine.mu.RUnlock()
	if engine.handle == nil {
		return editor.Result{}, ErrModuleUnavailable
	}

	inputSize := len(inputBytes)
	if inputSize == 0 {
		inputSize = 1
	}
	input := C.malloc(C.size_t(inputSize))
	if input == nil {
		return editor.Result{}, fmt.Errorf("allocate WaveEditor input")
	}
	defer C.free(input)
	if len(inputBytes) > 0 {
		C.memcpy(input, unsafe.Pointer(&inputBytes[0]), C.size_t(len(inputBytes)))
	}

	outputCapacity := len(inputBytes) + 32
	output := C.malloc(C.size_t(outputCapacity))
	if output == nil {
		return editor.Result{}, fmt.Errorf("allocate WaveEditor output")
	}
	defer C.free(output)
	var outputStart, outputEnd C.int64_t
	outputLength := C.wave_editor_call_transform(
		engine.handle, (*C.uint8_t)(input), C.int64_t(len(inputBytes)), C.int64_t(start), C.int64_t(end), C.int32_t(command),
		(*C.uint8_t)(output), C.int64_t(outputCapacity), &outputStart, &outputEnd,
	)
	if outputLength < 0 {
		return editor.Result{}, fmt.Errorf("WaveEditor returned status %d", int64(outputLength))
	}
	contentBytes := C.GoBytes(output, C.int(outputLength))
	if !utf8.Valid(contentBytes) {
		return editor.Result{}, fmt.Errorf("WaveEditor returned invalid UTF-8")
	}
	content := string(contentBytes)
	selectionStart, ok := byteToRuneOffset(content, int(outputStart))
	if !ok {
		return editor.Result{}, fmt.Errorf("WaveEditor returned invalid selection start")
	}
	selectionEnd, ok := byteToRuneOffset(content, int(outputEnd))
	if !ok {
		return editor.Result{}, fmt.Errorf("WaveEditor returned invalid selection end")
	}
	var lines, words C.int64_t
	if status := C.wave_editor_call_analyze(engine.handle, (*C.uint8_t)(output), outputLength, &lines, &words); status != 0 {
		return editor.Result{}, fmt.Errorf("WaveEditor analysis returned status %d", int(status))
	}
	return editor.Result{Content: content, SelectionStart: selectionStart, SelectionEnd: selectionEnd,
		Engine: "wave", Lines: int(lines), Words: int(words)}, nil
}

func editorCommandID(command editor.Command) (int, bool) {
	switch command {
	case editor.CommandBold:
		return 1, true
	case editor.CommandItalic:
		return 2, true
	case editor.CommandInlineCode:
		return 3, true
	case editor.CommandHeading:
		return 4, true
	case editor.CommandQuote:
		return 5, true
	case editor.CommandUnorderedList:
		return 6, true
	case editor.CommandLink:
		return 7, true
	default:
		return 0, false
	}
}

func runeToByteOffset(value string, offset int) (int, bool) {
	if offset < 0 {
		return 0, false
	}
	if offset == 0 {
		return 0, true
	}
	count := 0
	for index := range value {
		if count == offset {
			return index, true
		}
		count++
	}
	if count == offset {
		return len(value), true
	}
	return 0, false
}

func byteToRuneOffset(value string, offset int) (int, bool) {
	if offset < 0 || offset > len(value) || (offset < len(value) && !utf8.RuneStart(value[offset])) {
		return 0, false
	}
	return utf8.RuneCountInString(value[:offset]), true
}

func (engine *NativeEditor) Close() error {
	engine.mu.Lock()
	defer engine.mu.Unlock()
	if engine.handle != nil {
		C.dlclose(engine.handle)
		engine.handle = nil
	}
	return nil
}
