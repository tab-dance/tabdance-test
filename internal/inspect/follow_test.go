package inspect

import (
	"strings"
	"testing"

	libghostty "go.mitchellh.com/libghostty"
)

// TestFrameStateTransitions checks the frame protocol: idempotent repaint
// with synchronized updates, alternate-screen enter on switch, and an
// explicit exit sequence when returning to primary (the renderer mirrors
// only enabled modes, so the "off" half must come from the frame layer).
func TestFrameStateTransitions(t *testing.T) {
	st := &FrameState{}

	frame, changed := st.Frame(&ViewResult{Content: "hello"})
	if !changed {
		t.Fatal("first frame must always be emitted")
	}
	for _, want := range []string{syncBegin, home, eraseScreen, "hello", syncEnd} {
		if !strings.Contains(frame, want) {
			t.Fatalf("first frame missing %q: %q", want, frame)
		}
	}
	if strings.Contains(frame, altEnter) || strings.Contains(frame, altExit) {
		t.Fatalf("primary frame must not toggle the alternate screen: %q", frame)
	}

	if _, changed := st.Frame(&ViewResult{Content: "hello"}); changed {
		t.Fatal("unchanged screen must produce no frame")
	}

	frame, changed = st.Frame(&ViewResult{Content: "menu", ActiveScreen: "alternate"})
	if !changed {
		t.Fatal("screen transition must produce a frame")
	}
	if !strings.HasPrefix(frame, altEnter) {
		if i := strings.Index(frame, altEnter); i < 0 || i > strings.Index(frame, "menu") {
			t.Fatalf("alternate frame must enter the alternate screen before painting: %q", frame)
		}
	}

	frame, changed = st.Frame(&ViewResult{Content: "prompt> done"})
	if !changed {
		t.Fatal("screen transition must produce a frame")
	}
	if !strings.HasPrefix(frame, altExit) {
		t.Fatalf("frame after alternate-screen exit must start with the exit sequence: %q", frame)
	}
}

// TestFollowFrameParity replays follow frames into a fresh consumer
// terminal, the way a watching client experiences them. After every stage
// the consumer grid must match the source grid, including through
// cursor-relative incremental redraws (which only work because each frame
// carries modes, style and cursor position) and the alternate-screen
// exit. Because frames repaint the complete active screen, parity holds
// even for content the consumer never received during the attach.
func TestFollowFrameParity(t *testing.T) {
	renderPlain := func(term *libghostty.Terminal) string {
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

	src, err := libghostty.NewTerminal(libghostty.WithSize(40, 10), libghostty.WithMaxScrollbackLines(500))
	if err != nil {
		t.Fatalf("create src: %v", err)
	}
	defer src.Close()
	consumer, err := libghostty.NewTerminal(libghostty.WithSize(40, 10), libghostty.WithMaxScrollbackLines(500))
	if err != nil {
		t.Fatalf("create consumer: %v", err)
	}
	defer consumer.Close()

	steps := [][]byte{
		// Shell session, then a full-screen menu with hidden cursor and
		// some style state set by the TUI.
		[]byte("prompt> git \r\n"),
		[]byte("\x1b[?1049h\x1b[?25l\x1b[?2004h\x1b[1;41mbold title\x1b[0m\x1b[4;5Hentry\x1b[6;5Hlabel\x1b[H"),
		// Cursor-relative incremental redraw (no absolute CUP), the
		// pattern raw deltas garble without a sync packet.
		[]byte("\x1b[9;5Hrefresh\x1b[6;5Hlabel2\x1b[H"),
		// Exit back to primary mid-session.
		[]byte("\x1b[?1049lmore shell output\r\n\x1b[?25h"),
	}

	st := &FrameState{}
	for i, step := range steps {
		src.VTWrite(step)

		content, err := RenderFollowContent(src, false, false)
		if err != nil {
			t.Fatalf("step %d render: %v", i, err)
		}
		screen, _ := src.ActiveScreen()
		screenStr := "primary"
		if screen == libghostty.ScreenAlternate {
			screenStr = "alternate"
		}
		frame, _ := st.Frame(&ViewResult{Content: content, ActiveScreen: screenStr})
		if frame == "" {
			t.Fatalf("step %d frame deduped unexpectedly", i)
		}
		consumer.Write([]byte(frame))

		if got, want := renderPlain(consumer), renderPlain(src); got != want {
			t.Fatalf("step %d grid mismatch:\nconsumer: %q\nsrc:      %q", i, got, want)
		}
	}

	// Unchanged screen renders dedupe (frames stop flowing when nothing
	// happens), which keeps follow quiet on idle terminals.
	content, _ := RenderFollowContent(src, false, false)
	if _, changed := st.Frame(&ViewResult{Content: content}); changed {
		t.Fatal("idle terminal must not produce a frame")
	}
}
