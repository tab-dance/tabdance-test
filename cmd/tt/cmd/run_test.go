package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func resetRunFlags(t *testing.T) {
	t.Helper()
	origStrict, origRecord, origUpdate := runStrict, runRecordPath, runUpdate
	origCols, origRows, origWorkingDir := runCols, runRows, runWorkingDir
	t.Cleanup(func() {
		runStrict, runRecordPath, runUpdate = origStrict, origRecord, origUpdate
		runCols, runRows, runWorkingDir = origCols, origRows, origWorkingDir
	})
	runStrict, runRecordPath, runUpdate = false, "", false
	runWorkingDir = ""
}

func writeSmokeTapes(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, name := range names {
		path := filepath.Join(dir, name+".tape")
		content := "Set Shell \"sh\"\n" +
			// Wait for the shell prompt so typing never races the line
			// editor's redraw (which can settle as e.g. "e$ cho ...").
			"Wait+Screen /\\$/\n" +
			"Type \"echo smoke-" + name + "\"\n" +
			"Enter\n" +
			// The Wait must anchor on output, not the typed echo (which
			// appears before Enter is processed); then settle so the next
			// step always sees a complete frame.
			"WaitStable @250ms\n" +
			"Expect /smoke-" + name + "/\n"
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			t.Fatal(err)
		}
		// Pre-bake a matching golden so the run verifies rather than skips.
		if err := os.WriteFile(filepath.Join(dir, name+".golden"), []byte("smoke\n"), 0644); err != nil {
			t.Fatal(err)
		}
		// Verify-mode needs a real matching golden; update mode below
		// creates them properly, so make verify runs pass by regenerating.
		_ = os.Remove(filepath.Join(dir, name+".golden"))
		if _, err := runTape(path, filepath.Join(dir, name+".golden"), true); err != nil {
			t.Fatalf("bake golden for %s: %v", name, err)
		}
	}
	return dir
}

func TestDirectoryJobsParity(t *testing.T) {
	resetRunFlags(t)
	dir := writeSmokeTapes(t, "alpha", "bravo", "charlie")

	serial, err := runDirectoryResults(dir, 1, "")
	if err != nil {
		t.Fatalf("serial run: %v", err)
	}
	parallel, err := runDirectoryResults(dir, 3, "")
	if err != nil {
		t.Fatalf("parallel run: %v", err)
	}

	if len(serial) != len(parallel) {
		t.Fatalf("result count mismatch: %d vs %d", len(serial), len(parallel))
	}
	for i := range serial {
		if serial[i].Tape != parallel[i].Tape {
			t.Fatalf("order mismatch at %d: %q vs %q", i, serial[i].Tape, parallel[i].Tape)
		}
		if (serial[i].Err == nil) != (parallel[i].Err == nil) {
			t.Fatalf("status mismatch for %s: serial=%v parallel=%v", serial[i].Tape, serial[i].Err, parallel[i].Err)
		}
	}
	for _, r := range parallel {
		if r.Err != nil {
			t.Fatalf("tape %s unexpectedly failed: %v", r.Tape, r.Err)
		}
	}
}

func TestDirectoryRunSkipsMissingGolden(t *testing.T) {
	resetRunFlags(t)
	dir := writeSmokeTapes(t, "skipme")
	// Remove the golden so the run skips it.
	if err := os.Remove(filepath.Join(dir, "skipme.golden")); err != nil {
		t.Fatal(err)
	}

	results, err := runDirectoryResults(dir, 1, "")
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if len(results) != 1 || !results[0].Skipped {
		t.Fatalf("expected skipped result, got %+v", results)
	}
}

func TestDirectoryRunRejectsRecordWithJobs(t *testing.T) {
	resetRunFlags(t)
	dir := writeSmokeTapes(t, "one", "two")

	_, err := runDirectoryResults(dir, 2, "/tmp/whatever.cast")
	if err == nil || !contains(err.Error(), "--record cannot be used with --jobs > 1") {
		t.Fatalf("expected record/jobs collision error, got: %v", err)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || stringContains(haystack, needle))
}

func stringContains(h, n string) bool {
	for i := 0; i+len(n) <= len(h); i++ {
		if h[i:i+len(n)] == n {
			return true
		}
	}
	return false
}
