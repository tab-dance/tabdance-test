package tape

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/vhs/lexer"
	"github.com/charmbracelet/vhs/parser"
	"github.com/charmbracelet/vhs/token"
	"github.com/tab-dance/tabdance-test/internal/input"
	libghostty "go.mitchellh.com/libghostty"
)

func ParseSettings(tapePath string) (TapeSettings, error) {
	data, err := os.ReadFile(tapePath)
	if err != nil {
		return TapeSettings{}, fmt.Errorf("read tape: %w", err)
	}

	// Extension verb lines (ExpectExit, WaitExit, WaitStable, Expect) are
	// not VHS commands; strip them before feeding the parser.
	_, segments, err := splitExtensions(string(data))
	if err != nil {
		return TapeSettings{}, fmt.Errorf("parse tape: %w", err)
	}

	parsed := ""
	for _, seg := range segments {
		parsed += seg + "\n"
	}

	l := lexer.New(parsed)
	p := parser.New(l)
	commands := p.Parse()

	if errs := p.Errors(); len(errs) > 0 {
		return TapeSettings{}, fmt.Errorf("parse tape: %s", errs[0].Msg)
	}

	var settings TapeSettings
	for _, cmd := range commands {
		if cmd.Type != token.SET {
			continue
		}
		setting := cmd.Options
		value := cmd.Args
		switch setting {
		case "Shell":
			settings.Shell = value
		case "Width":
			settings.Width = parseCount(value, 0)
		case "Height":
			settings.Height = parseCount(value, 0)
		case "TypingSpeed":
			if d, err := parseTypingSpeed(value); err == nil {
				settings.TypingSpeed = d
			}
		case "WaitTimeout":
			if d, err := time.ParseDuration(value); err == nil {
				settings.WaitTimeout = d
			}
		}
	}

	return settings, nil
}

func ShellToCommand(shell string) string {
	switch shell {
	case "bash":
		return "bash --norc --noprofile"
	case "zsh":
		return "zsh --no-rcs"
	case "fish":
		return `fish -C "function fish_greeting; end"`
	case "powershell", "pwsh":
		return shell + " -NoProfile"
	case "cmd":
		return "cmd"
	case "nu":
		return "nu --no-config-file"
	case "osh":
		return "osh --norc"
	case "elvish":
		// -norc: user rc files (starship etc.) leak prompts into goldens
		return "elvish -norc"
	case "xonsh":
		return "xonsh --no-rc -DXONSH_SUPPRESS_WELCOME=True"
	default:
		return shell
	}
}

func parseCount(s string, defaultVal int) int {
	if s == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return defaultVal
	}
	if n <= 0 {
		return 1
	}
	return n
}

func parseDuration(s string) time.Duration {
	if d, err := time.ParseDuration(s); err == nil {
		return d
	}
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return time.Duration(f * float64(time.Second))
	}
	return time.Second
}

func parseTypingSpeed(s string) (time.Duration, error) {
	s = strings.TrimPrefix(s, "@")
	return time.ParseDuration(s)
}

func parseCtrlCombo(s string) (libghostty.Key, libghostty.Mods) {
	parts := strings.Split(s, "+")
	key := libghostty.KeyUnidentified
	mods := libghostty.ModCtrl

	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch strings.ToLower(p) {
		case "ctrl", "control":
			mods |= libghostty.ModCtrl
		case "alt":
			mods |= libghostty.ModAlt
		case "shift":
			mods |= libghostty.ModShift
		default:
			key = parseKeyChar(p)
		}
	}
	return key, mods
}

func parseAltCombo(s string) (libghostty.Key, libghostty.Mods) {
	parts := strings.Split(s, "+")
	key := libghostty.KeyUnidentified
	mods := libghostty.ModAlt

	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch strings.ToLower(p) {
		case "ctrl":
			mods |= libghostty.ModCtrl
		case "alt":
			mods |= libghostty.ModAlt
		case "shift":
			mods |= libghostty.ModShift
		default:
			key = parseKeyChar(p)
		}
	}
	return key, mods
}

func parseShiftCombo(s string) (libghostty.Key, libghostty.Mods) {
	parts := strings.Split(s, "+")
	key := libghostty.KeyUnidentified
	mods := libghostty.ModShift

	for _, p := range parts {
		p = strings.TrimSpace(p)
		switch strings.ToLower(p) {
		case "ctrl":
			mods |= libghostty.ModCtrl
		case "alt":
			mods |= libghostty.ModAlt
		case "shift":
			mods |= libghostty.ModShift
		default:
			key = parseKeyChar(p)
		}
	}
	return key, mods
}

func parseKeyChar(s string) libghostty.Key {
	k, err := libghostty.ParseKey(input.NormalizeKey(s))
	if err == nil {
		return k
	}
	return libghostty.KeyUnidentified
}
