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
		{"a", "key_a"},
		{"F", "key_f"},
		{"z", "key_z"},
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

func TestParseKeyCombo(t *testing.T) {
	tests := []struct {
		input string
		want  KeyInput
	}{
		{"F", KeyInput{Key: "F"}},
		{"Enter", KeyInput{Key: "Enter"}},
		{"Ctrl+F", KeyInput{Key: "F", Mods: "ctrl"}},
		{"ctrl+f", KeyInput{Key: "f", Mods: "ctrl"}},
		{"Ctrl+Alt+Enter", KeyInput{Key: "Enter", Mods: "ctrl+alt"}},
		{"Shift+Tab", KeyInput{Key: "Tab", Mods: "shift"}},
		{"Ctrl+Shift+F", KeyInput{Key: "F", Mods: "ctrl+shift"}},
		{"Alt+F1", KeyInput{Key: "F1", Mods: "alt"}},
		{"Super+c", KeyInput{Key: "c", Mods: "super"}},
		{"Control+X", KeyInput{Key: "X", Mods: "control"}},
		{"Cmd+Q", KeyInput{Key: "Q", Mods: "cmd"}},
		{"arrow_down", KeyInput{Key: "arrow_down"}},
		{"Ctrl", KeyInput{Key: "Ctrl"}},
		{"", KeyInput{Key: ""}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseKeyCombo(tt.input)
			if got != tt.want {
				t.Errorf("ParseKeyCombo(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}
