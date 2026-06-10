package inspect

import (
	"encoding/json"
	"fmt"

	libghostty "go.mitchellh.com/libghostty"
)

type ViewResult struct {
	Content       string `json:"content"`
	CursorX       uint16 `json:"cursor_x"`
	CursorY       uint16 `json:"cursor_y"`
	CursorVisible bool   `json:"cursor_visible"`
	CursorStyle   string `json:"cursor_style"`
	ActiveScreen  string `json:"active_screen"`
}

func GetView(term *libghostty.Terminal, format string, trim, unwrap bool) (*ViewResult, error) {
	var fmtFormat libghostty.FormatterFormat
	var content string
	var err error

	switch format {
	case "plain":
		fmtFormat = libghostty.FormatterFormatPlain
		content, err = formatTerminal(term, fmtFormat, trim, unwrap)
	case "vt":
		fmtFormat = libghostty.FormatterFormatVT
		content, err = formatTerminal(term, fmtFormat, trim, unwrap)
	case "html":
		fmtFormat = libghostty.FormatterFormatHTML
		content, err = formatTerminal(term, fmtFormat, trim, unwrap)
	case "svg":
		content, err = RenderSVG(term, trim, unwrap)
	case "json":
		content, err = RenderJSONGrid(term)
	default:
		return nil, fmt.Errorf("unknown format %q", format)
	}

	if err != nil {
		return nil, err
	}

	if format == "svg" || format == "json" {
		return &ViewResult{Content: content}, nil
	}

	return buildViewResult(term, content)
}

func formatTerminal(term *libghostty.Terminal, fmtFormat libghostty.FormatterFormat, trim, unwrap bool) (string, error) {
	f, err := libghostty.NewFormatter(term,
		libghostty.WithFormatterFormat(fmtFormat),
		libghostty.WithFormatterTrim(trim),
		libghostty.WithFormatterUnwrap(unwrap),
	)
	if err != nil {
		return "", fmt.Errorf("create formatter: %w", err)
	}
	defer f.Close()

	content, err := f.FormatString()
	if err != nil {
		return "", fmt.Errorf("format: %w", err)
	}
	return content, nil
}

func buildViewResult(term *libghostty.Terminal, content string) (*ViewResult, error) {
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

	styleStr := "bar"
	if cs, err := term.CursorStyle(); err == nil && cs != nil {
		if cs.Bold() {
			styleStr = "block"
		}
	}

	return &ViewResult{
		Content:       content,
		CursorX:       cursorX,
		CursorY:       cursorY,
		CursorVisible: cursorVisible,
		CursorStyle:   styleStr,
		ActiveScreen:  screenStr,
	}, nil
}

func (v *ViewResult) ToJSON() string {
	data, _ := json.Marshal(v)
	return string(data)
}
