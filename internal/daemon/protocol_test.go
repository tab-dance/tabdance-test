package daemon

import "testing"

func TestGetBoolDefault(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]any
		def    bool
		want   bool
	}{
		{"absent key uses default", map[string]any{}, true, true},
		{"absent key uses default false", map[string]any{}, false, false},
		{"explicit true", map[string]any{"k": true}, false, true},
		{"explicit false", map[string]any{"k": false}, true, false},
		{"non-bool falls back", map[string]any{"k": "yes"}, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := Request{Params: tt.params}
			if got := req.GetBoolDefault("k", tt.def); got != tt.want {
				t.Errorf("GetBoolDefault() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestInstanceLockExclusive(t *testing.T) {
	path := t.TempDir() + "/daemon.lock"

	lock1, err := acquireInstanceLock(path)
	if err != nil {
		t.Fatalf("first acquireInstanceLock: %v", err)
	}
	defer lock1.Close()

	if _, err := acquireInstanceLock(path); err == nil {
		t.Fatal("expected second lock on the same path to fail")
	}
	if _, err := acquireInstanceLock(path); err == nil {
		t.Fatal("expected third lock on the same path to fail")
	}

	lock1.Close()

	lock2, err := acquireInstanceLock(path)
	if err != nil {
		t.Fatalf("expected acquire after release to succeed: %v", err)
	}
	lock2.Close()
}
