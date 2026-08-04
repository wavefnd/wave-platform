package storage

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type BlobMetadata struct {
	XMLName xml.Name `xml:"https://wave-lang.dev/ns/platform/storage/v1 blob"`
	Hash    string   `xml:"hash"`
	Size    int      `xml:"size"`
	MIME    string   `xml:"mime"`
}

func (database *Database) PutBlob(content []byte, mimeType string) (string, error) {
	digest := sha256.Sum256(content)
	hash := hex.EncodeToString(digest[:])
	path := database.blobPath(hash)

	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return "", fmt.Errorf("create blob directory: %w", err)
	}
	if err := os.WriteFile(path, content, 0o640); err != nil {
		return "", fmt.Errorf("write blob %q: %w", hash, err)
	}

	metadata, err := xml.Marshal(BlobMetadata{Hash: hash, Size: len(content), MIME: mimeType})
	if err != nil {
		return "", fmt.Errorf("encode blob metadata: %w", err)
	}
	if err := database.Set(Key("blob", "object", hash), metadata); err != nil {
		return "", err
	}
	return hash, nil
}

func (database *Database) Blob(hash string) ([]byte, error) {
	content, err := os.ReadFile(database.blobPath(hash))
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("read blob %q: %w", hash, err)
	}
	return content, nil
}

func (database *Database) blobPath(hash string) string {
	if len(hash) < 4 {
		return filepath.Join(database.Root, "blobs", "invalid", hash)
	}
	return filepath.Join(database.Root, "blobs", hash[:2], hash[2:4], hash)
}
