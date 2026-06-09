package client

import "fmt"

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
