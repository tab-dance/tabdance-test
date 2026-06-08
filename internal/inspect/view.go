package inspect

import (
	"encoding/json"
	"fmt"

	libghostty "go.mitchellh.com/libghostty"
)

type ViewResult struct {
	Content      string `json:"content"`
	CursorX      uint16 `json:"cursor_x"`
	CursorY      uint16 `json:"cursor_y"`
	CursorVisible bool  `json:"cursor_visible"`
	ActiveScreen string `json:"active_screen"`
}

func GetView(term *libghostty.Terminal, format string, trim, unwrap bool) (*ViewResult, error) {
	var fmtFormat libghostty.FormatterFormat
	switch format {
	case "plain":
		fmtFormat = libghostty.FormatterFormatPlain
	case "vt":
		fmtFormat = libghostty.FormatterFormatVT
	case "html":
		fmtFormat = libghostty.FormatterFormatHTML
	default:
		return nil, fmt.Errorf("unknown format %q", format)
	}

	f, err := libghostty.NewFormatter(term,
		libghostty.WithFormatterFormat(fmtFormat),
		libghostty.WithFormatterTrim(trim),
		libghostty.WithFormatterUnwrap(unwrap),
	)
	if err != nil {
		return nil, fmt.Errorf("create formatter: %w", err)
	}
	defer f.Close()

	content, err := f.FormatString()
	if err != nil {
		return nil, fmt.Errorf("format: %w", err)
	}

	cursorX, _ := term.CursorX()
	cursorY, _ := term.CursorY()
	cursorVisible, _ := term.CursorVisible()
	screen, _ := term.ActiveScreen()

	var screenStr string
	switch screen {
	case libghostty.ScreenPrimary:
		screenStr = "primary"
	case libghostty.ScreenAlternate:
		screenStr = "alternate"
	}

	return &ViewResult{
		Content:      content,
		CursorX:      cursorX,
		CursorY:      cursorY,
		CursorVisible: cursorVisible,
		ActiveScreen: screenStr,
	}, nil
}

func (v *ViewResult) ToJSON() string {
	data, _ := json.Marshal(v)
	return string(data)
}
