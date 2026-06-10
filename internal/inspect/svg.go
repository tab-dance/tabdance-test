package inspect

import (
	"encoding/json"
	"fmt"
	"strings"

	libghostty "go.mitchellh.com/libghostty"
)

// svgRenderer holds state for building SVG output.
type svgRenderer struct {
	term          *libghostty.Terminal
	cols          uint16
	rows          uint16
	cellWidth     float64
	cellHeight    float64
	reverseColors bool
	colors        *libghostty.RenderStateColors
	buf           strings.Builder
}

// cellInfo holds all data needed for a cell tooltip.
type cellInfo struct {
	Text      string `json:"text"`
	Codepoint uint32 `json:"codepoint"`
	Wide      bool   `json:"wide"`
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	FgColor   string `json:"fg"`
	BgColor   string `json:"bg"`
	Hyperlink string `json:"hyperlink,omitempty"`
	X         uint16 `json:"x"`
	Y         uint32 `json:"y"`
}

// RenderSVG renders the terminal as an SVG with per-cell mouseover tooltips.
func RenderSVG(term *libghostty.Terminal, trim, unwrap bool) (string, error) {
	cols, _ := term.Cols()
	rows, _ := term.Rows()

	if cols == 0 || rows == 0 {
		return "", fmt.Errorf("terminal has zero dimensions")
	}

	widthPx, _ := term.WidthPx()
	heightPx, _ := term.HeightPx()

	if widthPx == 0 {
		widthPx = uint32(cols) * 8
	}
	if heightPx == 0 {
		heightPx = uint32(rows) * 16
	}

	cellWidth := float64(widthPx) / float64(cols)
	cellHeight := float64(heightPx) / float64(rows)

	reverseColors, _ := term.ModeGet(libghostty.ModeReverseColors)
	r := &svgRenderer{
		term:          term,
		cols:          cols,
		rows:          rows,
		cellWidth:     cellWidth,
		cellHeight:    cellHeight,
		reverseColors: reverseColors,
	}

	r.buildSVG(trim, unwrap)
	return r.buf.String(), nil
}

func (r *svgRenderer) buildSVG(trim, unwrap bool) {
	width := float64(r.cols) * r.cellWidth
	height := float64(r.rows) * r.cellHeight

	// Escape for use inside XML attribute JavaScript object literal.
	// Use &quot; for quotes (valid XML entity), and \ for backslash.
	esc := func(s string) string {
		s = strings.ReplaceAll(s, `\`, `\\`)
		s = strings.ReplaceAll(s, `"`, `&quot;`)
		s = strings.ReplaceAll(s, "'", `\'`)
		s = strings.ReplaceAll(s, "\n", `\n`)
		s = strings.ReplaceAll(s, "\r", `\r`)
		s = strings.ReplaceAll(s, "\t", `\t`)
		return s
	}

	r.buf.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="`)
	r.buf.WriteString(fmt.Sprintf("%.0f", width))
	r.buf.WriteString(`" height="`)
	r.buf.WriteString(fmt.Sprintf("%.0f", height))
	r.buf.WriteString(`" viewBox="0 0 `)
	r.buf.WriteString(fmt.Sprintf("%.0f", width))
	r.buf.WriteString(` `)
	r.buf.WriteString(fmt.Sprintf("%.0f", height))
	r.buf.WriteString(`">`)

	// Embedded styles with JetBrains Mono
	r.buf.WriteString(`<style>
		@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&amp;display=swap');
		text { font-family: 'JetBrains Mono', monospace; font-size: `)
	r.buf.WriteString(fmt.Sprintf("%.0fpx", r.cellHeight*0.8))
	r.buf.WriteString(`; dominant-baseline: middle; }
		.cell-rect { cursor: crosshair; }
	</style>`)

	// Cursor, selection, and color data
	cursorX, _ := r.term.CursorX()
	cursorY, _ := r.term.CursorY()
	cursorVisible, _ := r.term.CursorVisible()
	cursorStyle := "block"
	sel, _ := r.term.Selection()
	rs, _ := libghostty.NewRenderState()
	if rs != nil {
		defer rs.Close()
		if err := rs.Update(r.term); err == nil {
			colors, _ := rs.Colors()
			r.colors = colors
			if cs, err := rs.CursorVisualStyle(); err == nil {
				switch cs {
				case libghostty.CursorVisualStyleBlock:
					cursorStyle = "block"
				case libghostty.CursorVisualStyleUnderline:
					cursorStyle = "underline"
				case libghostty.CursorVisualStyleBlockHollow:
					cursorStyle = "block_hollow"
				default:
					cursorStyle = "bar"
				}
			}
		}
	}

	// Background
	r.buf.WriteString(`<rect width="100%" height="100%" fill="`)
	r.buf.WriteString(r.bgFill("default"))
	r.buf.WriteString(`"/>`)

	

	for y := uint32(0); y < uint32(r.rows); y++ {
		for x := uint16(0); x < r.cols; x++ {
			point := libghostty.Point{
				Tag: libghostty.PointTagViewport,
				X:   x,
				Y:   y,
			}

			ref, err := r.term.GridRef(point)
			if err != nil {
				continue
			}

			ci := r.collectCellInfo(ref, x, y)

			cx := float64(x) * r.cellWidth
			cy := float64(y) * r.cellHeight

			cellW := r.cellWidth
			if ci.Wide {
				cellW = r.cellWidth * 2
			}

			// Escape text for SVG tspan content
			textEscaped := r.escapeXML(ci.Text)
			fgEscaped := r.fgColor(ci.FgColor)
			bgFillEsc := r.bgFill(ci.BgColor)

			// Build tooltip data as JSON
			tooltip := map[string]interface{}{
				"x":          ci.X,
				"y":          ci.Y,
				"text":       ci.Text,
				"codepoint":  ci.Codepoint,
				"wide":       ci.Wide,
				"bold":       ci.Bold,
				"italic":     ci.Italic,
				"fg":         ci.FgColor,
				"bg":         ci.BgColor,
				"hyperlink":  ci.Hyperlink,
			}
			tooltipJSON, _ := json.Marshal(tooltip)
			tooltipEsc := esc(string(tooltipJSON))

			// Hover group
			isSelected := false
			if sel != nil {
				if selected, _ := r.term.SelectionContains(sel, point); selected {
					isSelected = true
				}
			}
			var invertStyle string
			if isSelected {
				invertStyle = ` style="filter: invert(100%)"`
			}
			r.buf.WriteString(`<g class="cell-rect" onmouseover="showTooltip(evt,`)
			r.buf.WriteString(tooltipEsc)
			r.buf.WriteString(`)" onmouseout="hideTooltip()"`)
			r.buf.WriteString(invertStyle)
			r.buf.WriteString(`>`)

			// Background rect
			r.buf.WriteString(`<rect x="`)
			r.buf.WriteString(fmt.Sprintf("%.2f", cx))
			r.buf.WriteString(`" y="`)
			r.buf.WriteString(fmt.Sprintf("%.2f", cy))
			r.buf.WriteString(`" width="`)
			r.buf.WriteString(fmt.Sprintf("%.2f", cellW))
			r.buf.WriteString(`" height="`)
			r.buf.WriteString(fmt.Sprintf("%.2f", r.cellHeight))
			r.buf.WriteString(`" fill="`)
			r.buf.WriteString(bgFillEsc)
			r.buf.WriteString(`"/>`)

			// Text (skip empty cells)
			if ci.Text != "" {
				r.buf.WriteString(`<text x="`)
				r.buf.WriteString(fmt.Sprintf("%.2f", cx))
				r.buf.WriteString(`" y="`)
				r.buf.WriteString(fmt.Sprintf("%.2f", cy+r.cellHeight/2))
				r.buf.WriteString(`" fill="`)
				r.buf.WriteString(fgEscaped)
				r.buf.WriteString(`"`)
				if ci.Bold {
					r.buf.WriteString(` font-weight="bold"`)
				}
				if ci.Italic {
					r.buf.WriteString(` font-style="italic"`)
				}
				r.buf.WriteString(`>`)
				r.buf.WriteString(`<tspan x="`)
				r.buf.WriteString(fmt.Sprintf("%.2f", cx))
				r.buf.WriteString(`">`)
				r.buf.WriteString(textEscaped)
				r.buf.WriteString(`</tspan></text>`)
			}

			// Cursor overlay
			if cursorVisible && ci.X == cursorX && int(ci.Y) == int(cursorY) {
				cursorFill := "#ffffff"
				cursorOpacity := "0.9"
				if cursorStyle == "underline" {
					r.buf.WriteString(`<rect x="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", cx))
					r.buf.WriteString(`" y="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", cy+r.cellHeight-2))
					r.buf.WriteString(`" width="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", r.cellWidth))
					r.buf.WriteString(`" height="2" fill="`)
					r.buf.WriteString(cursorFill)
					r.buf.WriteString(`" opacity="`)
					r.buf.WriteString(cursorOpacity)
					r.buf.WriteString(`"/>`)
				} else if cursorStyle == "block_hollow" {
					r.buf.WriteString(`<rect x="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", cx))
					r.buf.WriteString(`" y="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", cy))
					r.buf.WriteString(`" width="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", r.cellWidth))
					r.buf.WriteString(`" height="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", r.cellHeight))
					r.buf.WriteString(`" fill="none" stroke="`)
					r.buf.WriteString(cursorFill)
					r.buf.WriteString(`" stroke-width="1" opacity="`)
					r.buf.WriteString(cursorOpacity)
					r.buf.WriteString(`"/>`)
				} else {
					// block (default) - fill with reversed colors
					r.buf.WriteString(`<rect x="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", cx))
					r.buf.WriteString(`" y="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", cy))
					r.buf.WriteString(`" width="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", r.cellWidth))
					r.buf.WriteString(`" height="`)
					r.buf.WriteString(fmt.Sprintf("%.2f", r.cellHeight))
					r.buf.WriteString(`" fill="`)
					r.buf.WriteString(fgEscaped)
					r.buf.WriteString(`" opacity="`)
					r.buf.WriteString(cursorOpacity)
					r.buf.WriteString(`"/>`)
					// Draw text on top with bg color for visibility
					if ci.Text != "" {
						r.buf.WriteString(`<text x="`)
						r.buf.WriteString(fmt.Sprintf("%.2f", cx))
						r.buf.WriteString(`" y="`)
						r.buf.WriteString(fmt.Sprintf("%.2f", cy+r.cellHeight/2))
						r.buf.WriteString(`" fill="`)
						r.buf.WriteString(bgFillEsc)
						r.buf.WriteString(`"`)
						if ci.Bold {
							r.buf.WriteString(` font-weight="bold"`)
						}
						if ci.Italic {
							r.buf.WriteString(` font-style="italic"`)
						}
						r.buf.WriteString(`>`)
						r.buf.WriteString(`<tspan x="`)
						r.buf.WriteString(fmt.Sprintf("%.2f", cx))
						r.buf.WriteString(`">`)
						r.buf.WriteString(textEscaped)
						r.buf.WriteString(`</tspan></text>`)
					}
				}
			}

			r.buf.WriteString(`</g>`)
		}
	}

	// Tooltip overlay div (foreignObject for HTML tooltip)
	r.buf.WriteString(`<foreignObject x="0" y="0" width="100%" height="100%" style="pointer-events:none;">
	<style>
		#cell-tooltip {
			position: fixed;
			z-index: 99999;
			background: #1e1e1e;
			color: #d4d4d4;
			border: 1px solid #555;
			border-radius: 4px;
			padding: 8px 12px;
			font-family: 'JetBrains Mono', monospace;
			font-size: 12px;
			max-width: 300px;
			box-shadow: 0 4px 12px rgba(0,0,0,0.4);
			display: none;
			pointer-events: none;
			line-height: 1.6;
		}
		#cell-tooltip .row { display: flex; gap: 8px; }
		#cell-tooltip .key { color: #6a9955; min-width: 80px; }
		#cell-tooltip .val { color: #ce9178; word-break: break-all; }
	</style>
	<div xmlns="http://www.w3.org/1999/xhtml">
		<div id="cell-tooltip">
			<div class="row"><span class="key">Position:</span><span class="val" id="tt-pos"></span></div>
			<div class="row"><span class="key">Text:</span><span class="val" id="tt-text"></span></div>
			<div class="row"><span class="key">Codepoint:</span><span class="val" id="tt-cp"></span></div>
			<div class="row"><span class="key">Wide:</span><span class="val" id="tt-wide"></span></div>
			<div class="row"><span class="key">Bold:</span><span class="val" id="tt-bold"></span></div>
			<div class="row"><span class="key">Italic:</span><span class="val" id="tt-italic"></span></div>
			<div class="row"><span class="key">FG:</span><span class="val" id="tt-fg"></span></div>
			<div class="row"><span class="key">BG:</span><span class="val" id="tt-bg"></span></div>
			<div class="row" id="tt-hl-row" style="display:none"><span class="key">Link:</span><span class="val" id="tt-hl"></span></div>
		</div>
	</div>
	</foreignObject>`)

	// JavaScript for tooltip
	r.buf.WriteString(`<script type="text/javascript"><![CDATA[
	var tooltip = document.getElementById('cell-tooltip');
	function showTooltip(evt, data) {
		var svg = evt.target.closest('svg');
		var rect = svg.getBoundingClientRect();
		var tw = tooltip.offsetWidth || 300;
		var th = tooltip.offsetHeight || 200;
		// Position based on cell location: opposite corner
		var x = data.x * 8;
		var y = data.y * 16;
		// If in top half, show below; if in bottom half, show above
		if (y < rect.height / 2) y += 20;
		else y -= th + 20;
		// If in left half, show to the right; if in right half, show to the left
		if (x < rect.width / 2) x += 20;
		else x -= tw + 20;
		// Clamp to SVG bounds
		x = Math.max(10, Math.min(x, rect.width - tw - 10));
		y = Math.max(10, Math.min(y, rect.height - th - 10));
		tooltip.style.display = 'block';
		tooltip.style.left = x + 'px';
		tooltip.style.top = y + 'px';
		document.getElementById('tt-pos').textContent = data.x + ',' + data.y;
		document.getElementById('tt-text').textContent = JSON.stringify(data.text);
		document.getElementById('tt-cp').textContent = data.codepoint + ' (0x' + data.codepoint.toString(16).toUpperCase() + ')';
		document.getElementById('tt-wide').textContent = String(data.wide);
		document.getElementById('tt-bold').textContent = String(data.bold);
		document.getElementById('tt-italic').textContent = String(data.italic);
		document.getElementById('tt-fg').textContent = data.fg;
		document.getElementById('tt-bg').textContent = data.bg;
		if (data.hyperlink) {
			document.getElementById('tt-hl-row').style.display = 'flex';
			document.getElementById('tt-hl').textContent = data.hyperlink;
		} else {
			document.getElementById('tt-hl-row').style.display = 'none';
		}
	}
	function hideTooltip() {
		tooltip.style.display = 'none';
	}
	]]></script>`)

	r.buf.WriteString(`</svg>`)
}

func (r *svgRenderer) collectCellInfo(ref *libghostty.GridRef, x uint16, y uint32) *cellInfo {
	ci := &cellInfo{X: x, Y: y}

	var cell *libghostty.Cell
	cell, _ = ref.Cell()

	if cell != nil {
		cp, _ := cell.Codepoint()
		ci.Codepoint = cp
		wide, _ := cell.Wide()
		ci.Wide = wide == libghostty.CellWideWide

		// Check for inline background color on erased/empty cells.
		// When a cell's ContentTag is BgColorPalette or BgColorRGB,
		// the background color is stored directly in the cell (not in the style map).
		if tag, err := cell.ContentTag(); err == nil {
			switch tag {
			case libghostty.CellContentBgColorPalette:
				if idx, err := cell.ColorPalette(); err == nil {
					ci.BgColor = r.resolvePaletteColor(idx)
				}
			case libghostty.CellContentBgColorRGB:
				if rgb, err := cell.ColorRGB(); err == nil {
					ci.BgColor = fmt.Sprintf("#%02x%02x%02x", rgb.R, rgb.G, rgb.B)
				}
			}
		}
	}

	if graphemes, err := ref.Graphemes(); err == nil && len(graphemes) > 0 {
		runes := make([]rune, len(graphemes))
		for i, cp := range graphemes {
			runes[i] = rune(cp)
		}
		ci.Text = string(runes)
	}

	if style, err := ref.Style(); err == nil && style != nil {
		ci.Bold = style.Bold()
		ci.Italic = style.Italic()
		inverse := style.Inverse() != r.reverseColors
		if inverse {
			fg := r.resolveStyleColor(style.FgColor(), true)
			bg := r.resolveStyleColor(style.BgColor(), false)
			if ci.BgColor != "" {
				// ContentTag-derived bg: swap it with the resolved fg
				ci.FgColor = ci.BgColor
				ci.BgColor = fg
			} else {
				ci.FgColor = bg
				ci.BgColor = fg
			}
		} else {
			ci.FgColor = styleColorString(style.FgColor())
			if ci.BgColor == "" {
				ci.BgColor = styleColorString(style.BgColor())
			}
		}
	}

	if ci.FgColor == "" {
		ci.FgColor = "default"
	}
	if ci.BgColor == "" {
		ci.BgColor = "default"
	}

	if uri, err := ref.HyperlinkURI(); err == nil && uri != "" {
		ci.Hyperlink = uri
	}

	return ci
}

// fgColor returns SVG fill color for foreground.
func (r *svgRenderer) fgColor(fg string) string {
	switch fg {
	case "default":
		if r.colors != nil {
			return fmt.Sprintf("#%02x%02x%02x", r.colors.Foreground.R, r.colors.Foreground.G, r.colors.Foreground.B)
		}
		return "#d4d4d4"
	default:
		return fg
	}
}

// bgFill returns SVG fill color for background.
func (r *svgRenderer) bgFill(bg string) string {
	switch bg {
	case "default":
		if r.colors != nil {
			return fmt.Sprintf("#%02x%02x%02x", r.colors.Background.R, r.colors.Background.G, r.colors.Background.B)
		}
		return "#171717"
	default:
		return bg
	}
}

// resolveStyleColor resolves a StyleColor to a hex color string using the
// terminal's palette and default colors. The isFg parameter determines which
// default color to use as fallback for StyleColorNone.
func (r *svgRenderer) resolveStyleColor(sc libghostty.StyleColor, isFg bool) string {
	switch sc.Tag {
	case libghostty.StyleColorRGB:
		return fmt.Sprintf("#%02x%02x%02x", sc.RGB.R, sc.RGB.G, sc.RGB.B)
	case libghostty.StyleColorPalette:
		if r.colors != nil && int(sc.Palette) < len(r.colors.Palette) {
			c := r.colors.Palette[sc.Palette]
			return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
		}
		return fmt.Sprintf("color%d", sc.Palette)
	default:
		if r.colors != nil {
			if isFg {
				return fmt.Sprintf("#%02x%02x%02x", r.colors.Foreground.R, r.colors.Foreground.G, r.colors.Foreground.B)
			}
			return fmt.Sprintf("#%02x%02x%02x", r.colors.Background.R, r.colors.Background.G, r.colors.Background.B)
		}
		if isFg {
			return "#d4d4d4"
		}
		return "#171717"
	}
}

// resolvePaletteColor resolves a palette index to a hex color string.
func (r *svgRenderer) resolvePaletteColor(idx uint8) string {
	if r.colors != nil && int(idx) < len(r.colors.Palette) {
		c := r.colors.Palette[idx]
		return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
	}
	return fmt.Sprintf("color%d", idx)
}

// escapeXML escapes characters that are illegal in XML text content.
func (r *svgRenderer) escapeXML(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, `"`, "&quot;")
	s = strings.ReplaceAll(s, "'", "&apos;")
	return s
}