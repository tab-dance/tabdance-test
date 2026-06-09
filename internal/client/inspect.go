package client

import (
	"encoding/json"
	"fmt"
)

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
