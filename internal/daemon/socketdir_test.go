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
	rootKey := SessionKey()
	changeDir(t, sub)
	nestedKey := SessionKey()

	if rootKey != nestedKey {
		t.Fatalf("nested dir should share the repo-root daemon key: %q vs %q", rootKey, nestedKey)
	}

}

func TestSessionKeysDistinctForSeparatePaths(t *testing.T) {
	a := t.TempDir()
	b := t.TempDir()
	changeDir(t, a)
	ka := SessionKey()
	changeDir(t, b)
	kb := SessionKey()
	if ka == kb && a != b {
		t.Fatalf("distinct roots must get distinct daemon sockets: %q == %q", ka, kb)
	}
	if len(ka) != 12 || len(kb) != 12 || strings.ContainsAny(ka, "/") {
		t.Fatalf("unexpected key format: %q", ka)
	}
}

// TestSocketPathIsGlobal pins the single-daemon model: the socket lives in
// one shared directory; project scoping happens via terminal tags and ls
// filtering, not via separate sockets.
func TestSocketPathIsGlobal(t *testing.T) {
	path := SocketPath()
	if name := filepath.Base(path); name != "daemon.sock" {
		t.Fatalf("unexpected socket file name: %q", name)
	}
	dir := filepath.Dir(path)
	base := filepath.Base(dir)
	if strings.ContainsAny(base, "0123456789abcdef") && len(base) == 12 {
		t.Fatalf("socket dir must not be project-keyed anymore: %q", dir)
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
