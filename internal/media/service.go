package media

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/deepteams/webp"
	"golang.org/x/image/draw"

	"github.com/wavefnd/wave-platform/internal/audit"
	"github.com/wavefnd/wave-platform/internal/identifier"
	"github.com/wavefnd/wave-platform/internal/mediapolicy"
)

const (
	MaxInputBytes  = 12 << 20
	MaxOutputBytes = 6 << 20
)

var (
	ErrUnavailable    = errors.New("LunaStev image service is unavailable")
	ErrInvalidImage   = errors.New("invalid image")
	ErrOutputTooLarge = errors.New("converted image is too large")
	assetNamePattern  = regexp.MustCompile(`^image-[0-9]+-[0-9a-f]{32}\.webp$`)
)

type Asset struct {
	ID       string
	Filename string
	Width    int
	Height   int
	Bytes    int64
}

type Service struct {
	root       string
	policy     mediapolicy.Planner
	audit      *audit.Repository
	processing chan struct{}
}

func NewService(root string, policy mediapolicy.Planner, auditRepository *audit.Repository) (*Service, error) {
	if err := os.MkdirAll(root, 0o750); err != nil {
		return nil, fmt.Errorf("create LunaStev image directory: %w", err)
	}
	return &Service{root: root, policy: policy, audit: auditRepository, processing: make(chan struct{}, 2)}, nil
}

func (service *Service) Store(ctx context.Context, reader io.Reader, actorID string) (Asset, error) {
	if service == nil || service.policy == nil {
		return Asset{}, ErrUnavailable
	}
	select {
	case service.processing <- struct{}{}:
		defer func() { <-service.processing }()
	case <-ctx.Done():
		return Asset{}, ctx.Err()
	}

	raw, err := io.ReadAll(io.LimitReader(reader, MaxInputBytes+1))
	if err != nil {
		return Asset{}, fmt.Errorf("read image: %w", err)
	}
	if len(raw) == 0 || len(raw) > MaxInputBytes {
		return Asset{}, mediapolicy.ErrInputTooLarge
	}

	configuration, formatName, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil {
		return Asset{}, ErrInvalidImage
	}
	format, ok := mediaFormat(formatName)
	if !ok {
		return Asset{}, mediapolicy.ErrUnsupportedFormat
	}
	plan, err := service.policy.Plan(ctx, configuration.Width, configuration.Height, int64(len(raw)), format)
	if err != nil {
		return Asset{}, err
	}
	if plan.Width < 1 || plan.Height < 1 || plan.Width > configuration.Width || plan.Height > configuration.Height {
		return Asset{}, fmt.Errorf("%w: Wave policy returned invalid dimensions", ErrInvalidImage)
	}

	decoded, decodedFormat, err := image.Decode(bytes.NewReader(raw))
	if err != nil || decodedFormat != formatName {
		return Asset{}, ErrInvalidImage
	}
	if err := ctx.Err(); err != nil {
		return Asset{}, err
	}
	processed := decoded
	if plan.Width != configuration.Width || plan.Height != configuration.Height {
		resized := image.NewNRGBA(image.Rect(0, 0, plan.Width, plan.Height))
		draw.CatmullRom.Scale(resized, resized.Bounds(), decoded, decoded.Bounds(), draw.Over, nil)
		processed = resized
	}

	encoded, err := encodeWebP(processed, 82)
	if err != nil {
		return Asset{}, fmt.Errorf("encode WebP: %w", err)
	}
	if encoded.Len() > MaxOutputBytes {
		encoded, err = encodeWebP(processed, 68)
		if err != nil {
			return Asset{}, fmt.Errorf("encode compact WebP: %w", err)
		}
	}
	if encoded.Len() > MaxOutputBytes {
		return Asset{}, ErrOutputTooLarge
	}
	if err := ctx.Err(); err != nil {
		return Asset{}, err
	}

	id, err := identifier.New("image")
	if err != nil {
		return Asset{}, err
	}
	filename := id + ".webp"
	if err := service.write(filename, encoded.Bytes()); err != nil {
		return Asset{}, err
	}
	if service.audit != nil {
		auditID, auditErr := identifier.New("audit")
		if auditErr == nil {
			_ = service.audit.Append(audit.Event{ID: auditID, ActorID: "account/" + actorID,
				ResourceID: "media/lunastev/" + id, Action: "media.image.upload", Result: "success", OccurredAt: time.Now().UTC()})
		}
	}
	return Asset{ID: id, Filename: filename, Width: plan.Width, Height: plan.Height, Bytes: int64(encoded.Len())}, nil
}

func (service *Service) Open(filename string) (*os.File, os.FileInfo, error) {
	if service == nil || !assetNamePattern.MatchString(filename) {
		return nil, nil, os.ErrNotExist
	}
	file, err := os.Open(filepath.Join(service.root, filename))
	if err != nil {
		return nil, nil, err
	}
	info, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return nil, nil, err
	}
	if !info.Mode().IsRegular() {
		_ = file.Close()
		return nil, nil, os.ErrNotExist
	}
	return file, info, nil
}

func (service *Service) write(filename string, content []byte) error {
	temporary, err := os.CreateTemp(service.root, ".image-*.tmp")
	if err != nil {
		return fmt.Errorf("create image file: %w", err)
	}
	temporaryName := temporary.Name()
	defer func() { _ = os.Remove(temporaryName) }()
	if err := temporary.Chmod(0o640); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(content); err != nil {
		_ = temporary.Close()
		return fmt.Errorf("write image file: %w", err)
	}
	if err := temporary.Close(); err != nil {
		return fmt.Errorf("close image file: %w", err)
	}
	if err := os.Rename(temporaryName, filepath.Join(service.root, filename)); err != nil {
		return fmt.Errorf("publish image file: %w", err)
	}
	return nil
}

func mediaFormat(name string) (mediapolicy.Format, bool) {
	switch name {
	case "jpeg":
		return mediapolicy.FormatJPEG, true
	case "png":
		return mediapolicy.FormatPNG, true
	case "webp":
		return mediapolicy.FormatWebP, true
	default:
		return 0, false
	}
}

func encodeWebP(source image.Image, quality float32) (*bytes.Buffer, error) {
	output := &bytes.Buffer{}
	options := webp.OptionsForPreset(webp.PresetPicture, quality)
	options.Method = 4
	options.EXIF = nil
	options.ICC = nil
	options.XMP = nil
	if err := webp.Encode(output, source, options); err != nil {
		return nil, err
	}
	return output, nil
}
