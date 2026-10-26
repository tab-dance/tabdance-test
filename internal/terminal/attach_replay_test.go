package terminal

import (
	"bytes"
	"testing"

	libghostty "go.mitchellh.com/libghostty"
)

// renderPlain renders a terminal's active screen like get_view does
// (plain, no trim, no unwrap) for grid comparisons.
func renderPlain(t *testing.T, term *libghostty.Terminal) string {
	t.Helper()
	f, err := libghostty.NewFormatter(term,
		libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
		libghostty.WithFormatterTrim(false),
		libghostty.WithFormatterUnwrap(false),
	)
	if err != nil {
		t.Fatalf("create formatter: %v", err)
	}
	defer f.Close()
	s, err := f.FormatString()
	if err != nil {
		t.Fatalf("format: %v", err)
	}
	return s
}

// TestSubscribeWithSnapshotReplay simulates an attach: a consumer terminal
// receives the snapshot, then every chunk the source terminal produces.
// The consumer grid must match the source grid after the initial state
// (an alternate-screen TUI) and after further relative redraws, which only
// work when mode state, SGR style and cursor position were replayed.
func TestSubscribeWithSnapshotReplay(t *testing.T) {
	src, err := libghostty.NewTerminal(libghostty.WithSize(40, 10), libghostty.WithMaxScrollbackLines(500))
	if err != nil {
		t.Fatalf("create src: %v", err)
	}
	defer src.Close()

	// Shell output on primary, then a full-screen menu UI on the
	// alternate screen with hidden cursor and bracketed paste on.
	src.VTWrite([]byte("prompt> git \r\n"))
	src.VTWrite([]byte("\x1b[?1049h\x1b[?25l\x1b[?2004h"))
	src.VTWrite([]byte("\x1b[3;5Hmenu\x1b[6;5Hlabel\x1b[H"))

	consumer, err := libghostty.NewTerminal(libghostty.WithSize(40, 10), libghostty.WithMaxScrollbackLines(500))
	if err != nil {
		t.Fatalf("create consumer: %v", err)
	}
	defer consumer.Close()

	// The consumer simulates a real client: it already had the same
	// primary history as the source at attach time, because the client
	// attaches from the very shell it launched into the TUI.
	consumer.VTWrite([]byte("prompt> git \r\n"))

	m := &ManagedTerminal{Terminal: src, done: make(chan struct{})}
	sub, snapshot, err := m.SubscribeWithSnapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}
	defer m.Unsubscribe(sub)

	consumer.Write(snapshot)
	if got, want := renderPlain(t, consumer), renderPlain(t, src); got != want {
		t.Fatalf("screen mismatch after snapshot:\nconsumer: %q\nsrc:      %q", got, want)
	}

	// Replay further deltas: a relative-flow redraw (no absolute CUP,
	// only cursor moves that assume a synchronized cursor) that rewrites
	// the "label" cell by moving from the snapshot cursor position.
	deltas := [][]byte{
		[]byte("\x1b[6;5Hlabel2\x1b[3;5H"),
		[]byte("\x1b[?1049lprompt> done\r\n\x1b[?25h"),
	}
	wantScreens := []string{
		renderPlain(t, src),
		"",
	}
	for i, delta := range deltas {
		src.VTWrite(delta)
		// This unit test has no ptyReadLoop driving the subscriber; deliver
		// the chunk manually exactly like the read loop would.
		if !sub.deliver(delta) {
			t.Fatalf("deliver failed for delta %d", i)
		}
		select {
		case chunk := <-sub.Channel():
			consumer.Write(chunk)
		default:
			t.Fatalf("no chunk broadcast for delta %d", i)
		}
		if screen, _ := src.ActiveScreen(); screen == libghostty.ScreenAlternate {
			if got, want := renderPlain(t, consumer), renderPlain(t, src); got != want {
				t.Fatalf("screen mismatch after delta %d:\nconsumer: %q\nsrc:      %q", i, got, want)
			}
		} else if want := wantScreens[i]; want != "" {
			t.Fatalf("src unexpectedly non-empty outside altscreen")
		}
	}

	// The final delta exits the alternate screen; the consumer must pop
	// back to primary and restore cursor visibility mode state it
	// replayed at attach time.
	if screen, _ := src.ActiveScreen(); screen != libghostty.ScreenPrimary {
		t.Fatalf("src not on primary after exit delta: %d", screen)
	}
	if screen, _ := consumer.ActiveScreen(); screen != libghostty.ScreenPrimary {
		t.Fatalf("consumer not on primary after exit delta: %d", screen)
	}
	if vis, _ := src.CursorVisible(); !vis {
		t.Fatalf("src cursor not visible after exit delta")
	}
	if vis, _ := consumer.CursorVisible(); !vis {
		t.Fatalf("consumer cursor not visible after exit delta")
	}
	if got, want := renderPlain(t, consumer), renderPlain(t, src); got != want {
		t.Fatalf("primary mismatch after altscreen exit:\nconsumer: %q\nsrc:      %q", got, want)
	}
}

// TestSubscribeWithSnapshotModeSync checks that the snapshot carries mode
// sequences before the grid content and an absolute cursor position last,
// so a consumer can turn them into a synchronized rendering state.
func TestSubscribeWithSnapshotModeSync(t *testing.T) {
	src, err := libghostty.NewTerminal(libghostty.WithSize(20, 5))
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	defer src.Close()
	src.VTWrite([]byte("\x1b[?1049h\x1b[?25l\x1b[?2004h\x1b[2;3Hmenu\x1b[4;2H"))

	m := &ManagedTerminal{Terminal: src, done: make(chan struct{})}
	_, snapshot, err := m.SubscribeWithSnapshot()
	if err != nil {
		t.Fatalf("snapshot: %v", err)
	}

	s := string(snapshot)
	for _, want := range []string{"\x1b[?1049h", "\x1b[?25l"} {
		if !bytes.Contains(snapshot, []byte(want)) {
			t.Fatalf("snapshot missing state sequence %q: %q", want, s)
		}
	}
	// Modes must come before the grid content, otherwise a mode like
	// alternate screen would clear the freshly drawn content on the
	// consumer. CUP must appear after the content (trailing style reset
	// is metadata-only and harmless).
	if bytes.Index(snapshot, []byte("menu")) < bytes.Index(snapshot, []byte("\x1b[?1049h")) {
		t.Fatalf("mode sync not emitted before content: %q", s)
	}
	if bytes.Index(snapshot, []byte("\x1b[4;2H")) < bytes.Index(snapshot, []byte("menu")) {
		t.Fatalf("cursor position not emitted after content: %q", s)
	}
}
