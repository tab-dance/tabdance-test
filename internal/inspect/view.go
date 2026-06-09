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
	OutputPath    string `json:"output_path,omitempty"`
}

func GetView(term *libghostty.Terminal, format string, trim, unwrap bool, outputPath string) (*ViewResult, error) {
	if isFreezeFormat(format) {
		return getFreezeView(term, format, trim, unwrap, outputPath)
	}

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

	return newViewResult(term, content, ""), nil
}

func getFreezeView(term *libghostty.Terminal, format string, trim, unwrap bool, outputPath string) (*ViewResult, error) {
	f, err := libghostty.NewFormatter(term,
		libghostty.WithFormatterFormat(libghostty.FormatterFormatVT),
		libghostty.WithFormatterTrim(trim),
		libghostty.WithFormatterUnwrap(unwrap),
	)
	if err != nil {
		return nil, fmt.Errorf("create formatter: %w", err)
	}
	defer f.Close()

	vtContent, err := f.FormatString()
	if err != nil {
		return nil, fmt.Errorf("format: %w", err)
	}

	content, err := renderFreeze(vtContent, format, outputPath)
	if err != nil {
		return nil, err
	}

	result := newViewResult(term, content, "")
	if format != "svg" {
		result.OutputPath = content
		result.Content = ""
	}
	return result, nil
}

func newViewResult(term *libghostty.Terminal, content, outputPath string) *ViewResult {
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
		ActiveScreen:   screenStr,
		OutputPath:    outputPath,
	}
}

func (v *ViewResult) ToJSON() string {
	data, _ := json.Marshal(v)
	return string(data)
}
