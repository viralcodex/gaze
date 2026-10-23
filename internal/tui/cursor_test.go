package tui

import "testing"

func TestCursorEscapeMapping(t *testing.T) {
	cases := []struct {
		cursor Cursor
		want   string
	}{
		{CursorDefault, DefaultCursor},
		{CursorPointer, PointerCursor},
		{CursorText, TextCursor},
		{Cursor(99), DefaultCursor}, // unknown falls back to default
	}

	for _, c := range cases {
		if got := cursorEscape(c.cursor); got != c.want {
			t.Errorf("cursorEscape(%d) = %q, want %q", c.cursor, got, c.want)
		}
	}
}

func TestApplyCursorTracksState(t *testing.T) {
	p := &Program{}

	// zero value is CursorDefault; applying it again is a no-op transition
	if p.cursor != CursorDefault {
		t.Fatalf("initial cursor = %d, want %d", p.cursor, CursorDefault)
	}

	p.applyCursor(CursorPointer)
	if p.cursor != CursorPointer {
		t.Fatalf("cursor = %d, want %d", p.cursor, CursorPointer)
	}

	p.applyCursor(CursorDefault)
	if p.cursor != CursorDefault {
		t.Fatalf("cursor = %d, want %d", p.cursor, CursorDefault)
	}
}
