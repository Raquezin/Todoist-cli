package sanitize

import "testing"

func TestTerminalRemovesControlSequences(t *testing.T) {
	got := Terminal("ok\x1b[2J\nnext\rline\tend")
	want := "ok[2J next lineend"
	if got != want {
		t.Fatalf("Terminal() = %q, want %q", got, want)
	}
}

func TestTerminalLimitCapsByRunes(t *testing.T) {
	got := TerminalLimit("áéíóúabcdef", 8)
	want := "áéíóú..."
	if got != want {
		t.Fatalf("TerminalLimit() = %q, want %q", got, want)
	}
}
