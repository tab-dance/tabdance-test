package client

import (
	"encoding/json"
	"fmt"
)

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
