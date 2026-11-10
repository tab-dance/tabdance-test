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

// RenderContent renders the terminal screen in the given format. It is the
// single rendering primitive shared by view, snapshot/verify and wait, so
// every surface formats a screen identically.
func RenderContent(term *libghostty.Terminal, format string, trim, unwrap bool) (string, error) {
	switch format {
	case "plain":
		return formatTerminal(term, libghostty.FormatterFormatPlain, trim, unwrap)
	case "vt":
		return formatTerminal(term, libghostty.FormatterFormatVT, trim, unwrap)
	case "html":
		return formatTerminal(term, libghostty.FormatterFormatHTML, trim, unwrap)
	case "svg":
		return RenderSVG(term, trim, unwrap)
	case "json":
		return RenderJSONGrid(term)
	default:
		return "", fmt.Errorf("unknown format %q", format)
	}
}

func GetView(term *libghostty.Terminal, format string, trim, unwrap bool) (*ViewResult, error) {
	content, err := RenderContent(term, format, trim, unwrap)
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

	styleStr := cursorStyleString(term)

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
