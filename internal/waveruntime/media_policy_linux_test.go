//go:build cgo && linux

package waveruntime

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/wavefnd/wave-platform/internal/mediapolicy"
)

func TestNativeMediaPolicy(t *testing.T) {
	library := filepath.Join("..", "..", "build", "wave", "libwave-media-policy.so")
	if _, err := os.Stat(library); err != nil {
		t.Skip("Wave media policy artifact has not been built")
	}

	policy, err := OpenMediaPolicy(library)
	if err != nil {
		t.Fatalf("open policy: %v", err)
	}
	t.Cleanup(func() { _ = policy.Close() })

	plan, err := policy.Plan(context.Background(), 4000, 2000, 1<<20, mediapolicy.FormatJPEG)
	if err != nil {
		t.Fatal(err)
	}
	if plan.Width != 1920 || plan.Height != 960 {
		t.Fatalf("plan = %+v", plan)
	}
	if _, err := policy.Plan(context.Background(), 12000, 12000, 1<<20, mediapolicy.FormatPNG); !errors.Is(err, mediapolicy.ErrPixelLimit) {
		t.Fatalf("pixel limit error = %v", err)
	}
}
