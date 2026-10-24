package client

import "fmt"

// Output returns the terminal's recent raw PTY output, with control
// characters rendered visibly.
func (c *Client) Output(id string, n int) (string, error) {
	if n <= 0 {
		n = 4096
	}
	resp, err := c.Call("output", map[string]any{"id": id, "bytes": n})
	if err != nil {
		return "", err
	}
	if !resp.OK {
		return "", fmt.Errorf("%s", resp.Error)
	}
	result, _ := resp.Result.(map[string]any)
	out, _ := result["output"].(string)
	return out, nil
}
