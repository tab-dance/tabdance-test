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
	if cs, err := term.CursorStyle(); err == nil && cs != nil {
		if cs.Bold() {
			styleStr = "block"
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
