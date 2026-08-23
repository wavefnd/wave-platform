package editor

import (
	"context"
	"errors"
	"testing"
)

func TestReferenceEngineTransformsUnicodeSelection(t *testing.T) {
	result, err := (ReferenceEngine{}).Transform(context.Background(), Request{
		Content: "Wave 언어", SelectionStart: 5, SelectionEnd: 7, Command: CommandBold,
	})
	if err != nil {
		t.Fatal(err)
	}
	if result.Content != "Wave **언어**" || result.SelectionStart != 7 || result.SelectionEnd != 9 {
		t.Fatalf("result=%+v", result)
	}
	if result.Engine != "go" || result.Lines != 1 || result.Words != 2 {
		t.Fatalf("metadata=%+v", result)
	}
}

func TestReferenceEngineRejectsInvalidRequests(t *testing.T) {
	_, err := (ReferenceEngine{}).Transform(context.Background(), Request{Content: "Wave", SelectionStart: 1, SelectionEnd: 5, Command: CommandBold})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error=%v", err)
	}
	_, err = (ReferenceEngine{}).Transform(context.Background(), Request{Content: "Wave", Command: "unknown"})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("error=%v", err)
	}
}
