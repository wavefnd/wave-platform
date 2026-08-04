package storage

import (
	"fmt"
	"strings"
)

func Key(parts ...string) []byte {
	cleaned := make([]string, 0, len(parts))

	for _, part := range parts {
		part = strings.Trim(part, "/")

		if part != "" {
			cleaned = append(cleaned, part)
		}
	}

	return []byte(strings.Join(cleaned, "/"))
}

func Prefix(parts ...string) []byte {
	key := Key(parts...)

	if len(key) == 0 {
		return nil
	}

	return []byte(fmt.Sprintf("%s/", key))
}
