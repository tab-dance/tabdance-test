package inspect

import (
	"encoding/json"

	libghostty "go.mitchellh.com/libghostty"
)

type CursorResult struct {
	X            uint16 `json:"x"`
	Y            uint16 `json:"y"`
	Visible      bool   `json:"visible"`
	PendingWrap  bool   `json:"pending_wrap"`
	Style        string `json:"style"`
}

func GetCursor(term *libghostty.Terminal) *CursorResult {
	x, _ := term.CursorX()
	y, _ := term.CursorY()
	visible, _ := term.CursorVisible()
	pendingWrap, _ := term.CursorPendingWrap()

	styleStr := "bar"
	rs, err := libghostty.NewRenderState()
	if err == nil {
		defer rs.Close()
		if err := rs.Update(term); err == nil {
			if cs, err := rs.CursorVisualStyle(); err == nil {
				switch cs {
				case libghostty.CursorVisualStyleBlock:
					styleStr = "block"
				case libghostty.CursorVisualStyleUnderline:
					styleStr = "underline"
				case libghostty.CursorVisualStyleBlockHollow:
					styleStr = "block_hollow"
				}
			}
		}
	}

	return &CursorResult{
		X:           x,
		Y:           y,
		Visible:     visible,
		PendingWrap: pendingWrap,
		Style:       styleStr,
	}
}

func (c *CursorResult) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}
