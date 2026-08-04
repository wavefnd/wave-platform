package xmlcodec

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
)

const DefaultRequestLimit int64 = 1 << 20

func Decode(reader io.Reader, destination any) error {
	return DecodeLimited(reader, destination, DefaultRequestLimit)
}

func DecodeLimited(reader io.Reader, destination any, limit int64) error {
	data, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return fmt.Errorf("read XML: %w", err)
	}

	if int64(len(data)) > limit {
		return fmt.Errorf("XML exceeds %d byte limit", limit)
	}

	decoder := xml.NewDecoder(bytes.NewReader(data))
	decoder.Strict = true

	if err := decoder.Decode(destination); err != nil {
		return fmt.Errorf("decode XML: %w", err)
	}

	if token, err := decoder.Token(); err != io.EOF {
		if err != nil {
			return fmt.Errorf("decode trailing XML: %w", err)
		}

		return fmt.Errorf("unexpected trailing XML token %T", token)
	}

	return nil
}

func Write(
	writer http.ResponseWriter,
	status int,
	value any,
) error {
	data, err := xml.MarshalIndent(value, "", "    ")
	if err != nil {
		return fmt.Errorf("encode XML: %w", err)
	}

	writer.Header().Set("Content-Type", "application/xml; charset=utf-8")
	writer.Header().Set("X-Content-Type-Options", "nosniff")
	writer.WriteHeader(status)

	if _, err := writer.Write([]byte(xml.Header)); err != nil {
		return fmt.Errorf("write XML header: %w", err)
	}

	if _, err := writer.Write(data); err != nil {
		return fmt.Errorf("write XML response: %w", err)
	}

	return nil
}
