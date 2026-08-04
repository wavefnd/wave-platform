package account

import (
	"errors"
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const maxLocalPartBytes = 60

func LocalPart(displayName string) (string, error) {
	var builder strings.Builder
	separator := false
characters:
	for _, character := range strings.TrimSpace(strings.ToLower(norm.NFKC.String(displayName))) {
		switch {
		case unicode.IsLetter(character) || unicode.IsDigit(character):
			separatorBytes := 0
			if separator && builder.Len() > 0 {
				separatorBytes = 1
			}
			if builder.Len()+separatorBytes+utf8.RuneLen(character) > maxLocalPartBytes {
				break characters
			}
			if separatorBytes == 1 {
				builder.WriteByte('-')
			}
			builder.WriteRune(character)
			separator = false
		case unicode.IsSpace(character) || unicode.IsPunct(character) || unicode.IsSymbol(character):
			separator = builder.Len() > 0
		}
	}
	value := strings.Trim(builder.String(), "-")
	if value == "" {
		return "", errors.New("display name must contain a letter or number")
	}
	return value, nil
}

func Address(username, domain string) (string, error) {
	username = strings.TrimSpace(strings.ToLower(username))
	domain = strings.Trim(strings.ToLower(strings.TrimSpace(domain)), ".")
	if username == "" || domain == "" || !strings.Contains(domain, ".") || strings.ContainsAny(domain, " /@") {
		return "", fmt.Errorf("invalid username or mail domain")
	}
	return username + "@" + domain, nil
}
