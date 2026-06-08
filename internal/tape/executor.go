package tape

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/vhs/lexer"
	"github.com/charmbracelet/vhs/parser"
	"github.com/charmbracelet/vhs/token"
	libghostty "go.mitchellh.com/libghostty"
)

type TapeExecutor struct {
	terminal    Terminal
	timeout     time.Duration
	typingSpeed time.Duration
	depth       int
}

type Terminal interface {
	Lock()
	Unlock()
	WritePty(data []byte) (int, error)
	GetTerminal() *libghostty.Terminal
	GetKeyEncoder() *libghostty.KeyEncoder
	GetMouseEncoder() *libghostty.MouseEncoder
	Resize(cols, rows uint16) error
	AddMarker(label string) error
}

func NewTapeExecutor(terminal Terminal, timeout time.Duration) *TapeExecutor {
	return &TapeExecutor{
		terminal:    terminal,
		timeout:     timeout,
		typingSpeed: 50 * time.Millisecond,
	}
}

func (e *TapeExecutor) Execute(tapePath string) error {
	data, err := os.ReadFile(tapePath)
	if err != nil {
		return fmt.Errorf("read tape: %w", err)
	}

	l := lexer.New(string(data))
	p := parser.New(l)
	commands := p.Parse()

	if errs := p.Errors(); len(errs) > 0 {
		return fmt.Errorf("parse tape: %s", errs[0].Msg)
	}

	for _, cmd := range commands {
		if err := e.executeCommand(cmd); err != nil {
			return fmt.Errorf("execute command %s: %w", cmd.Type, err)
		}
	}
	return nil
}

func (e *TapeExecutor) executeCommand(cmd parser.Command) error {
	switch cmd.Type {
	case token.TYPE:
		return e.executeType(cmd)
	case token.ENTER:
		return e.executeKey(token.ENTER, cmd.Args, libghostty.KeyEnter)
	case token.BACKSPACE:
		return e.executeKey(token.BACKSPACE, cmd.Args, libghostty.KeyBackspace)
	case token.TAB:
		return e.executeKey(token.TAB, cmd.Args, libghostty.KeyTab)
	case token.SPACE:
		return e.executeKey(token.SPACE, cmd.Args, libghostty.KeySpace)
	case token.ESCAPE:
		return e.sendKeys(libghostty.KeyEscape, 0, 1)
	case token.DELETE:
		return e.sendKeys(libghostty.KeyDelete, 0, 1)
	case token.INSERT:
		return e.sendKeys(libghostty.KeyInsert, 0, 1)
	case token.CTRL:
		return e.executeCtrl(cmd)
	case token.ALT:
		return e.executeAlt(cmd)
	case token.SHIFT:
		return e.executeShift(cmd)
	case token.UP:
		return e.sendKeys(libghostty.KeyArrowUp, 0, parseCount(cmd.Args, 1))
	case token.DOWN:
		return e.sendKeys(libghostty.KeyArrowDown, 0, parseCount(cmd.Args, 1))
	case token.LEFT:
		return e.sendKeys(libghostty.KeyArrowLeft, 0, parseCount(cmd.Args, 1))
	case token.RIGHT:
		return e.sendKeys(libghostty.KeyArrowRight, 0, parseCount(cmd.Args, 1))
	case token.PAGE_UP:
		return e.sendKeys(libghostty.KeyPageUp, 0, parseCount(cmd.Args, 1))
	case token.PAGE_DOWN:
		return e.sendKeys(libghostty.KeyPageDown, 0, parseCount(cmd.Args, 1))
	case token.SLEEP:
		return e.executeSleep(cmd)
	case token.WAIT:
		return e.executeWait(cmd)
	case token.SET:
		return e.executeSet(cmd)
	case token.HIDE:
		return e.terminal.AddMarker("hide")
	case token.SHOW:
		return e.terminal.AddMarker("show")
	case token.SCREENSHOT:
		return e.terminal.AddMarker(fmt.Sprintf("screenshot: %s", cmd.Args))
	case token.OUTPUT:
		return nil
	case token.REQUIRE:
		return nil
	case token.SOURCE:
		if e.depth >= 10 {
			return fmt.Errorf("source recursion depth limit exceeded")
		}
		e.depth++
		return e.Execute(cmd.Args)
	case token.ENV:
		return nil
	case token.SCROLL_UP:
		count := parseCount(cmd.Args, 1)
		e.terminal.Lock()
		e.terminal.GetTerminal().ScrollViewportDelta(-count)
		e.terminal.Unlock()
		return nil
	case token.SCROLL_DOWN:
		count := parseCount(cmd.Args, 1)
		e.terminal.Lock()
		e.terminal.GetTerminal().ScrollViewportDelta(count)
		e.terminal.Unlock()
		return nil
	case token.COPY, token.PASTE:
		return nil
	default:
		return fmt.Errorf("unsupported command: %s", cmd.Type)
	}
}

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

func (e *TapeExecutor) executeKey(cmdType, args string, key libghostty.Key) error {
	count := parseCount(args, 1)
	return e.sendKeys(key, 0, count)
}

func (e *TapeExecutor) sendKeys(key libghostty.Key, mods libghostty.Mods, count int) error {
	for range count {
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
	return e.sendKeys(key, mods, 1)
}

func (e *TapeExecutor) executeAlt(cmd parser.Command) error {
	key, mods := parseAltCombo(cmd.Args)
	return e.sendKeys(key, mods, 1)
}

func (e *TapeExecutor) executeShift(cmd parser.Command) error {
	key, mods := parseShiftCombo(cmd.Args)
	return e.sendKeys(key, mods, 1)
}

func (e *TapeExecutor) executeSleep(cmd parser.Command) error {
	d := parseDuration(cmd.Args)
	time.Sleep(d)
	return nil
}

func (e *TapeExecutor) executeWait(cmd parser.Command) error {
	// The upstream VHS parser puts scope and pattern in cmd.Args as
	// "Line <pattern>" or "Screen <pattern>", and the timeout in cmd.Options.
	parts := strings.SplitN(cmd.Args, " ", 2)
	scope := parts[0]
	pattern := ""
	if len(parts) > 1 {
		pattern = parts[1]
	}

	if cmd.Options != "" {
		if d, err := time.ParseDuration(cmd.Options); err == nil {
			if d > 0 {
				e.timeout = d
			}
		}
	}
	timeout := e.timeout

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

func (e *TapeExecutor) executeSet(cmd parser.Command) error {
	setting := cmd.Options
	value := cmd.Args

	switch setting {
	case "Width":
		cols := uint16(parseCount(value, 80))
		e.terminal.Lock()
		rows, _ := e.terminal.GetTerminal().Rows()
		e.terminal.Unlock()
		return e.terminal.Resize(cols, rows)
	case "Height":
		e.terminal.Lock()
		cols, _ := e.terminal.GetTerminal().Cols()
		e.terminal.Unlock()
		rows := uint16(parseCount(value, 24))
		return e.terminal.Resize(cols, rows)
	case "TypingSpeed":
		if d, err := parseTypingSpeed(value); err == nil {
			e.typingSpeed = d
		}
	case "Shell", "FontSize", "FontFamily", "Theme", "Framerate", "PlaybackSpeed", "WaitTimeout":
		return nil
	}
	return nil
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
	if len(s) == 1 {
		r := strings.ToUpper(s)
		k, err := libghostty.ParseKey("Key" + r)
		if err == nil {
			return k
		}
	}
	k, err := libghostty.ParseKey(s)
	if err == nil {
		return k
	}
	return libghostty.KeyUnidentified
}
