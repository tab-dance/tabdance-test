package client

import (
	"encoding/json"
	"fmt"

	"github.com/tab-dance/tabdance-test/internal/wait"
)

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

func (c *Client) Wait(id, pattern, scope string, timeout float64, format string, trim, unwrap bool) (*wait.WaitForResult, error) {
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
	var result wait.WaitForResult
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
