package cmd

import (
	"os"
	"testing"
)

func TestEffectiveViewFormatPiped(t *testing.T) {
	// Without a flag and with stdout redirected, the default is plain.
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	orig := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = orig }()

	cmd := viewCmd
	if cmd.Flags().Changed("format") {
		t.Skip("format flag explicitly set by test harness; rerun standalone")
	}
	if got := effectiveViewFormat(cmd); got != "plain" {
		t.Fatalf("expected plain for piped stdout, got %q", got)
	}
	_ = r
}
