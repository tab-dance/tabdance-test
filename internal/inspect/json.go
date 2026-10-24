package inspect

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	libghostty "go.mitchellh.com/libghostty"
)

type JSONViewResult struct {
	Cols          uint16              `json:"cols"`
	Rows          uint16              `json:"rows"`
	CursorX       uint16              `json:"cursor_x"`
	CursorY       uint16              `json:"cursor_y"`
	CursorVisible bool                `json:"cursor_visible"`
	CursorStyle   string              `json:"cursor_style"`
	ActiveScreen  string              `json:"active_screen"`
	Cells         map[string]jsonCell `json:"cells,omitempty"`
}

type jsonCell struct {
	Text      string `json:"text,omitempty"`
	Codepoint uint32 `json:"codepoint,omitempty"`
	Wide      bool   `json:"wide,omitempty"`
	Bold      bool   `json:"bold,omitempty"`
	Italic    bool   `json:"italic,omitempty"`
	FgColor   string `json:"fg,omitempty"`
	BgColor   string `json:"bg,omitempty"`
	Inverse   bool   `json:"inverse,omitempty"`
	Hyperlink string `json:"hyperlink,omitempty"`
}

func RenderJSONGrid(term *libghostty.Terminal) (string, error) {
	result, err := BuildJSONView(term)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(result)
	if err != nil {
		return "", fmt.Errorf("marshal json: %w", err)
	}
	return string(data) + "\n", nil
}

func BuildJSONView(term *libghostty.Terminal) (*JSONViewResult, error) {
	cols, _ := term.Cols()
	rows, _ := term.Rows()
	if cols == 0 || rows == 0 {
		return nil, fmt.Errorf("terminal has zero dimensions")
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

	cursorStyleStr := cursorStyleString(term)

	reverseColors, _ := term.Mode(libghostty.ModeReverseColors)

	cells := make(map[string]jsonCell)
	for y := uint32(0); y < uint32(rows); y++ {
		for x := uint16(0); x < cols; x++ {
			point := libghostty.Point{
				Tag: libghostty.PointTagViewport,
				X:   x,
				Y:   y,
			}
			ref, err := term.GridRef(point)
			if err != nil {
				continue
			}
			c := collectJSONCell(ref, reverseColors)
			if !isDefaultCell(c) {
				cells[fmt.Sprintf("%d,%d", y, x)] = c
			}
		}
	}

	return &JSONViewResult{
		Cols:          cols,
		Rows:          rows,
		CursorX:       cursorX,
		CursorY:       cursorY,
		CursorVisible: cursorVisible,
		CursorStyle:   cursorStyleStr,
		ActiveScreen:  screenStr,
		Cells:         cells,
	}, nil
}

// isDefaultCell returns true if the cell has all default/zero values
// (empty text, default colors, no styling).
func isDefaultCell(c jsonCell) bool {
	return c.Text == "" && c.FgColor == "" && c.BgColor == "" &&
		!c.Wide && !c.Bold && !c.Italic && !c.Inverse && c.Hyperlink == ""
}

func collectJSONCell(ref *libghostty.GridRef, reverseColors bool) jsonCell {
	c := jsonCell{}

	var cell *libghostty.Cell
	cell, _ = ref.Cell()

	if cell != nil {
		cp, _ := cell.Codepoint()
		c.Codepoint = cp
		wide, _ := cell.Wide()
		c.Wide = wide == libghostty.CellWideWide

		if tag, err := cell.ContentTag(); err == nil {
			switch tag {
			case libghostty.CellContentBgColorPalette:
				if idx, err := cell.ColorPalette(); err == nil {
					c.BgColor = fmt.Sprintf("color%d", idx)
				}
			case libghostty.CellContentBgColorRGB:
				if rgb, err := cell.ColorRGB(); err == nil {
					c.BgColor = fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
				}
			}
		}
	}

	if graphemes, err := ref.Graphemes(); err == nil && len(graphemes) > 0 {
		runes := make([]rune, len(graphemes))
		for i, cp := range graphemes {
			runes[i] = rune(cp)
		}
		c.Text = string(runes)
	}

	if style, err := ref.Style(); err == nil && style != nil {
		c.Bold = style.Bold()
		c.Italic = style.Italic()
		c.Inverse = style.Inverse() != reverseColors
		c.FgColor = styleColorString(style.FgColor())
		if c.BgColor == "" {
			c.BgColor = styleColorString(style.BgColor())
		}
	}

	if c.FgColor == "default" {
		c.FgColor = ""
	}
	if c.BgColor == "default" {
		c.BgColor = ""
	}

	if uri, err := ref.HyperlinkURI(); err == nil && uri != "" {
		c.Hyperlink = uri
	}

	return c
}

// ExtractSVGGridJSON extracts the JSON grid from an SVG file's
// <script id="cell-grid"> element. Returns the raw JSON string.
func ExtractSVGGridJSON(svgContent string) (string, error) {
	pattern := regexp.MustCompile(
		`<script id="` + regexp.QuoteMeta(SVGGridJSONID) + `" type="application/json">` +
			`(.*?)` +
			`</script>`,
	)
	matches := pattern.FindStringSubmatch(svgContent)
	if len(matches) < 2 {
		return "", fmt.Errorf("no <%s> element found in SVG", SVGGridJSONID)
	}

	// Unescape XML entities that escapeXML may have added
	escaped := matches[1]
	escaped = strings.ReplaceAll(escaped, "&amp;", "&")
	escaped = strings.ReplaceAll(escaped, "&lt;", "<")
	escaped = strings.ReplaceAll(escaped, "&gt;", ">")
	escaped = strings.ReplaceAll(escaped, "&quot;", `"`)
	escaped = strings.ReplaceAll(escaped, "&apos;", "'")

	// Validate it's valid JSON by normalizing: parse and re-marshal
	// This also handles both sparse (map) and dense ([][]jsonCell) formats
	normalized, err := normalizeGridJSON(escaped)
	if err != nil {
		return "", fmt.Errorf("invalid JSON in <%s>: %w", SVGGridJSONID, err)
	}

	return normalized, nil
}

// normalizeGridJSON parses either sparse or dense grid JSON and returns
// a canonical sparse JSON string for comparison.
func normalizeGridJSON(raw string) (string, error) {
	// Try sparse format first (map[string]jsonCell)
	var sparse struct {
		Cols          uint16              `json:"cols"`
		Rows          uint16              `json:"rows"`
		CursorX       uint16              `json:"cursor_x"`
		CursorY       uint16              `json:"cursor_y"`
		CursorVisible bool                `json:"cursor_visible"`
		CursorStyle   string              `json:"cursor_style"`
		ActiveScreen  string              `json:"active_screen"`
		Cells         map[string]jsonCell `json:"cells"`
	}
	if err := json.Unmarshal([]byte(raw), &sparse); err == nil && sparse.Cells != nil {
		data, err := json.Marshal(sparse)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}

	// Try dense format ([][]jsonCell) for backward compatibility
	var dense struct {
		Cols          uint16       `json:"cols"`
		Rows          uint16       `json:"rows"`
		CursorX       uint16       `json:"cursor_x"`
		CursorY       uint16       `json:"cursor_y"`
		CursorVisible bool         `json:"cursor_visible"`
		CursorStyle   string       `json:"cursor_style"`
		ActiveScreen  string       `json:"active_screen"`
		Cells         [][]jsonCell `json:"cells"`
	}
	if err := json.Unmarshal([]byte(raw), &dense); err != nil {
		return "", err
	}

	// Convert dense to sparse
	cells := make(map[string]jsonCell)
	for y, row := range dense.Cells {
		for x, c := range row {
			if !isDefaultCell(c) {
				cells[fmt.Sprintf("%d,%d", y, x)] = c
			}
		}
	}

	result := struct {
		Cols          uint16              `json:"cols"`
		Rows          uint16              `json:"rows"`
		CursorX       uint16              `json:"cursor_x"`
		CursorY       uint16              `json:"cursor_y"`
		CursorVisible bool                `json:"cursor_visible"`
		CursorStyle   string              `json:"cursor_style"`
		ActiveScreen  string              `json:"active_screen"`
		Cells         map[string]jsonCell `json:"cells"`
	}{
		Cols:          dense.Cols,
		Rows:          dense.Rows,
		CursorX:       dense.CursorX,
		CursorY:       dense.CursorY,
		CursorVisible: dense.CursorVisible,
		CursorStyle:   dense.CursorStyle,
		ActiveScreen:  dense.ActiveScreen,
		Cells:         cells,
	}

	data, err := json.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
