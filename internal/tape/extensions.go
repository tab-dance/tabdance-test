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
	"github.com/tab-dance/tabdance-test/internal/wait"
	libghostty "go.mitchellh.com/libghostty"
)

type extKind int

const (
	extExpectExit extKind = iota
	extWaitExit
	extWaitStable
	extExpect
	extWaitPrompt
	extWaitCommand
	extResize
)

// extension is one of the tabdance-test extension verbs that live outside
// the VHS tape language: ExpectExit, WaitExit, WaitStable and Expect.
type extension struct {
	kind    extKind
	code    int
	hasCode bool
	cols    int
	rows    int
	pattern string
	scope   string
	timeout time.Duration
	line    int // 1-based source line for error messages
}

// Extension verbs are recognized only at the start of a line; every other
// line must remain valid VHS tape syntax.
var extVerbs = map[string]extKind{
	"ExpectExit":  extExpectExit,
	"WaitExit":    extWaitExit,
	"WaitStable":  extWaitStable,
	"Expect":      extExpect,
	"WaitPrompt":  extWaitPrompt,
	"WaitCommand": extWaitCommand,
	"Resize":      extResize,
}

// splitExtensions separates extension lines from VHS tape segments while
// preserving execution order.
func splitExtensions(src string) (extensions []extension, segments []string, err error) {
	lines := strings.Split(src, "\n")
	var segment []string
	for i, line := range lines {
		verb, rest := firstWord(line)
		if kind, ok := extVerbs[verb]; ok && !strings.HasPrefix(strings.TrimSpace(line), "#") {
			ext, perr := parseExtension(kind, strings.TrimSpace(rest), i+1)
			if perr != nil {
				return nil, nil, fmt.Errorf("line %d: %w", i+1, perr)
			}
			extensions = append(extensions, ext)
			segments = append(segments, strings.Join(segment, "\n"))
			segment = nil
			continue
		}
		segment = append(segment, line)
	}
	segments = append(segments, strings.Join(segment, "\n"))
	return extensions, segments, nil
}

func firstWord(line string) (string, string) {
	trimmed := strings.TrimSpace(line)
	idx := strings.IndexAny(trimmed, " \t")
	if idx < 0 {
		return trimmed, ""
	}
	return trimmed[:idx], trimmed[idx:]
}

func parseExtension(kind extKind, rest string, line int) (extension, error) {
	ext := extension{kind: kind, line: line}
	rest, dur, err := splitTrailingDuration(rest)
	if err != nil {
		return ext, fmt.Errorf("invalid duration: %w", err)
	}
	ext.timeout = dur

	switch kind {
	case extWaitCommand:
		fields := strings.Fields(rest)
		if len(fields) > 1 {
			return ext, fmt.Errorf("WaitCommand takes an optional exit code, e.g. WaitCommand 0 @5s")
		}
		if len(fields) == 1 {
			code, err := strconv.Atoi(fields[0])
			if err != nil {
				return ext, fmt.Errorf("WaitCommand code %q is not a number", fields[0])
			}
			ext.code = code
			ext.hasCode = true
		}
	case extExpect:
		re, scope, ok := splitRegex(rest)
		if !ok {
			return ext, fmt.Errorf("Expect needs a /regex/, e.g. Expect /ready/")
		}
		ext.pattern = re
		ext.scope = scope
	case extExpectExit:
		fields := strings.Fields(rest)
		if len(fields) > 1 {
			return ext, fmt.Errorf("ExpectExit takes an optional exit code, e.g. ExpectExit 0")
		}
		if len(fields) == 0 {
			return ext, nil // any exit status is accepted
		}
		code, err := strconv.Atoi(fields[0])
		if err != nil {
			return ext, fmt.Errorf("ExpectExit code %q is not a number", fields[0])
		}
		ext.code = code
	case extResize:
		fields := strings.Fields(rest)
		if len(fields) != 2 {
			return ext, fmt.Errorf("Resize takes two dimensions, e.g. Resize 80 24")
		}
		cols, err1 := strconv.Atoi(fields[0])
		rows, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil || cols <= 0 || rows <= 0 {
			return ext, fmt.Errorf("Resize expects positive column and row counts")
		}
		ext.cols, ext.rows = cols, rows
	}
	return ext, nil
}

// splitTrailingDuration strips a trailing "@<dur>" or "<dur>" token.
func splitTrailingDuration(arg string) (string, time.Duration, error) {
	fields := strings.Fields(arg)
	if len(fields) == 0 {
		return "", 0, nil
	}
	last := fields[len(fields)-1]
	if !strings.HasPrefix(last, "@") && !looksLikeDuration(last) {
		return arg, 0, nil
	}
	d, err := time.ParseDuration(strings.TrimPrefix(last, "@"))
	if err != nil {
		return arg, 0, err
	}
	return strings.TrimSpace(strings.TrimSuffix(strings.TrimSpace(arg), last)), d, nil
}

var durationRe = regexp.MustCompile(`^\d+(\.\d+)?(ms|s|m|h)$`)

func looksLikeDuration(s string) bool { return durationRe.MatchString(s) }

// splitRegex parses a regex argument like "/pattern/" followed by an
// optional scope token (+Screen/+Line/Line/Screen).
func splitRegex(arg string) (pattern, scope string, ok bool) {
	first := strings.Index(arg, "/")
	last := strings.LastIndex(arg, "/")
	if first < 0 || last <= first {
		return "", "", false
	}
	pattern = arg[first+1 : last]
	scope = strings.TrimSpace(arg[last+1:])
	scope = strings.TrimPrefix(scope, "+")
	switch strings.ToLower(scope) {
	case "line":
		return pattern, "line", true
	default:
		return pattern, "screen", true
	}
}

// CheckStrict returns an error for any Sleep verb in a tape, for `run
// --strict` mode where sleeps are considered a determinism smell.
func CheckStrict(tapePath string) error {
	src, err := os.ReadFile(tapePath)
	if err != nil {
		return fmt.Errorf("read tape: %w", err)
	}
	_, segments, err := splitExtensions(string(src))
	if err != nil {
		return err
	}
	for _, seg := range segments {
		if strings.TrimSpace(seg) == "" {
			continue
		}
		commands, err := parseCommands(seg)
		if err != nil {
			return err
		}
		for _, cmd := range commands {
			if cmd.Type == token.SLEEP {
				return fmt.Errorf("strict mode: Sleep verbs are not allowed (remove it, or use WaitStable/Wait)")
			}
		}
	}
	return nil
}

// executeExtension dispatches a single extension verb against the terminal.
func (e *TapeExecutor) executeExtension(ext extension, defaultTimeout time.Duration) error {
	timeout := ext.timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	switch ext.kind {
	case extWaitExit:
		return e.waitExit(timeout, -1)
	case extExpectExit:
		return e.waitExit(timeout, ext.code)
	case extWaitStable:
		quiet := ext.timeout
		if quiet <= 0 {
			quiet = 250 * time.Millisecond
		}
		return e.waitStable(quiet)
	case extExpect:
		return e.expect(ext.pattern, ext.scope)
	case extWaitPrompt:
		return e.waitPrompt(timeout)
	case extWaitCommand:
		return e.waitCommand(timeout, ext.code, ext.hasCode)
	case extResize:
		return e.resize(ext.cols, ext.rows)
	}
	return fmt.Errorf("unknown extension kind %d", ext.kind)
}

// waitExit blocks until the child exits. When wantCode >= 0, it also
// asserts the exit code (signal-exited programs never match).
func (e *TapeExecutor) waitExit(timeout time.Duration, wantCode int) error {
	deadline := time.Now().Add(timeout)
	for {
		exited, code, signal := e.terminal.ExitState()
		if exited {
			if wantCode < 0 {
				return nil
			}
			if signal != "" {
				return fmt.Errorf("child killed by signal %s, but expected exit code %d", signal, wantCode)
			}
			if code != wantCode {
				return fmt.Errorf("expected exit code %d, got %d", wantCode, code)
			}
			return nil
		}
		if time.Now().After(deadline) {
			return e.timeoutError(fmt.Sprintf("child did not %s within %v", exitWording(wantCode), timeout))
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// resize resizes the terminal at the point the verb occurs.
func (e *TapeExecutor) resize(cols, rows int) error {
	if err := e.terminal.Resize(uint16(cols), uint16(rows)); err != nil {
		return fmt.Errorf("resize: %w", err)
	}
	return nil
}

func exitWording(wantCode int) string {
	if wantCode < 0 {
		return "exit"
	}
	return fmt.Sprintf("exit with code %d", wantCode)
}

// waitStable blocks until the terminal has produced no output for the
// given quiet window
func (e *TapeExecutor) waitStable(quiet time.Duration) error {
	if quiet < 20*time.Millisecond {
		quiet = 20 * time.Millisecond
	}
	deadline := time.Now().Add(e.timeout)
	for {
		last := e.terminal.LastOutputTime()
		if last.IsZero() || time.Since(last) >= quiet {
			return nil
		}
		if time.Now().After(deadline) {
			return e.timeoutError(fmt.Sprintf("terminal never settled: output received within the last %v", quiet))
		}
		time.Sleep(10 * time.Millisecond)
	}
}

// expect asserts that the current screen matches the pattern immediately,
// without waiting for it to appear.
func (e *TapeExecutor) expect(pattern, scope string) error {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return fmt.Errorf("compile pattern: %w", err)
	}

	e.terminal.Lock()
	content, serr := e.formattedScreenLocked()
	e.terminal.Unlock()
	if serr != nil {
		return serr
	}

	matched := false
	if scope == "line" {
		e.terminal.Lock()
		cursorY, _ := e.terminal.GetTerminal().CursorY()
		e.terminal.Unlock()
		lines := strings.Split(content, "\n")
		if int(cursorY) < len(lines) {
			matched = re.MatchString(lines[cursorY])
		}
	} else {
		matched = re.MatchString(content)
	}

	if matched {
		return nil
	}
	return fmt.Errorf("expect failed: pattern %q not on screen\n--- screen ---\n%s%s",
		pattern, content, recentSuffix(e.terminal))
}

// timeoutError builds a "pattern not found" error carrying the screen and
// recent PTY traffic for diagnosis.
func (e *TapeExecutor) timeoutError(msg string) error {
	e.terminal.Lock()
	content, err := e.formattedScreenLocked()
	e.terminal.Unlock()
	if err != nil {
		return fmt.Errorf("%s; could not render screen: %v", msg, err)
	}
	return fmt.Errorf("%s\n--- screen at timeout ---\n%s%s", msg, content, recentSuffix(e.terminal))
}

// formattedScreenLocked renders the current screen as plain text. The
// caller must hold the terminal lock.
func (e *TapeExecutor) formattedScreenLocked() (string, error) {
	f, err := e.newPlainFormatter()
	if err != nil {
		return "", err
	}
	defer f.Close()
	return f.FormatString()
}

// recentSuffix renders the recent-output section for error messages.
func recentSuffix(term Terminal) string {
	recent := term.RecentOutput(2048)
	if len(recent) == 0 {
		return ""
	}
	return "\n--- recent output (last " + strconv.Itoa(len(recent)) + " bytes, control chars shown) ---\n" + wait.Repr(recent)
}

// parseCommands runs the VHS lexer/parser over a source string.
func parseCommands(src string) ([]parser.Command, error) {
	l := lexer.New(src)
	p := parser.New(l)
	commands := p.Parse()
	if errs := p.Errors(); len(errs) > 0 {
		return nil, fmt.Errorf("parse tape: %s", errs[0].Msg)
	}
	return commands, nil
}

// newPlainFormatter creates a plain-text, trimmed, unwrapped formatter for
// the executor's terminal.
func (e *TapeExecutor) newPlainFormatter() (*libghostty.Formatter, error) {
	return libghostty.NewFormatter(e.terminal.GetTerminal(),
		libghostty.WithFormatterFormat(libghostty.FormatterFormatPlain),
		libghostty.WithFormatterTrim(true),
		libghostty.WithFormatterUnwrap(true),
	)
}

// waitPrompt blocks until the shell emits an OSC 133 prompt-start (A)
// marker other than those before this verb, so it works anywhere in the
// tape: initial prompt or the prompt redraw after a command.
func (e *TapeExecutor) waitPrompt(timeout time.Duration) error {
	base := e.terminal.PromptCount()
	deadline := time.Now().Add(timeout)
	for {
		if e.terminal.PromptCount() > base {
			return nil
		}
		if time.Now().After(deadline) {
			return e.timeoutError("no OSC 133 prompt marker received within " + timeout.String() +
				" (the shell must emit prompt markers, e.g. via shell integration)")
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// waitCommand blocks until the shell emits an OSC 133 command-finished
// (D) marker newer than this verb. With a code argument the marker must
// carry that exit status; markers without an exit code are ignored (the
// shell may choose not to send one).
func (e *TapeExecutor) waitCommand(timeout time.Duration, wantCode int, hasCode bool) error {
	base := e.terminal.CommandEndCount()
	deadline := time.Now().Add(timeout)
	for {
		if count := e.terminal.CommandEndCount(); count > base {
			_, _, code, hasExit := e.terminal.LastCommandEnd()
			if !hasCode {
				return nil
			}
			if !hasExit {
				// No status in the marker; keep waiting (rare).
			} else if code == wantCode {
				return nil
			} else {
				// The shell always sends its status, so a mismatch is a
				// real failure.
				return fmt.Errorf("command exited with status %d, expected %d (OSC 133 marker)", code, wantCode)
			}
		}
		if time.Now().After(deadline) {
			wording := "command-finished marker"
			if hasCode {
				wording = fmt.Sprintf("command-finished marker with status %d", wantCode)
			}
			return e.timeoutError("no OSC 133 " + wording + " within " + timeout.String() +
				" (the shell must emit command markers, e.g. via shell integration)")
		}
		time.Sleep(20 * time.Millisecond)
	}
}
