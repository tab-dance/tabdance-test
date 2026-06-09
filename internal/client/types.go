package client

type terminalInfo struct {
	ID      string `json:"id"`
	Cols    uint16 `json:"cols"`
	Rows    uint16 `json:"rows"`
	Command string `json:"command"`
}

type viewResult struct {
	Content       string `json:"content"`
	CursorX       uint16 `json:"cursor_x"`
	CursorY       uint16 `json:"cursor_y"`
	CursorVisible bool   `json:"cursor_visible"`
	CursorStyle   string `json:"cursor_style"`
	ActiveScreen  string `json:"active_screen"`
}

type cursorResult struct {
	X           uint16 `json:"x"`
	Y           uint16 `json:"y"`
	Visible     bool   `json:"visible"`
	PendingWrap bool   `json:"pending_wrap"`
	Style       string `json:"style"`
}

type sizeResult struct {
	Cols     uint16 `json:"cols"`
	Rows     uint16 `json:"rows"`
	WidthPx  uint32 `json:"width_px"`
	HeightPx uint32 `json:"height_px"`
}

type cellResult struct {
	Codepoint uint32 `json:"codepoint"`
	Text      string `json:"text"`
	Wide      bool   `json:"wide"`
	Bold      bool   `json:"bold"`
	Italic    bool   `json:"italic"`
	FgColor   string `json:"fg"`
	BgColor   string `json:"bg"`
	Hyperlink string `json:"hyperlink,omitempty"`
}

type waitResult struct {
	Found     bool   `json:"found"`
	MatchLine string `json:"match_line,omitempty"`
	ElapsedMs int64  `json:"elapsed_ms"`
}

type verifyResult struct {
	Match bool   `json:"match"`
	Diff  string `json:"diff,omitempty"`
}
