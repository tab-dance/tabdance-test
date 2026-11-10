package wait

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	libghostty "go.mitchellh.com/libghostty"

	"github.com/tab-dance/tabdance-test/internal/inspect"
)

type WaitForOptions struct {
	Pattern      string
	Scope        string
	Timeout      time.Duration
	PollInterval time.Duration
	Format       string
	Trim         bool
	Unwrap       bool
}

type WaitForResult struct {
	Found     bool   `json:"found"`
	MatchLine string `json:"match_line,omitempty"`
	ElapsedMs int64  `json:"elapsed_ms"`
}

func (r WaitForResult) ToJSON() string {
	data, _ := json.Marshal(r)
	return string(data)
}

func WaitFor(mu sync.Locker, terminal *libghostty.Terminal, opts WaitForOptions) (*WaitForResult, error) {
	re, err := regexp.Compile(opts.Pattern)
	if err != nil {
		return nil, fmt.Errorf("compile pattern: %w", err)
	}

	if opts.PollInterval <= 0 {
		opts.PollInterval = 50 * time.Millisecond
	}

	start := time.Now()
	deadline := start.Add(opts.Timeout)
	for time.Now().Before(deadline) {
		mu.Lock()
		content, err := getTerminalContent(terminal, opts.Format, opts.Trim, opts.Unwrap)
		mu.Unlock()
		if err != nil {
			return nil, err
		}

		switch normalizeScope(opts.Scope) {
		case "line":
			mu.Lock()
			cursorY, _ := terminal.CursorY()
			mu.Unlock()
			lines := strings.Split(content, "\n")
			if int(cursorY) < len(lines) && re.MatchString(lines[cursorY]) {
				elapsed := time.Since(start)
				return &WaitForResult{Found: true, MatchLine: lines[cursorY], ElapsedMs: elapsed.Milliseconds()}, nil
			}
		default:
			if re.MatchString(content) {
				elapsed := time.Since(start)
				return &WaitForResult{Found: true, ElapsedMs: elapsed.Milliseconds()}, nil
			}
		}

		time.Sleep(opts.PollInterval)
	}

	return &WaitForResult{Found: false, ElapsedMs: opts.Timeout.Milliseconds()}, nil
}

func normalizeScope(scope string) string {
	switch strings.ToLower(strings.TrimSpace(strings.TrimPrefix(scope, "+"))) {
	case "line":
		return "line"
	default:
		return "screen"
	}
}

func getTerminalContent(terminal *libghostty.Terminal, format string, trim, unwrap bool) (string, error) {
	return inspect.RenderContent(terminal, format, trim, unwrap)
}

// Repr renders raw PTY bytes with control characters shown, for putting
// recent output into failure diagnostics.
func Repr(b []byte) string {
	var sb strings.Builder
	for _, c := range b {
		switch {
		case c == '\r':
			sb.WriteString("\\r")
		case c == 0x1b:
			sb.WriteString("\\x1b")
		case c < 0x20 && c != '\n':
			sb.WriteRune(rune('@' + c))
		case c >= 0x7f:
			sb.WriteString(fmt.Sprintf("\\x%02x", c))
		default:
			sb.WriteByte(c)
		}
	}
	return sb.String()
}
