package daemon

import (
	"strings"
	"testing"
	"time"

	"github.com/tab-dance/tabdance-test/internal/terminal"
)

func TestHandleOutput(t *testing.T) {
	d := New()

	mt, err := d.manager.Create(terminal.CreateConfig{
		Cols: 20, Rows: 5, Command: "sh -c 'echo daemon-output-handler'; exit 0",
	})
	if err != nil {
		t.Fatalf("create terminal: %v", err)
	}
	defer d.manager.Delete(mt.ID)

	deadline := time.Now().Add(5 * time.Second)
	for !mt.Exited() && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	resp := d.handleOutput(Request{Params: map[string]any{"id": mt.ID, "bytes": 4096}})
	if !resp.OK {
		t.Fatalf("handleOutput error: %s", resp.Error)
	}
	result, _ := resp.Result.(map[string]any)
	out, _ := result["output"].(string)
	if !strings.Contains(out, "daemon-output-handler") {
		t.Fatalf("output missing token: %q", out)
	}
}
