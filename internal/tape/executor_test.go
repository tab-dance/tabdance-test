package tape

import (
	"os"
	"path/filepath"
	"testing"
	"time"
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
		{"xonsh", "xonsh --no-rc"},
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
