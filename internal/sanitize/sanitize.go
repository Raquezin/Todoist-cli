package sanitize

import (
	"strings"
	"unicode"
)

const ellipsis = "..."

// Terminal returns a single-line string that cannot emit terminal control
// sequences when printed.
func Terminal(s string) string {
	s = strings.ReplaceAll(s, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")

	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) || unicode.Is(unicode.Cf, r) {
			return -1
		}
		return r
	}, s)
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
