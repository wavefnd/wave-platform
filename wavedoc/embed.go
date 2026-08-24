// Package wavedoc owns the source files for Wave's official documentation.
// Keeping the Markdown at the repository root makes it usable independently
// from the platform server while still allowing the server to embed it.
package wavedoc

import "embed"

var SupportedLocales = []string{"en", "ko", "ja", "zh", "es", "de", "ru", "id", "vi"}

// Content contains every translated Markdown document below this directory.
//
//go:embed */*/*.md
var Content embed.FS

func SupportsLocale(locale string) bool {
	for _, supported := range SupportedLocales {
		if locale == supported {
			return true
		}
	}
	return false
}
