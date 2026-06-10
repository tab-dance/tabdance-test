package inspect

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	libghostty "go.mitchellh.com/libghostty"
)

type JSONViewResult struct {
	Cols          uint16      `json:"cols"`
	Rows          uint16      `json:"rows"`
	CursorX       uint16      `json:"cursor_x"`
	CursorY       uint16      `json:"cursor_y"`
	CursorVisible bool        `json:"cursor_visible"`
	CursorStyle   string      `json:"cursor_style"`
	ActiveScreen  string      `json:"active_screen"`
	Cells         [][]jsonCell `json:"cells"`
}

type jsonCell struct {
	Text      string `json:"text"`
	Codepoint uint32 `json:"codepoint"`
	Wide      bool   `json:"wide"`
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	FgColor   string `json:"fg"`
	BgColor   string `json:"bg"`
	Inverse   bool   `json:"inverse"`
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

	cursorStyleStr := "bar"
	rs, _ := libghostty.NewRenderState()
	if rs != nil {
		defer rs.Close()
		if err := rs.Update(term); err == nil {
			if cs, err := rs.CursorVisualStyle(); err == nil {
				switch cs {
				case libghostty.CursorVisualStyleBlock:
					cursorStyleStr = "block"
				case libghostty.CursorVisualStyleUnderline:
					cursorStyleStr = "underline"
				case libghostty.CursorVisualStyleBlockHollow:
					cursorStyleStr = "block_hollow"
				default:
					cursorStyleStr = "bar"
				}
			}
		}
	}

	reverseColors, _ := term.ModeGet(libghostty.ModeReverseColors)

	cells := make([][]jsonCell, rows)
	for y := uint32(0); y < uint32(rows); y++ {
		row := make([]jsonCell, cols)
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
			row[x] = collectJSONCell(ref, x, y, term, reverseColors)
		}
		cells[y] = row
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

func collectJSONCell(ref *libghostty.GridRef, _ uint16, _ uint32, _ *libghostty.Terminal, reverseColors bool) jsonCell {
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

	if c.FgColor == "" {
		c.FgColor = "default"
	}
	if c.BgColor == "" {
		c.BgColor = "default"
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

	// Validate it's valid JSON
	var parsed JSONViewResult
	if err := json.Unmarshal([]byte(escaped), &parsed); err != nil {
		return "", fmt.Errorf("invalid JSON in <%s>: %w", SVGGridJSONID, err)
	}

	return escaped, nil
}
