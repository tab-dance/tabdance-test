package inspect

import (
	libghostty "go.mitchellh.com/libghostty"
)

// cursorStyleString returns the visual cursor style ("bar", "block",
// "underline", "block_hollow") using the render state, which is the only
// reliable source for the currently active cursor visual.
func cursorStyleString(term *libghostty.Terminal) string {
	rs, err := libghostty.NewRenderState()
	if err != nil {
		return "bar"
	}
	defer rs.Close()

	if err := rs.Update(term); err != nil {
		return "bar"
	}

	cs, err := rs.CursorVisualStyle()
	if err != nil {
		return "bar"
	}

	switch cs {
	case libghostty.CursorStyleBlock:
		return "block"
	case libghostty.CursorStyleUnderline:
		return "underline"
	case libghostty.CursorStyleBlockHollow:
		return "block_hollow"
	default:
		return "bar"
	}
}
