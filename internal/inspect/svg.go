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
	intCoords     bool
}

// cellInfo holds all data needed for a cell tooltip and golden file verification.
type cellInfo struct {
	Text      string `json:"text"`
	Codepoint uint32 `json:"codepoint"`
	Wide      bool   `json:"wide"`
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	FgColor   string `json:"fg"`
	BgColor   string `json:"bg"`
	Inverse   bool   `json:"inverse"`
	Hyperlink string `json:"hyperlink,omitempty"`
	X         uint16 `json:"x"`
	Y         uint32 `json:"y"`
}

// SVGGridJSONID is the id of the <script> element containing the cell grid JSON.
const SVGGridJSONID = "cell-grid"

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
		intCoords:     cellWidth == float64(int(cellWidth)) && cellHeight == float64(int(cellHeight)),
	}

	r.buildSVG(trim, unwrap)
	return r.buf.String(), nil
}

// coord formats a float coordinate, using integer format when possible.
func (r *svgRenderer) coord(v float64) string {
	if r.intCoords {
		return fmt.Sprintf("%d", int(v))
	}
	return fmt.Sprintf("%.2f", v)
}

func (r *svgRenderer) buildSVG(trim, unwrap bool) {
	width := float64(r.cols) * r.cellWidth
	height := float64(r.rows) * r.cellHeight

	fmt.Fprintf(&r.buf, `<svg xmlns="http://www.w3.org/2000/svg" width="%s" height="%s" viewBox="0 0 %s %s" onmousemove="_sh(evt)" onmouseout="hideTooltip()" style="cursor:crosshair">`,
		r.coord(width), r.coord(height), r.coord(width), r.coord(height))

	// Embedded styles with JetBrains Mono
	fmt.Fprintf(&r.buf, `<style>
		@import url('https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;700&amp;display=swap');
		text{font-family:'JetBrains Mono',monospace;font-size:%spx;dominant-baseline:middle}
		.b{font-weight:bold}
		.i{font-style:italic}
		.s{filter:invert(100%%)}
	</style>`, r.coord(r.cellHeight*0.8))

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
	defaultBg := r.bgFill("default")
	fmt.Fprintf(&r.buf, `<rect width="100%%" height="100%%" fill="%s"/>`, defaultBg)

	// Define a reusable default-bg rect for empty cells
	fmt.Fprintf(&r.buf, `<defs><rect id="d" width="%s" height="%s" fill="%s"/></defs>`,
		r.coord(r.cellWidth), r.coord(r.cellHeight), defaultBg)

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

			// Skip empty cells (default bg, no text, not selected, not cursor)
			isSelected := false
			if sel != nil {
				if selected, _ := r.term.SelectionContains(sel, point); selected {
					isSelected = true
				}
			}
			isCursor := cursorVisible && ci.X == cursorX && int(ci.Y) == int(cursorY)
			isEmpty := ci.Text == "" && ci.BgColor == "default" && !isSelected && !isCursor

			if isEmpty {
				continue
			}

			textEscaped := r.escapeXML(ci.Text)
			fgEscaped := r.fgColor(ci.FgColor)
			bgFillEsc := r.bgFill(ci.BgColor)

			// Hover group — tooltip data looked up from embedded grid by (x,y)
			gCls := ""
			if isSelected {
				gCls = ` class="s"`
			}
			if gCls == "" {
				r.buf.WriteString(`<g>`)
			} else {
				r.buf.WriteString(`<g class="s">`)
			}

			// Background rect — use <use> for default bg, full <rect> for custom bg
			if ci.BgColor == "default" {
				fmt.Fprintf(&r.buf, `<use href="#d" x="%s" y="%s"/>`, r.coord(cx), r.coord(cy))
			} else {
				fmt.Fprintf(&r.buf, `<rect x="%s" y="%s" width="%s" height="%s" fill="%s"/>`,
					r.coord(cx), r.coord(cy), r.coord(cellW), r.coord(r.cellHeight), bgFillEsc)
			}

			// Text (skip empty cells)
			if ci.Text != "" {
				var cls string
				if ci.Bold && ci.Italic {
					cls = ` class="b i"`
				} else if ci.Bold {
					cls = ` class="b"`
				} else if ci.Italic {
					cls = ` class="i"`
				}
				fmt.Fprintf(&r.buf, `<text x="%s" y="%s" fill="%s"%s`, r.coord(cx), r.coord(cy+r.cellHeight/2), fgEscaped, cls)
				fmt.Fprintf(&r.buf, `><tspan x="%s">%s</tspan></text>`, r.coord(cx), textEscaped)
			}

			// Cursor overlay
			if isCursor {
				cursorFill := "#fff"
				cursorOpacity := "0.9"
				switch cursorStyle {
				case "underline":
					fmt.Fprintf(&r.buf, `<rect x="%s" y="%s" width="%s" height="2" fill="%s" opacity="%s"/>`,
						r.coord(cx), r.coord(cy+r.cellHeight-2), r.coord(r.cellWidth), cursorFill, cursorOpacity)
				case "block_hollow":
					fmt.Fprintf(&r.buf, `<rect x="%s" y="%s" width="%s" height="%s" fill="none" stroke="%s" stroke-width="1" opacity="%s"/>`,
						r.coord(cx), r.coord(cy), r.coord(r.cellWidth), r.coord(r.cellHeight), cursorFill, cursorOpacity)
				default:
					// block — fill with fg color, redraw text with bg color
					fmt.Fprintf(&r.buf, `<rect x="%s" y="%s" width="%s" height="%s" fill="%s" opacity="%s"/>`,
						r.coord(cx), r.coord(cy), r.coord(r.cellWidth), r.coord(r.cellHeight), fgEscaped, cursorOpacity)
					if ci.Text != "" {
						var cls string
						if ci.Bold && ci.Italic {
							cls = ` class="b i"`
						} else if ci.Bold {
							cls = ` class="b"`
						} else if ci.Italic {
							cls = ` class="i"`
						}
						fmt.Fprintf(&r.buf, `<text x="%s" y="%s" fill="%s"%s`, r.coord(cx), r.coord(cy+r.cellHeight/2), bgFillEsc, cls)
						fmt.Fprintf(&r.buf, `><tspan x="%s">%s</tspan></text>`, r.coord(cx), textEscaped)
					}
				}
			}

			r.buf.WriteString(`</g>`)
		}
	}

	// Tooltip overlay div (foreignObject for HTML tooltip)
	r.buf.WriteString(`<foreignObject x="0" y="0" width="100%" height="100%" style="pointer-events:none;">
	<style>
		#ct{position:fixed;z-index:99999;background:#1e1e1e;color:#d4d4d4;border:1px solid #555;border-radius:4px;padding:8px 12px;font-family:'JetBrains Mono',monospace;font-size:12px;max-width:300px;box-shadow:0 4px 12px rgba(0,0,0,.4);display:none;pointer-events:none;line-height:1.6}
		#ct .r{display:flex;gap:8px}
		#ct .k{color:#6a9955;min-width:80px}
		#ct .v{color:#ce9178;word-break:break-all}
	</style>
	<div xmlns="http://www.w3.org/1999/xhtml">
		<div id="ct">
			<div class="r"><span class="k">Position:</span><span class="v" id="tp"></span></div>
			<div class="r"><span class="k">Text:</span><span class="v" id="tt"></span></div>
			<div class="r"><span class="k">Codepoint:</span><span class="v" id="tc"></span></div>
			<div class="r"><span class="k">Wide:</span><span class="v" id="tw"></span></div>
			<div class="r"><span class="k">Bold:</span><span class="v" id="tb"></span></div>
			<div class="r"><span class="k">Italic:</span><span class="v" id="ti"></span></div>
			<div class="r"><span class="k">FG:</span><span class="v" id="tf"></span></div>
			<div class="r"><span class="k">BG:</span><span class="v" id="tg"></span></div>
			<div class="r"><span class="k">Inverse:</span><span class="v" id="tv"></span></div>
			<div class="r" id="th" style="display:none"><span class="k">Link:</span><span class="v" id="tl"></span></div>
		</div>
	</div>
	</foreignObject>`)

	// JavaScript for tooltip — computes cell position from mouse coordinates
	r.buf.WriteString(`<script type="text/javascript"><![CDATA[
	var _g=null,_cx=-1,_cy=-1,_cw=`)
	r.buf.WriteString(r.coord(r.cellWidth))
	r.buf.WriteString(`,_ch=`)
	r.buf.WriteString(r.coord(r.cellHeight))
	r.buf.WriteString(`;
	function _gd(){if(!_g){var e=document.getElementById('cell-grid');if(e){_g=JSON.parse(e.textContent)}}return _g}
	var _t=document.getElementById('ct');
	function _sh(evt){
		var svg=evt.target.closest('svg');
		var rect=svg.getBoundingClientRect();
		var mx=evt.clientX-rect.left,my=evt.clientY-rect.top;
		var x=Math.floor(mx/_cw),y=Math.floor(my/_ch);
		if(x===_cx&&y===_cy)return;
		_cx=x;_cy=y;
		if(x<0||y<0)return;
		var d=_gd();if(!d)return;
		if(x>=d.cols||y>=d.rows)return;
		var c=d.cells[y+','+x]||{text:'',codepoint:0,fg:'',bg:''};
		var tw=_t.offsetWidth||300;
		var th=_t.offsetHeight||200;
		var px=x*_cw;var py=y*_ch;
		if(py<rect.height/2)py+=20;else py-=th+20;
		if(px<rect.width/2)px+=20;else px-=tw+20;
		px=Math.max(10,Math.min(px,rect.width-tw-10));
		py=Math.max(10,Math.min(py,rect.height-th-10));
		_t.style.display='block';
		_t.style.left=px+'px';
		_t.style.top=py+'px';
		document.getElementById('tp').textContent=x+','+y;
		document.getElementById('tt').textContent=JSON.stringify(c.text||'');
		document.getElementById('tc').textContent=(c.codepoint||0)+' (0x'+(c.codepoint||0).toString(16).toUpperCase()+')';
		document.getElementById('tw').textContent=String(!!c.wide);
		document.getElementById('tb').textContent=String(!!c.bold);
		document.getElementById('ti').textContent=String(!!c.italic);
		document.getElementById('tf').textContent=c.fg||'default';
		document.getElementById('tg').textContent=c.bg||'default';
		document.getElementById('tv').textContent=String(!!c.inverse);
		if(c.hyperlink){document.getElementById('th').style.display='flex';document.getElementById('tl').textContent=c.hyperlink}
		else{document.getElementById('th').style.display='none'}
	}
	function hideTooltip(){_t.style.display='none';_cx=-1;_cy=-1}
	]]></script>`)

	// Embed cell grid as JSON for golden file verification and tooltip lookup
	if jsonView, err := BuildJSONView(r.term); err == nil {
		if jsonData, err := json.Marshal(jsonView); err == nil {
			fmt.Fprintf(&r.buf, `<script id="%s" type="application/json">%s</script>`,
				SVGGridJSONID, r.escapeXML(string(jsonData)))
		}
	}

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
		ci.Inverse = style.Inverse() != r.reverseColors
		if inverse := ci.Inverse; inverse {
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
		return r.resolveColorString(fg, true)
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
		return r.resolveColorString(bg, false)
	}
}

// resolveColorString resolves a color string (e.g. "color3", "#rrggbb") to a
// valid SVG hex color. Palette references like "colorN" are resolved using the
// terminal's palette; hex colors pass through unchanged.
func (r *svgRenderer) resolveColorString(s string, isFg bool) string {
	if len(s) > 5 && s[:5] == "color" {
		var idx int
		if _, err := fmt.Sscanf(s, "color%d", &idx); err == nil {
			if r.colors != nil && idx < len(r.colors.Palette) {
				c := r.colors.Palette[idx]
				return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
			}
			if isFg {
				return "#d4d4d4"
			}
			return "#171717"
		}
	}
	return s
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
