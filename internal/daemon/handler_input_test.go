package daemon

import (
	"testing"

	"github.com/tab-dance/tabdance-test/internal/terminal"
)

func TestHandleKeysStringArray(t *testing.T) {
	d := New()
	mt, err := d.manager.Create(terminal.CreateConfig{Cols: 20, Rows: 5, Command: "sleep 300"})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	t.Cleanup(func() { d.manager.Delete(mt.ID) })

	// The CLI sends keys as an array of combo strings plus optional mods.
	resp := d.handleKeys(Request{Params: map[string]any{
		"id":   mt.ID,
		"keys": []any{"Enter", "Tab"},
		"mods": "ctrl",
	}})
	if !resp.OK {
		t.Fatalf("keys failed: %s", resp.Error)
	}

	resp = d.handleKeys(Request{Params: map[string]any{
		"id":   mt.ID,
		"keys": []any{"Ctrl+X", "Ctrl+C"},
	}})
	if !resp.OK {
		t.Fatalf("keys with credit combos failed: %s", resp.Error)
	}

	// Empty key list must produce an error, not silently succeed.
	resp = d.handleKeys(Request{Params: map[string]any{"id": mt.ID}})
	if resp.OK || resp.Error == "" {
		t.Fatal("expected error for empty keys")
	}
}
