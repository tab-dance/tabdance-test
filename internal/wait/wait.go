package wait

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	libghostty "go.mitchellh.com/libghostty"
)

type WaitForOptions struct {
	Pattern     string
	Scope       string
	Timeout     time.Duration
	PollInterval time.Duration
	Format      string
	Trim        bool
	Unwrap      bool
}

type WaitForResult struct {
	Matched   bool
	MatchLine string
	Elapsed   time.Duration
}

func (r WaitForResult) ToJSON() string {
	return fmt.Sprintf(`{"matched":%v,"elapsed":%q}`, r.Matched, r.Elapsed)
}

func WaitFor(terminal *libghostty.Terminal, opts WaitForOptions) (*WaitForResult, error) {
	re, err := regexp.Compile(opts.Pattern)
	if err != nil {
		return nil, fmt.Errorf("compile pattern: %w", err)
	}

	if opts.PollInterval <= 0 {
		opts.PollInterval = 50 * time.Millisecond
	}

	deadline := time.Now().Add(opts.Timeout)
	for time.Now().Before(deadline) {
		content, err := getTerminalContent(terminal, opts.Format, opts.Trim, opts.Unwrap)
		if err != nil {
			return nil, err
		}

		switch opts.Scope {
		case "line", "+Line":
			cursorY, _ := terminal.CursorY()
			lines := strings.Split(content, "\n")
			if int(cursorY) < len(lines) && re.MatchString(lines[cursorY]) {
				return &WaitForResult{Matched: true, MatchLine: lines[cursorY], Elapsed: time.Since(deadline.Add(-opts.Timeout))}, nil
			}
		default:
			if re.MatchString(content) {
				return &WaitForResult{Matched: true, Elapsed: time.Since(deadline.Add(-opts.Timeout))}, nil
			}
		}

		time.Sleep(opts.PollInterval)
	}

	return &WaitForResult{Matched: false, Elapsed: opts.Timeout}, nil
}

func getTerminalContent(terminal *libghostty.Terminal, format string, trim, unwrap bool) (string, error) {
	f, err := libghostty.NewFormatter(terminal,
		libghostty.WithFormatterFormat(parseFormat(format)),
		libghostty.WithFormatterTrim(trim),
		libghostty.WithFormatterUnwrap(unwrap),
	)
	if err != nil {
		return "", err
	}
	defer f.Close()

	content, err := f.FormatString()
	if err != nil {
		return "", err
	}
	return content, nil
}

func parseFormat(s string) libghostty.FormatterFormat {
	switch s {
	case "vt":
		return libghostty.FormatterFormatVT
	case "html":
		return libghostty.FormatterFormatHTML
	default:
		return libghostty.FormatterFormatPlain
	}
}
