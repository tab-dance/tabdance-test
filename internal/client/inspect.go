package client

import (
	"encoding/json"
	"fmt"

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
