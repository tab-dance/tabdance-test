package client

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"

	"github.com/tab-dance/tabdance-test/internal/daemon"
	"github.com/tab-dance/tabdance-test/internal/inspect"
)

func (c *Client) View(id, format string, trim, unwrap bool) (*inspect.ViewResult, error) {
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
	var result inspect.ViewResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// Follow streams live-view frames from the daemon to stdout. Each frame
// is a complete rendering of the screen (modes, style, cursor), so the
// consumer stays visually correct no matter when it connects. Ctrl+C
// detaches. Use this for watching a session; attach streams raw PTY
// bytes instead, for diagnostics and piping.
func (c *Client) Follow(id string, trim, unwrap bool) error {
	conn, err := c.connect()
	if err != nil {
		return err
	}
	defer conn.Close()

	req := daemon.Request{
		Method: "follow",
		Params: map[string]any{"id": id, "trim": trim, "unwrap": unwrap},
	}
	if err := json.NewEncoder(conn).Encode(&req); err != nil {
		return fmt.Errorf("send follow request: %w", err)
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT)
	defer signal.Stop(sigCh)
	go func() {
		<-sigCh
		conn.Close()
		os.Exit(0)
	}()

	dec := json.NewDecoder(conn)
	for {
		var resp daemon.Response
		if err := dec.Decode(&resp); err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if !resp.OK {
			return fmt.Errorf("%s", resp.Error)
		}
		data, _ := json.Marshal(resp.Result)
		var frame struct {
			Content string `json:"content"`
		}
		if err := json.Unmarshal(data, &frame); err != nil {
			return err
		}
		os.Stdout.WriteString(frame.Content)
	}
}

func (c *Client) Cursor(id string) (*inspect.CursorResult, error) {
	resp, err := c.Call("cursor", map[string]any{"id": id})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var result inspect.CursorResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Size(id string) (*inspect.SizeResult, error) {
	resp, err := c.Call("size", map[string]any{"id": id})
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var result inspect.SizeResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

func (c *Client) Cell(id string, x, y int) (*inspect.CellResult, error) {
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
	var result inspect.CellResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
