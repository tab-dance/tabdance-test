package daemon

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSessionKeyBindsToGitRoot(t *testing.T) {
	dir := t.TempDir()
	sub := filepath.Join(dir, "nested", "deep")
	if err := os.MkdirAll(sub, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(dir, ".git"), 0755); err != nil {
		t.Fatal(err)
	}

	changeDir(t, dir)
	rootKey := sessionKey()
	changeDir(t, sub)
	nestedKey := sessionKey()

	if rootKey != nestedKey {
		t.Fatalf("nested dir should share the repo-root daemon key: %q vs %q", rootKey, nestedKey)
	}


}

func TestSessionKeyDistinctForSeparatePaths(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	changeDir(t, a)
	ka := sessionKey()
	changeDir(t, b)
	kb := sessionKey()
	if ka == kb && a != b {
		t.Fatalf("distinct roots must get distinct daemon sockets: %q == %q", ka, kb)
	}
	if len(ka) != 12 || len(kb) != 12 || strings.ContainsAny(ka, "/") {
		t.Fatalf("unexpected key format: %q", ka)
	}
}

func changeDir(t *testing.T, dir string) {
	t.Helper()
	orig, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chdir(orig) })
}

func sessionKey_blank(d string) string { return "" }
