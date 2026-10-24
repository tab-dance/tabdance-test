package input

import (
	"fmt"
	"io"
	"strings"
	"unicode"

	libghostty "go.mitchellh.com/libghostty"
)

type KeyInput struct {
	Key  string `json:"key"`
	Mods string `json:"mods,omitempty"`
	Text string `json:"text,omitempty"`
}

var modifierNames = map[string]bool{
	"ctrl":      true,
	"alt":       true,
	"shift":     true,
	"super":     true,
	"control":   true,
	"cmd":       true,
	"command":   true,
	"opt":       true,
	"option":    true,
	"caps_lock": true,
	"num_lock":  true,
}

// ParseKeyCombo splits a key combo string like "Ctrl+F" or "Ctrl+Alt+Enter"
// into a KeyInput with separate Key and Mods fields. If the string contains
// no "+", it is returned as-is in the Key field with empty Mods.
func ParseKeyCombo(s string) KeyInput {
	parts := strings.Split(s, "+")
	if len(parts) == 1 {
		return KeyInput{Key: s}
	}

	var mods []string
	var key string
	for _, p := range parts {
		lower := strings.ToLower(p)
		if modifierNames[lower] && key == "" {
			mods = append(mods, lower)
		} else if key == "" {
			key = p
		} else {
			key = p
		}
	}

	if key == "" {
		return KeyInput{Key: s}
	}

	return KeyInput{
		Key:  key,
		Mods: strings.Join(mods, "+"),
	}
}

// w3cAliases maps W3C PascalCase key names to their libghostty snake_case
// equivalents. These are the names most likely to be used by agents
// following the W3C UI Events spec.
var w3cAliases = map[string]string{
	"Tab":         "tab",
	"Enter":       "enter",
	"Escape":      "escape",
	"Backspace":   "backspace",
	"Delete":      "delete",
	"Insert":      "insert",
	"Home":        "home",
	"End":         "end",
	"PageUp":      "page_up",
	"PageDown":    "page_down",
	"Space":       "space",
	"ArrowUp":     "arrow_up",
	"ArrowDown":   "arrow_down",
	"ArrowLeft":   "arrow_left",
	"ArrowRight":  "arrow_right",
	"CapsLock":    "caps_lock",
	"PrintScreen": "print_screen",
	"ScrollLock":  "scroll_lock",
	"NumLock":     "num_lock",
	"Pause":       "pause",
	"ContextMenu": "context_menu",
	"F1":          "f1",
	"F2":          "f2",
	"F3":          "f3",
	"F4":          "f4",
	"F5":          "f5",
	"F6":          "f6",
	"F7":          "f7",
	"F8":          "f8",
	"F9":          "f9",
	"F10":         "f10",
	"F11":         "f11",
	"F12":         "f12",
	"F13":         "f13",
	"F14":         "f14",
	"F15":         "f15",
	"F16":         "f16",
	"F17":         "f17",
	"F18":         "f18",
	"F19":         "f19",
	"F20":         "f20",
	"F21":         "f21",
	"F22":         "f22",
	"F23":         "f23",
	"F24":         "f24",
	"F25":         "f25",
	"Digit0":      "digit_0",
	"Digit1":      "digit_1",
	"Digit2":      "digit_2",
	"Digit3":      "digit_3",
	"Digit4":      "digit_4",
	"Digit5":      "digit_5",
	"Digit6":      "digit_6",
	"Digit7":      "digit_7",
	"Digit8":      "digit_8",
	"Digit9":      "digit_9",
	"Numpad0":     "numpad_0",
	"Numpad1":     "numpad_1",
	"Numpad2":     "numpad_2",
	"Numpad3":     "numpad_3",
	"Numpad4":     "numpad_4",
	"Numpad5":     "numpad_5",
	"Numpad6":     "numpad_6",
	"Numpad7":     "numpad_7",
	"Numpad8":     "numpad_8",
	"Numpad9":     "numpad_9",
}

// NormalizeKey converts W3C PascalCase key names (e.g. "Tab", "ArrowDown",
// "KeyA") to the snake_case format expected by libghostty.ParseKey
// (e.g. "tab", "arrow_down", "key_a"). If the name is already snake_case
// it is returned unchanged. Single ASCII letters are prefixed with "key_"
// (e.g. "F" → "key_f", "a" → "key_a").
func NormalizeKey(name string) string {
	if strings.Contains(name, "_") {
		return name
	}
	if mapped, ok := w3cAliases[name]; ok {
		return mapped
	}

	// Single ASCII letter → key_<lower>
	if len(name) == 1 {
		r := rune(name[0])
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return "key_" + strings.ToLower(name)
		}
	}

	var b strings.Builder
	for i, r := range name {
		if i > 0 && unicode.IsUpper(r) {
			b.WriteByte('_')
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

func SendKeys(enc *libghostty.KeyEncoder, ptmxWriter io.Writer, term *libghostty.Terminal, keys []KeyInput) error {
	for _, k := range keys {
		if k.Text != "" {
			if _, err := ptmxWriter.Write([]byte(k.Text)); err != nil {
				return fmt.Errorf("write text: %w", err)
			}
			continue
		}

		key, err := libghostty.ParseKey(NormalizeKey(k.Key))
		if err != nil {
			return fmt.Errorf("parse key %q: %w", k.Key, err)
		}

		var mods libghostty.Mods
		if k.Mods != "" {
			mods, err = libghostty.ParseMods(k.Mods)
			if err != nil {
				return fmt.Errorf("parse mods %q: %w", k.Mods, err)
			}
		}

		event, err := libghostty.NewKeyEvent()
		if err != nil {
			return fmt.Errorf("create key event: %w", err)
		}

		event.SetAction(libghostty.KeyActionPress)
		event.SetKey(key)
		event.SetMods(mods)

		seq, err := enc.Encode(event)
		event.Close()
		if err != nil {
			return fmt.Errorf("encode key: %w", err)
		}

		if seq != nil {
			if _, err := ptmxWriter.Write(seq); err != nil {
				return fmt.Errorf("write key seq: %w", err)
			}
		}

		kkf, _ := term.KittyKeyboardFlags()
		if kkf != libghostty.KittyKeyDisabled {
			releaseEvent, err := libghostty.NewKeyEvent()
			if err != nil {
				return fmt.Errorf("create release event: %w", err)
			}
			releaseEvent.SetAction(libghostty.KeyActionRelease)
			releaseEvent.SetKey(key)
			releaseEvent.SetMods(mods)

			seq, err = enc.Encode(releaseEvent)
			releaseEvent.Close()
			if err != nil {
				return fmt.Errorf("encode release: %w", err)
			}
			if seq != nil {
				if _, err := ptmxWriter.Write(seq); err != nil {
					return fmt.Errorf("write release seq: %w", err)
				}
			}
		}
	}
	return nil
}
