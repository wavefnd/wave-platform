package xmlcodec

import (
	"strings"
	"testing"
)

func TestDecodeRejectsTrailingDocument(t *testing.T) {
	var value struct {
		Name string `xml:"name"`
	}

	err := Decode(strings.NewReader(`<value><name>Wave</name></value><other/>`), &value)
	if err == nil {
		t.Fatal("Decode() should reject a trailing document")
	}
}

func TestDecodeLimited(t *testing.T) {
	var value struct{}
	if err := DecodeLimited(strings.NewReader(`<value>too large</value>`), &value, 8); err == nil {
		t.Fatal("DecodeLimited() should enforce the byte limit")
	}
}
