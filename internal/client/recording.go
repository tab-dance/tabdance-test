package client

import "fmt"

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
