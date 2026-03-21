package helpers

import (
	"strings"
	"unicode"
)

// SanitizeChatMessage removes control characters from a chat message
// while preserving printable characters, spaces, and common whitespace.
// This prevents XSS via terminal escape sequences and other control char attacks.
func SanitizeChatMessage(message string) string {
	var sb strings.Builder
	sb.Grow(len(message))

	for _, r := range message {
		// Allow printable characters (including Unicode)
		// Allow space, tab, and newline for formatting
		if unicode.IsPrint(r) || r == '\t' || r == '\n' {
			sb.WriteRune(r)
		}
		// Control characters are silently dropped
	}

	return sb.String()
}
