package mediapolicy

import (
	"context"
	"errors"
)

var (
	ErrUnavailable       = errors.New("Wave media policy is unavailable")
	ErrInvalidDimensions = errors.New("image dimensions are invalid")
	ErrInputTooLarge     = errors.New("image input is too large")
	ErrUnsupportedFormat = errors.New("image format is unsupported")
	ErrPixelLimit        = errors.New("image pixel count is too large")
)

type Format int32

const (
	FormatJPEG Format = 1
	FormatPNG  Format = 2
	FormatWebP Format = 3
)

type Plan struct {
	Width  int
	Height int
}

type Planner interface {
	Plan(ctx context.Context, width, height int, inputBytes int64, format Format) (Plan, error)
}
