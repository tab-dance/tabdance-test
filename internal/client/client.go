package client

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/tab-dance/tabdance-test/internal/daemon"
)

type Client struct {
	socketPath string
}

func New() *Client {
	return &Client{
		socketPath: daemon.SocketPath(),
	}
}

func (c *Client) connect() (net.Conn, error) {
	conn, err := net.DialTimeout("unix", c.socketPath, 5*time.Second)
	if err != nil {
		return nil, fmt.Errorf("connect to daemon: %w", err)
	}
	return conn, nil
}

func (c *Client) Call(method string, params map[string]any) (*daemon.Response, error) {
	conn, err := c.connect()
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	req := daemon.Request{Method: method, Params: params}
	enc := json.NewEncoder(conn)
	if err := enc.Encode(&req); err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}

	dec := json.NewDecoder(conn)
	var resp daemon.Response
	if err := dec.Decode(&resp); err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &resp, nil
}

func (c *Client) Exec(command string, cols, rows, scrollback int, env map[string]string, workingDir, recordPath string) (string, error) {
	params := map[string]any{
		"command":           command,
		"cols":              cols,
		"rows":              rows,
		"scrollback":        scrollback,
		"working_directory": workingDir,
	}
	if env != nil {
		params["env"] = env
	}
	if recordPath != "" {
		params["record"] = recordPath
	}

	resp, err := c.Call("exec", params)
	if err != nil {
		return "", err
	}
	if !resp.OK {
		return "", fmt.Errorf("%s", resp.Error)
	}

	result, ok := resp.Result.(map[string]any)
	if !ok {
		return "", fmt.Errorf("unexpected result type")
	}
	id, _ := result["id"].(string)
	return id, nil
}

func (c *Client) Ls() ([]terminalInfo, error) {
	resp, err := c.Call("ls", nil)
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var infos []terminalInfo
	if err := json.Unmarshal(data, &infos); err != nil {
		return nil, err
	}
	return infos, nil
}

func (c *Client) Inspect(id string) (map[string]any, error) {
	resp, err := c.Call("inspect", map[string]any{"id": id})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}
	result, _ := resp.Result.(map[string]any)
	return result, nil
}

func (c *Client) Kill(id string) error {
	resp, err := c.Call("kill", map[string]any{"id": id})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Rm(id string) error {
	resp, err := c.Call("rm", map[string]any{"id": id})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) View(id, format string, trim, unwrap bool) (*viewResult, error) {
	resp, err := c.Call("view", map[string]any{
		"id":     id,
		"format": format,
		"trim":   trim,
		"unwrap": unwrap,
	})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var result viewResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Cursor(id string) (*cursorResult, error) {
	resp, err := c.Call("cursor", map[string]any{"id": id})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var result cursorResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Size(id string) (*sizeResult, error) {
	resp, err := c.Call("size", map[string]any{"id": id})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var result sizeResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Cell(id string, x, y int) (*cellResult, error) {
	resp, err := c.Call("cell", map[string]any{
		"id": id,
		"x":  x,
		"y":  y,
	})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var result cellResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Send(id, text string) error {
	resp, err := c.Call("send", map[string]any{"id": id, "text": text})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Keys(id string, keys []string) error {
	resp, err := c.Call("keys", map[string]any{"id": id, "keys": keys})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Mouse(id, action, button string, x, y float32, mods string) error {
	resp, err := c.Call("mouse", map[string]any{
		"id":     id,
		"action": action,
		"button": button,
		"x":      x,
		"y":      y,
		"mods":   mods,
	})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Paste(id, text string) error {
	resp, err := c.Call("paste", map[string]any{"id": id, "text": text})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Focus(id, event string) error {
	resp, err := c.Call("focus", map[string]any{"id": id, "event": event})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Resize(id string, cols, rows int) error {
	resp, err := c.Call("resize", map[string]any{"id": id, "cols": cols, "rows": rows})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Tape(id, tapePath string, timeout float64) error {
	resp, err := c.Call("tape", map[string]any{
		"id":        id,
		"tape_path": tapePath,
		"timeout":   timeout,
	})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Wait(id, pattern, scope string, timeout float64, format string, trim, unwrap bool) (*waitResult, error) {
	resp, err := c.Call("wait", map[string]any{
		"id":      id,
		"pattern": pattern,
		"scope":   scope,
		"timeout": timeout,
		"format":  format,
		"trim":    trim,
		"unwrap":  unwrap,
	})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var result waitResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Verify(id, goldenPath, format string, trim, unwrap bool) (*verifyResult, error) {
	resp, err := c.Call("verify", map[string]any{
		"id":          id,
		"golden_path": goldenPath,
		"format":      format,
		"trim":        trim,
		"unwrap":      unwrap,
	})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var result verifyResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Snapshot(id, goldenPath, format string, trim, unwrap bool) error {
	resp, err := c.Call("snapshot", map[string]any{
		"id":          id,
		"golden_path": goldenPath,
		"format":      format,
		"trim":        trim,
		"unwrap":      unwrap,
	})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Record(id, path string, idleTimeLimit float64) error {
	resp, err := c.Call("record", map[string]any{
		"id":              id,
		"path":            path,
		"idle_time_limit": idleTimeLimit,
	})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Stop(id string) error {
	resp, err := c.Call("stop", map[string]any{"id": id})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

func (c *Client) Marker(id, label string) error {
	resp, err := c.Call("marker", map[string]any{"id": id, "label": label})
	if err != nil {
		return err
	}
	if !resp.OK {
		return fmt.Errorf("%s", resp.Error)
	}
	return nil
}

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
