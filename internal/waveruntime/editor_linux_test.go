//go:build cgo && linux

package waveruntime

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	editor "github.com/wavefnd/wave-platform/internal/editor"
)

func TestNativeEditor(t *testing.T) {
	library := filepath.Join("..", "..", "build", "wave", "libwave-editor.so")
	if _, err := os.Stat(library); err != nil { t.Skip("WaveEditor native library is not built") }
	engine, err := OpenEditor(library)
	if err != nil { t.Fatal(err) }
	t.Cleanup(func() { _ = engine.Close() })
	result, err := engine.Transform(context.Background(), editor.Request{
		Content: "Wave 언어", SelectionStart: 5, SelectionEnd: 7, Command: editor.CommandBold,
	})
	if err != nil { t.Fatal(err) }
	if result.Content != "Wave **언어**" || result.SelectionStart != 7 || result.SelectionEnd != 9 || result.Engine != "wave" {
		t.Fatalf("result=%+v", result)
	}
}
