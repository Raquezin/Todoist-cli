package sanitize

import (
	"strings"
	"unicode"
)

const ellipsis = "..."

func isUnsafe(r rune) bool {
	return unicode.IsControl(r) || unicode.Is(unicode.Cf, r)
}

// Terminal returns a single-line string that cannot emit terminal control
// sequences when printed.
func Terminal(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")

	var b strings.Builder
	b.Grow(len(s))
	for _, r := range s {
		if isUnsafe(r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// TerminalLimit sanitizes s for terminal output and caps it to max runes.
func TerminalLimit(s string, max int) string {
	s = Terminal(s)
	if max <= 0 {
		return ""
	}

	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	if max <= len(ellipsis) {
		return string(runes[:max])
	}
	return string(runes[:max-len(ellipsis)]) + ellipsis
}
