package tape

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/vhs/parser"
	libghostty "go.mitchellh.com/libghostty"
)

func (e *TapeExecutor) executeType(cmd parser.Command) error {
	text := cmd.Args
	typingSpeed := e.typingSpeed
	if cmd.Options != "" {
		if d, err := parseTypingSpeed(cmd.Options); err == nil {
			typingSpeed = d
		}
	}

	for _, r := range text {
		e.terminal.Lock()
		event, err := libghostty.NewKeyEvent()
		if err != nil {
			e.terminal.Unlock()
			return err
		}
		event.SetAction(libghostty.KeyActionPress)
		event.SetUTF8(string(r))
		seq, err := e.terminal.GetKeyEncoder().Encode(event)
		event.Close()
		e.terminal.Unlock()
		if err != nil {
			return err
		}
		if seq != nil {
			if _, err := e.terminal.WritePty(seq); err != nil {
				return fmt.Errorf("write type seq: %w", err)
			}
		}
		time.Sleep(typingSpeed)
	}
	return nil
}

func (e *TapeExecutor) keyDelay(cmd parser.Command) time.Duration {
	if cmd.Options != "" {
		if d, err := parseTypingSpeed(cmd.Options); err == nil {
			return d
		}
	}
	return e.typingSpeed
}

func (e *TapeExecutor) executeKeyWithSpeed(cmd parser.Command, key libghostty.Key) error {
	count := parseCount(cmd.Args, 1)
	delay := e.keyDelay(cmd)
	return e.sendKeys(key, 0, count, delay)
}

func (e *TapeExecutor) sendKeys(key libghostty.Key, mods libghostty.Mods, count int, delay time.Duration) error {
	for i := range count {
		if i > 0 && delay > 0 {
			time.Sleep(delay)
		}
		e.terminal.Lock()
		event, err := libghostty.NewKeyEvent()
		if err != nil {
			e.terminal.Unlock()
			return err
		}
		event.SetAction(libghostty.KeyActionPress)
		event.SetKey(key)
		event.SetMods(mods)

		seq, err := e.terminal.GetKeyEncoder().Encode(event)
		if err != nil {
			event.Close()
			e.terminal.Unlock()
			return err
		}

		if seq != nil {
			if _, err := e.terminal.WritePty(seq); err != nil {
				event.Close()
				e.terminal.Unlock()
				return fmt.Errorf("write key seq: %w", err)
			}
		}

		kkf, _ := e.terminal.GetTerminal().KittyKeyboardFlags()
		if kkf != libghostty.KittyKeyDisabled {
			event.SetAction(libghostty.KeyActionRelease)
			seq, _ = e.terminal.GetKeyEncoder().Encode(event)
			if seq != nil {
				if _, err := e.terminal.WritePty(seq); err != nil {
					event.Close()
					e.terminal.Unlock()
					return fmt.Errorf("write release seq: %w", err)
				}
			}
		}

		event.Close()
		e.terminal.Unlock()
	}
	return nil
}

func (e *TapeExecutor) executeCtrl(cmd parser.Command) error {
	key, mods := parseCtrlCombo(cmd.Args)
	return e.sendKeys(key, mods, 1, 0)
}

func (e *TapeExecutor) executeAlt(cmd parser.Command) error {
	key, mods := parseAltCombo(cmd.Args)
	return e.sendKeys(key, mods, 1, 0)
}

func (e *TapeExecutor) executeShift(cmd parser.Command) error {
	key, mods := parseShiftCombo(cmd.Args)
	return e.sendKeys(key, mods, 1, 0)
}

func (e *TapeExecutor) executeSleep(cmd parser.Command) error {
	d := parseDuration(cmd.Args)
	time.Sleep(d)
	return nil
}

func (e *TapeExecutor) executeWait(cmd parser.Command) error {
	parts := strings.SplitN(cmd.Args, " ", 2)
	scope := parts[0]
	pattern := ""
	if len(parts) > 1 {
		pattern = parts[1]
	}

	timeout := e.timeout
	if cmd.Options != "" {
		if d, err := time.ParseDuration(cmd.Options); err == nil {
			if d > 0 {
				timeout = d
			}
		}
	}

	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("compile pattern: %w", err)
	}

	deadline := time.Now().Add(timeout)
	for time.Now().Before(deadline) {
		e.terminal.Lock()
		f, err := libghostty.NewFormatter(e.terminal.GetTerminal(),
			libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
			libghostty.WithFormatterTrim(true),
			libghostty.WithFormatterUnwrap(true),
		)
		if err != nil {
			e.terminal.Unlock()
			return err
		}
		content, _ := f.FormatString()
		f.Close()

		matched := false
		switch scope {
		case "Screen":
			matched = re.MatchString(content)
		case "Line":
			cursorY, _ := e.terminal.GetTerminal().CursorY()
			lines := strings.Split(content, "\n")
			if int(cursorY) < len(lines) {
				matched = re.MatchString(lines[cursorY])
			}
		}
		e.terminal.Unlock()

		if matched {
			return nil
		}

		time.Sleep(50 * time.Millisecond)
	}

	return fmt.Errorf("wait timeout: pattern %q not found within %v", pattern, timeout)
}
