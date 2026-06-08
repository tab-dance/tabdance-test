package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"

	"github.com/tab-dance/tabdance-test/internal/client"
	"github.com/tab-dance/tabdance-test/internal/daemon"
)

func newClient() *client.Client {
	if err := client.EnsureDaemon(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	return client.New()
}

func runMcp() error {
	d := daemon.New()
	return d.Run()
}

func runExec(opts execOptions) error {
	c := newClient()

	env := make(map[string]string)
	for _, e := range opts.env {
		k, v, _ := parseKeyValue(e)
		env[k] = v
	}

	id, err := c.Exec(opts.command, opts.cols, opts.rows, opts.scrollback, env, opts.workingDir, opts.recordPath)
	if err != nil {
		return err
	}

	if opts.detach {
		fmt.Println(id)
		return nil
	}

	return runAttach(id)
}

func runLs(jsonOutput bool) error {
	c := newClient()
	infos, err := c.Ls()
	if err != nil {
		return err
	}

	if jsonOutput {
		data, _ := json.Marshal(infos)
		fmt.Println(string(data))
		return nil
	}

	if len(infos) == 0 {
		fmt.Println("No active terminals")
		return nil
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tCOLS\tROWS\tCOMMAND")
	for _, info := range infos {
		fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", info.ID, info.Cols, info.Rows, info.Command)
	}
	w.Flush()
	return nil
}

func runInspect(id string, jsonOutput bool) error {
	c := newClient()
	info, err := c.Inspect(id)
	if err != nil {
		return err
	}

	if jsonOutput {
		data, _ := json.Marshal(info)
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("ID:         %s\n", info["id"])
	fmt.Printf("Command:    %s\n", info["command"])
	if cols, ok := info["cols"].(float64); ok {
		if rows, ok := info["rows"].(float64); ok {
			fmt.Printf("Size:       %.0fx%.0f\n", cols, rows)
		}
	}
	fmt.Printf("Recording:  %v\n", info["recording"])
	return nil
}

func runAttach(id string) error {
	c := newClient()
	return c.Attach(id)
}

func runKill(id string) error {
	c := newClient()
	if err := c.Kill(id); err != nil {
		return err
	}
	fmt.Printf("terminal %q killed\n", id)
	return nil
}

func runRm(id string) error {
	c := newClient()
	if err := c.Rm(id); err != nil {
		return err
	}
	fmt.Printf("terminal %q removed\n", id)
	return nil
}

func runView(id, format string, trim, unwrap, jsonOutput bool) error {
	c := newClient()
	result, err := c.View(id, format, trim, unwrap)
	if err != nil {
		return err
	}

	if jsonOutput {
		data, _ := json.Marshal(result)
		fmt.Println(string(data))
		return nil
	}

	fmt.Print(result.Content)
	return nil
}

func runCursor(id string, jsonOutput bool) error {
	c := newClient()
	result, err := c.Cursor(id)
	if err != nil {
		return err
	}

	if jsonOutput {
		data, _ := json.Marshal(result)
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("x:%d y:%d visible:%v style:%s pending_wrap:%v\n",
		result.X, result.Y, result.Visible, result.Style, result.PendingWrap)
	return nil
}

func runSize(id string, jsonOutput bool) error {
	c := newClient()
	result, err := c.Size(id)
	if err != nil {
		return err
	}

	if jsonOutput {
		data, _ := json.Marshal(result)
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("%dx%d (%dx%dpx)\n", result.Cols, result.Rows, result.WidthPx, result.HeightPx)
	return nil
}

func runCell(id, xStr, yStr string, jsonOutput bool) error {
	x, err := strconv.Atoi(xStr)
	if err != nil {
		return fmt.Errorf("invalid x: %w", err)
	}
	y, err := strconv.Atoi(yStr)
	if err != nil {
		return fmt.Errorf("invalid y: %w", err)
	}

	c := newClient()
	result, err := c.Cell(id, x, y)
	if err != nil {
		return err
	}

	if jsonOutput {
		data, _ := json.Marshal(result)
		fmt.Println(string(data))
		return nil
	}

	fmt.Printf("codepoint:%d text:%q wide:%v bold:%v italic:%v fg:%s bg:%s\n",
		result.Codepoint, result.Text, result.Wide, result.Bold, result.Italic, result.FgColor, result.BgColor)
	if result.Hyperlink != "" {
		fmt.Printf("hyperlink:%s\n", result.Hyperlink)
	}
	return nil
}

func runSend(id, text string) error {
	c := newClient()
	return c.Send(id, text)
}

func runKeys(id string, keys []string) error {
	c := newClient()
	return c.Keys(id, keys)
}

func runMouse(id, action, button, xStr, yStr, mods string) error {
	x, err := strconv.ParseFloat(xStr, 32)
	if err != nil {
		return fmt.Errorf("invalid x: %w", err)
	}
	y, err := strconv.ParseFloat(yStr, 32)
	if err != nil {
		return fmt.Errorf("invalid y: %w", err)
	}

	c := newClient()
	return c.Mouse(id, action, button, float32(x), float32(y), mods)
}

func runPaste(id, text string) error {
	c := newClient()
	return c.Paste(id, text)
}

func runFocus(id, event string) error {
	c := newClient()
	return c.Focus(id, event)
}

func runResize(id, colsStr, rowsStr string) error {
	cols, err := strconv.Atoi(colsStr)
	if err != nil {
		return fmt.Errorf("invalid cols: %w", err)
	}
	rows, err := strconv.Atoi(rowsStr)
	if err != nil {
		return fmt.Errorf("invalid rows: %w", err)
	}

	c := newClient()
	return c.Resize(id, cols, rows)
}

func runTape(id, tapePath string, timeout float64) error {
	c := newClient()
	return c.Tape(id, tapePath, timeout)
}

func runWait(id, pattern, scope string, timeout float64, format string, trim, unwrap bool) error {
	c := newClient()
	result, err := c.Wait(id, pattern, scope, timeout, format, trim, unwrap)
	if err != nil {
		return err
	}

	if !result.Found {
		fmt.Fprintf(os.Stderr, "pattern not found after %dms\n", result.ElapsedMs)
		os.Exit(1)
	}

	fmt.Printf("found after %dms\n", result.ElapsedMs)
	return nil
}

func runVerify(id, goldenPath, format string, trim, unwrap bool) error {
	c := newClient()
	result, err := c.Verify(id, goldenPath, format, trim, unwrap)
	if err != nil {
		return err
	}

	if result.Match {
		fmt.Println("PASS")
		return nil
	}

	fmt.Fprintf(os.Stderr, "FAIL\n%s\n", result.Diff)
	os.Exit(1)
	return nil
}

func runSnapshot(id, goldenPath, format string, trim, unwrap bool) error {
	c := newClient()
	if err := c.Snapshot(id, goldenPath, format, trim, unwrap); err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "UPDATED %s\n", goldenPath)
	return nil
}

func runRecord(id, path string, idleTimeLimit float64) error {
	c := newClient()
	return c.Record(id, path, idleTimeLimit)
}

func runStop(id string) error {
	c := newClient()
	return c.Stop(id)
}

func runMarker(id, label string) error {
	c := newClient()
	return c.Marker(id, label)
}

func runRun(tapePath string, opts runOptions) error {
	return runRunInProcess(tapePath, opts)
}

func listTerminalIDs() ([]string, error) {
	c, err := tryClient()
	if err != nil {
		return nil, err
	}
	infos, err := c.Ls()
	if err != nil {
		return nil, err
	}
	ids := make([]string, len(infos))
	for i, info := range infos {
		ids[i] = info.ID
	}
	return ids, nil
}

func tryClient() (*client.Client, error) {
	if !client.IsDaemonRunning() {
		return nil, fmt.Errorf("daemon not running")
	}
	return client.New(), nil
}

func parseKeyValue(s string) (string, string, error) {
	for i := 0; i < len(s); i++ {
		if s[i] == '=' {
			return s[:i], s[i+1:], nil
		}
	}
	return s, "", nil
}
