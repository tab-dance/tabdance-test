package wait

import "testing"

func TestNormalizeScope(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"screen", "screen"},
		{"Screen", "screen"},
		{"line", "line"},
		{"Line", "line"},
		{"+Line", "line"},
		{"+line", "line"},
		{"", "screen"},
		{"bogus", "screen"},
	}

	for _, tt := range tests {
		if got := normalizeScope(tt.input); got != tt.want {
			t.Errorf("normalizeScope(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
