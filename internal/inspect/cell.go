package inspect

import (
	"encoding/json"
	"fmt"

	libghostty "go.mitchellh.com/libghostty"
)

type CellResult struct {
	Codepoint uint32 `json:"codepoint"`
	Text      string `json:"text"`
	Wide      bool   `json:"wide"`
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	FgColor   string `json:"fg"`
	BgColor   string `json:"bg"`
	Inverse   bool   `json:"inverse"`
	Hyperlink string `json:"hyperlink,omitempty"`
}

func GetCell(term *libghostty.Terminal, x uint16, y uint32) (*CellResult, error) {
	point := libghostty.Point{
		Tag: libghostty.PointTagViewport,
		X:   x,
		Y:   y,
	}

	ref, err := term.GridRef(point)
	if err != nil {
		return nil, fmt.Errorf("grid ref: %w", err)
	}

	result := &CellResult{}

	cell, err := ref.Cell()
	if err == nil {
		cp, _ := cell.Codepoint()
		result.Codepoint = cp

		wide, _ := cell.Wide()
		result.Wide = wide == libghostty.CellWideWide

		if tag, err := cell.ContentTag(); err == nil {
			switch tag {
			case libghostty.CellContentBgColorPalette:
				if idx, err := cell.ColorPalette(); err == nil {
					result.BgColor = fmt.Sprintf("color%d", idx)
				}
			case libghostty.CellContentBgColorRGB:
				if rgb, err := cell.ColorRGB(); err == nil {
					result.BgColor = fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
				}
			}
		}
	}

	if graphemes, err := ref.Graphemes(); err == nil && len(graphemes) > 0 {
		runes := make([]rune, len(graphemes))
		for i, cp := range graphemes {
			runes[i] = rune(cp)
		}
		result.Text = string(runes)
	}

	reverseColors, _ := term.Mode(libghostty.ModeReverseColors)

	if style, err := ref.Style(); err == nil && style != nil {
		result.Bold = style.Bold()
		result.Italic = style.Italic()
		result.Inverse = style.Inverse() != reverseColors
		result.FgColor = styleColorString(style.FgColor())
		if result.BgColor == "" {
			result.BgColor = styleColorString(style.BgColor())
		}
	}

	if uri, err := ref.HyperlinkURI(); err == nil && uri != "" {
		result.Hyperlink = uri
	}

	return result, nil
}

func (c *CellResult) ToJSON() string {
	data, _ := json.Marshal(c)
	return string(data)
}

func styleColorString(sc libghostty.StyleColor) string {
	switch sc.Tag {
	case libghostty.StyleColorNone:
		return "default"
	case libghostty.StyleColorPalette:
		return fmt.Sprintf("color%d", sc.Palette)
	case libghostty.StyleColorRGB:
		return fmt.Sprintf("#%02x%02x%02x", sc.RGB.R, sc.RGB.G, sc.RGB.B)
	default:
		return "default"
	}
}
