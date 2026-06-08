package input

import "testing"

func TestNormalizeKey(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"Tab", "tab"},
		{"Enter", "enter"},
		{"ArrowDown", "arrow_down"},
		{"ArrowUp", "arrow_up"},
		{"ArrowLeft", "arrow_left"},
		{"ArrowRight", "arrow_right"},
		{"KeyA", "key_a"},
		{"KeyZ", "key_z"},
		{"Escape", "escape"},
		{"Backspace", "backspace"},
		{"Delete", "delete"},
		{"PageDown", "page_down"},
		{"PageUp", "page_up"},
		{"Space", "space"},
		{"F1", "f1"},
		{"F12", "f12"},
		{"CapsLock", "caps_lock"},
		{"PrintScreen", "print_screen"},
		{"Numpad0", "numpad_0"},
		{"NumpadEnter", "numpad_enter"},
		{"AudioVolumeDown", "audio_volume_down"},
		{"MediaPlayPause", "media_play_pause"},
		{"already_snake", "already_snake"},
		{"arrow_down", "arrow_down"},
		{"key_a", "key_a"},
		{"a", "a"},
		{"f1", "f1"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeKey(tt.input)
			if got != tt.want {
				t.Errorf("normalizeKey(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}
