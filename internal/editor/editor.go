package editor

import (
	"context"
	"errors"
	"fmt"
	"unicode"
)

const (
	MaxDocumentRunes = 200_000
	MaxDocumentBytes = 1 << 20
)

var ErrInvalidRequest = errors.New("invalid editor request")

type Command string

const (
	CommandBold          Command = "bold"
	CommandItalic        Command = "italic"
	CommandInlineCode    Command = "inline-code"
	CommandHeading       Command = "heading"
	CommandQuote         Command = "quote"
	CommandUnorderedList Command = "unordered-list"
	CommandLink          Command = "link"
)

type Request struct {
	Content        string  `xml:"content"`
	SelectionStart int     `xml:"selection-start"`
	SelectionEnd   int     `xml:"selection-end"`
	Command        Command `xml:"command"`
}

type Result struct {
	Content        string `xml:"content"`
	SelectionStart int    `xml:"selection-start"`
	SelectionEnd   int    `xml:"selection-end"`
	Engine         string `xml:"engine"`
	Lines          int    `xml:"lines"`
	Words          int    `xml:"words"`
}

type Engine interface {
	Transform(context.Context, Request) (Result, error)
}

type ReferenceEngine struct{}

func (ReferenceEngine) Transform(ctx context.Context, request Request) (Result, error) {
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	if err := Validate(request); err != nil {
		return Result{}, err
	}
	prefix, suffix := wrappers(request.Command)
	runes := []rune(request.Content)
	output := make([]rune, 0, len(runes)+len([]rune(prefix))+len([]rune(suffix)))
	output = append(output, runes[:request.SelectionStart]...)
	output = append(output, []rune(prefix)...)
	output = append(output, runes[request.SelectionStart:request.SelectionEnd]...)
	output = append(output, []rune(suffix)...)
	output = append(output, runes[request.SelectionEnd:]...)
	content := string(output)
	prefixRunes := len([]rune(prefix))
	return Result{
		Content:        content,
		SelectionStart: request.SelectionStart + prefixRunes,
		SelectionEnd:   request.SelectionEnd + prefixRunes,
		Engine:         "go",
		Lines:          countLines(content),
		Words:          countWords(content),
	}, nil
}

func Validate(request Request) error {
	length := len([]rune(request.Content))
	if length > MaxDocumentRunes || len(request.Content) > MaxDocumentBytes {
		return fmt.Errorf("%w: document exceeds %d characters", ErrInvalidRequest, MaxDocumentRunes)
	}
	if request.SelectionStart < 0 || request.SelectionEnd < request.SelectionStart || request.SelectionEnd > length {
		return fmt.Errorf("%w: selection is outside the document", ErrInvalidRequest)
	}
	if _, suffix := wrappers(request.Command); suffix == "" && request.Command != CommandHeading && request.Command != CommandQuote && request.Command != CommandUnorderedList {
		return fmt.Errorf("%w: unsupported command", ErrInvalidRequest)
	}
	return nil
}

func wrappers(command Command) (string, string) {
	switch command {
	case CommandBold:
		return "**", "**"
	case CommandItalic:
		return "*", "*"
	case CommandInlineCode:
		return "`", "`"
	case CommandHeading:
		return "## ", ""
	case CommandQuote:
		return "> ", ""
	case CommandUnorderedList:
		return "- ", ""
	case CommandLink:
		return "[", "](https://)"
	default:
		return "", ""
	}
}

func countLines(content string) int {
	lines := 1
	for _, current := range content {
		if current == '\n' {
			lines++
		}
	}
	return lines
}

func countWords(content string) int {
	words := 0
	inWord := false
	for _, current := range content {
		space := unicode.IsSpace(current)
		if !space && !inWord {
			words++
		}
		inWord = !space
	}
	return words
}
