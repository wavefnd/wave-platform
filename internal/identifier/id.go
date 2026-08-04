package identifier

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

func New(prefix string) (string, error) {
	random := make([]byte, 16)

	if _, err := rand.Read(random); err != nil {
		return "", fmt.Errorf("generate identifier: %w", err)
	}

	value := fmt.Sprintf(
		"%d-%s",
		time.Now().UTC().UnixMilli(),
		hex.EncodeToString(random),
	)

	if prefix == "" {
		return value, nil
	}

	return prefix + "-" + value, nil
}
