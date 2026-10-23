package tape

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	libghostty "go.mitchellh.com/libghostty"
)

func TestShellToCommand(t *testing.T) {
	tests := []struct {
		shell   string
		command string
	}{
		{"bash", "bash --norc --noprofile"},
		{"zsh", "zsh --no-rcs"},
		{"fish", `fish -C "function fish_greeting; end"`},
		{"powershell", "powershell -NoProfile"},
		{"pwsh", "pwsh -NoProfile"},
		{"cmd", "cmd"},
		{"nu", "nu --no-config-file"},
		{"osh", "osh --norc"},
		{"xonsh", "xonsh --no-rc -DXONSH_SUPPRESS_WELCOME=True"},
		{"elvish", "elvish"},
		{"custom-shell", "custom-shell"},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			got := ShellToCommand(tt.shell)
			if got != tt.command {
				t.Errorf("ShellToCommand(%q) = %q, want %q", tt.shell, got, tt.command)
			}
		})
	}
}

func TestParseSettings(t *testing.T) {
	tests := []struct {
		name     string
		tape     string
		settings TapeSettings
	}{
		{
			name: "shell setting",
			tape: `Set Shell "elvish"`,
			settings: TapeSettings{
				Shell: "elvish",
			},
		},
		{
			name: "all settings",
			tape: `Set Shell "fish"
Set Width 80
Set Height 24
Set TypingSpeed 100ms
Set WaitTimeout 30s`,
			settings: TapeSettings{
				Shell:       "fish",
				Width:        80,
				Height:       24,
				TypingSpeed:  100 * time.Millisecond,
				WaitTimeout:  30 * time.Second,
			},
		},
		{
			name: "no settings",
			tape: `Type "hello"
Enter`,
			settings: TapeSettings{},
		},
		{
			name: "shell with other commands",
			tape: `Set Shell "zsh"
Type "echo hello"
Enter`,
			settings: TapeSettings{
				Shell: "zsh",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "test.tape")
			if err := os.WriteFile(path, []byte(tt.tape), 0644); err != nil {
				t.Fatal(err)
			}

			settings, err := ParseSettings(path)
			if err != nil {
				t.Fatalf("ParseSettings() error: %v", err)
			}

			if settings.Shell != tt.settings.Shell {
				t.Errorf("Shell = %q, want %q", settings.Shell, tt.settings.Shell)
			}
			if settings.Width != tt.settings.Width {
				t.Errorf("Width = %d, want %d", settings.Width, tt.settings.Width)
			}
			if settings.Height != tt.settings.Height {
				t.Errorf("Height = %d, want %d", settings.Height, tt.settings.Height)
			}
			if settings.TypingSpeed != tt.settings.TypingSpeed {
				t.Errorf("TypingSpeed = %v, want %v", settings.TypingSpeed, tt.settings.TypingSpeed)
			}
			if settings.WaitTimeout != tt.settings.WaitTimeout {
				t.Errorf("WaitTimeout = %v, want %v", settings.WaitTimeout, tt.settings.WaitTimeout)
			}
		})
	}
}

func TestParseSettingsFileNotFound(t *testing.T) {
	_, err := ParseSettings("/nonexistent/path.tape")
	if err == nil {
		t.Fatal("expected error for nonexistent file")
	}
}

func TestNormalizeKeyName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"F", "key_f"},
		{"a", "key_a"},
		{"z", "key_z"},
		{"A", "key_a"},
		{"Z", "key_z"},
		{"KeyA", "key_a"},
		{"KeyZ", "key_z"},
		{"ArrowDown", "arrow_down"},
		{"ArrowUp", "arrow_up"},
		{"PageDown", "page_down"},
		{"PageUp", "page_up"},
		{"already_snake", "already_snake"},
		{"key_a", "key_a"},
		{"f1", "f1"},
		{"Enter", "enter"},
		{"Backspace", "backspace"},
		{"NumpadEnter", "numpad_enter"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := normalizeKeyName(tt.input)
			if got != tt.want {
				t.Errorf("normalizeKeyName(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestParseCtrlCombo(t *testing.T) {
	tests := []struct {
		input    string
		wantKey  libghostty.Key
		wantMods libghostty.Mods
	}{
		{"F", libghostty.KeyF, libghostty.ModCtrl},
		{"a", libghostty.KeyA, libghostty.ModCtrl},
		{"Ctrl+F", libghostty.KeyF, libghostty.ModCtrl},
		{"Ctrl+Alt+F", libghostty.KeyF, libghostty.ModCtrl | libghostty.ModAlt},
		{"Ctrl+Shift+L", libghostty.KeyL, libghostty.ModCtrl | libghostty.ModShift},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			key, mods := parseCtrlCombo(tt.input)
			if key != tt.wantKey {
				t.Errorf("parseCtrlCombo(%q) key = %v, want %v", tt.input, key, tt.wantKey)
			}
			if mods != tt.wantMods {
				t.Errorf("parseCtrlCombo(%q) mods = %v, want %v", tt.input, mods, tt.wantMods)
			}
		})
	}
}
