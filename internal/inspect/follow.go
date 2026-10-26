package inspect

import (
	"fmt"

	libghostty "go.mitchellh.com/libghostty"
)

// Follow frame protocol: each frame is a self-contained rendering of the
// terminal's whole state, so a consumer display can never import garbling
// from output it was not around for (mid-session attach's core problem
// with raw PTY deltas). Frames are wrapped in synchronized-update
// sequences when the consumer supports them (ignored otherwise) and
// cleared before paint, making every repaint idempotent.

const (
	syncBegin   = "\x1b[?2026h"
	syncEnd     = "\x1b[?2026l"
	home        = "\x1b[H"
	eraseScreen = "\x1b[2J"
	altEnter    = "\x1b[?1049h"
	altExit     = "\x1b[?1049l"

	screenPrimary   = "primary"
	screenAlternate = "alternate"
)

// FrameState tracks what a follow consumer was last shown so the next
// frame can bridge any mismatch between consumer display state and the
// headless terminal (observable on alternate-screen transitions: the
// renderer mirrors mode state without emitting the "off" half).
type FrameState struct {
	lastContent string
	lastScreen  string
}

// Frame builds the next follow frame from a rendered screen. It returns
// whether the screen changed since the previous frame (a screen-mode
// transition always counts) and the frame bytes.
func (s *FrameState) Frame(res *ViewResult) (string, bool) {
	screen := screenPrimary
	if res.ActiveScreen == screenAlternate {
		screen = screenAlternate
	}

	transition := screen != s.lastScreen
	changed := transition || res.Content != s.lastContent
	if !changed {
		return "", false
	}

	var frame []byte
	if transition && s.lastScreen == screenAlternate && screen == screenPrimary {
		// Leave the alternate screen: the renderer only mirrors enabled
		// modes, so the exit sequence must come from here.
		frame = append(frame, altExit...)
	}
	frame = append(frame, syncBegin...)
	if screen == screenAlternate {
		frame = append(frame, altEnter...)
	}
	frame = append(frame, home...)
	frame = append(frame, eraseScreen...)
	frame = append(frame, res.Content...)
	frame = append(frame, syncEnd...)

	s.lastScreen = screen
	s.lastContent = res.Content

	return string(frame), true
}

// RenderFollowContent renders the whole screen as replayable VT bytes
// with the extra state emission the plain view formatter omits: private
// modes, SGR style, cursor position, keyboard modes, scroll region,
// tabstops, protection and charsets. Ordering inside the emitted bytes is
// mode sequences, then content, then absolute cursor positioning, so the
// consumer ends up with both the screen and the live state the child
// output will draw against.
func RenderFollowContent(term *libghostty.Terminal, trim, unwrap bool) (string, error) {
	f, err := libghostty.NewFormatter(term,
		libghostty.WithFormatterFormat(libghostty.FormatterFormatVT),
		libghostty.WithFormatterTrim(trim),
		libghostty.WithFormatterUnwrap(unwrap),
		libghostty.WithFormatterExtraModes(true),
		libghostty.WithFormatterExtraCursor(true),
		libghostty.WithFormatterExtraStyle(true),
		libghostty.WithFormatterExtraKittyKeyboard(true),
		libghostty.WithFormatterExtraKeyboard(true),
		libghostty.WithFormatterExtraScrollingRegion(true),
		libghostty.WithFormatterExtraTabstops(true),
		libghostty.WithFormatterExtraProtection(true),
		libghostty.WithFormatterExtraCharsets(true),
	)
	if err != nil {
		return "", fmt.Errorf("create formatter: %w", err)
	}
	defer f.Close()

	content, err := f.FormatString()
	if err != nil {
		return "", fmt.Errorf("render frame: %w", err)
	}
	return content, nil
}
