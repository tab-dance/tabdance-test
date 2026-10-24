package tape

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/vhs/parser"
	"github.com/charmbracelet/vhs/token"
	libghostty "go.mitchellh.com/libghostty"
)

type TapeSettings struct {
	Shell       string
	Width       int
	Height      int
	TypingSpeed time.Duration
	WaitTimeout time.Duration
}

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
	RecentOutput(n int) []byte
	LastOutputTime() time.Time
	ExitState() (exited bool, exitCode int, signal string)
}

func NewTapeExecutor(terminal Terminal, timeout time.Duration) *TapeExecutor {
	return &TapeExecutor{
		terminal:    terminal,
		timeout:     timeout,
		typingSpeed: 50 * time.Millisecond,
	}
}

func (e *TapeExecutor) SetTypingSpeed(d time.Duration) {
	e.typingSpeed = d
}

func (e *TapeExecutor) Execute(tapePath string) error {
	data, err := os.ReadFile(tapePath)
	if err != nil {
		return fmt.Errorf("read tape: %w", err)
	}
	return e.ExecuteString(string(data))
}

// ExecuteString plays a tape from an in-memory source. Tabdance-test
// extension verbs (ExpectExit, WaitExit, WaitStable, Expect) are executed
// inline, interleaved with the VHS commands around them.
func (e *TapeExecutor) ExecuteString(src string) error {
	exts, segments, err := splitExtensions(src)
	if err != nil {
		return err
	}

	if len(exts) == 0 {
		commands, err := parseCommands(src)
		if err != nil {
			return err
		}
		return e.executeCommands(commands)
	}

	for i, seg := range segments {
		if strings.TrimSpace(seg) != "" {
			commands, err := parseCommands(seg)
			if err != nil {
				return err
			}
			if err := e.executeCommands(commands); err != nil {
				return err
			}
		}
		if i < len(exts) {
			if err := e.executeExtension(exts[i], e.timeout); err != nil {
				return fmt.Errorf("%s (line %d): %w", extVerbName(exts[i]), exts[i].line, err)
			}
		}
	}
	return nil
}

func extVerbName(ext extension) string {
	switch ext.kind {
	case extExpectExit:
		return "ExpectExit"
	case extWaitExit:
		return "WaitExit"
	case extWaitStable:
		return "WaitStable"
	case extExpect:
		return "Expect"
	default:
		return "extension"
	}
}

func (e *TapeExecutor) executeCommands(commands []parser.Command) error {
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
		return e.executeKeyWithSpeed(cmd, libghostty.KeyEnter)
	case token.BACKSPACE:
		return e.executeKeyWithSpeed(cmd, libghostty.KeyBackspace)
	case token.TAB:
		return e.executeKeyWithSpeed(cmd, libghostty.KeyTab)
	case token.SPACE:
		return e.executeKeyWithSpeed(cmd, libghostty.KeySpace)
	case token.ESCAPE:
		return e.sendKeys(libghostty.KeyEscape, 0, 1, 0)
	case token.DELETE:
		return e.sendKeys(libghostty.KeyDelete, 0, 1, 0)
	case token.INSERT:
		return e.sendKeys(libghostty.KeyInsert, 0, 1, 0)
	case token.CTRL:
		return e.executeCtrl(cmd)
	case token.ALT:
		return e.executeAlt(cmd)
	case token.SHIFT:
		return e.executeShift(cmd)
	case token.UP:
		return e.sendKeys(libghostty.KeyArrowUp, 0, parseCount(cmd.Args, 1), e.keyDelay(cmd))
	case token.DOWN:
		return e.sendKeys(libghostty.KeyArrowDown, 0, parseCount(cmd.Args, 1), e.keyDelay(cmd))
	case token.LEFT:
		return e.sendKeys(libghostty.KeyArrowLeft, 0, parseCount(cmd.Args, 1), e.keyDelay(cmd))
	case token.RIGHT:
		return e.sendKeys(libghostty.KeyArrowRight, 0, parseCount(cmd.Args, 1), e.keyDelay(cmd))
	case token.PAGE_UP:
		return e.sendKeys(libghostty.KeyPageUp, 0, parseCount(cmd.Args, 1), e.keyDelay(cmd))
	case token.PAGE_DOWN:
		return e.sendKeys(libghostty.KeyPageDown, 0, parseCount(cmd.Args, 1), e.keyDelay(cmd))
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
	case "WaitTimeout":
		if d, err := time.ParseDuration(value); err == nil {
			e.timeout = d
		}
	case "Shell", "FontSize", "FontFamily", "Theme", "Framerate", "PlaybackSpeed":
		return nil
	}
	return nil
}
