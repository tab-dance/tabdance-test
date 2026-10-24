package client

import (
	"encoding/json"
	"fmt"

	"github.com/tab-dance/tabdance-test/internal/terminal"
)

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

func (c *Client) Ls() ([]terminal.TerminalInfo, error) {
	resp, err := c.Call("ls", nil)
	if err != nil {
		return nil, err
	}
	if !resp.OK {
		return nil, fmt.Errorf("%s", resp.Error)
	}

	data, _ := json.Marshal(resp.Result)
	var infos []terminal.TerminalInfo
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
