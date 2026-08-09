package media

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/png"
	"os"
	"testing"

	"github.com/deepteams/webp"

	"github.com/wavefnd/wave-platform/internal/mediapolicy"
)

type fixedPolicy struct {
	plan mediapolicy.Plan
	err  error
}

func (policy fixedPolicy) Plan(context.Context, int, int, int64, mediapolicy.Format) (mediapolicy.Plan, error) {
	return policy.plan, policy.err
}

func TestStoreConvertsAndResizesToWebP(t *testing.T) {
	root := t.TempDir()
	service, err := NewService(root, fixedPolicy{plan: mediapolicy.Plan{Width: 100, Height: 50}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	source := image.NewNRGBA(image.Rect(0, 0, 400, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 400; x++ {
			source.Set(x, y, color.NRGBA{R: uint8(x % 255), G: uint8(y % 255), B: 120, A: 255})
		}
	}
	var input bytes.Buffer
	if err := png.Encode(&input, source); err != nil {
		t.Fatal(err)
	}

	asset, err := service.Store(context.Background(), &input, "owner")
	if err != nil {
		t.Fatal(err)
	}
	if asset.Width != 100 || asset.Height != 50 || asset.Bytes < 1 {
		t.Fatalf("asset = %+v", asset)
	}
	stored, err := os.ReadFile(root + "/" + asset.Filename)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored) < 12 || string(stored[:4]) != "RIFF" || string(stored[8:12]) != "WEBP" {
		t.Fatalf("stored image is not WebP: %x", stored[:min(12, len(stored))])
	}
	configuration, err := webp.DecodeConfig(bytes.NewReader(stored))
	if err != nil {
		t.Fatal(err)
	}
	if configuration.Width != 100 || configuration.Height != 50 {
		t.Fatalf("stored dimensions = %dx%d", configuration.Width, configuration.Height)
	}
}

func TestStoreRequiresWavePolicyAndEnforcesInputLimit(t *testing.T) {
	service, err := NewService(t.TempDir(), nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.Store(context.Background(), bytes.NewReader([]byte("image")), "owner"); !errors.Is(err, ErrUnavailable) {
		t.Fatalf("unavailable error = %v", err)
	}
	service.policy = fixedPolicy{plan: mediapolicy.Plan{Width: 1, Height: 1}}
	if _, err := service.Store(context.Background(), bytes.NewReader(make([]byte, MaxInputBytes+1)), "owner"); !errors.Is(err, mediapolicy.ErrInputTooLarge) {
		t.Fatalf("limit error = %v", err)
	}
}
