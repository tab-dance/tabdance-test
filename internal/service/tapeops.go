package service

import (
	"fmt"
	"time"

	"github.com/tab-dance/tabdance-test/internal/golden"
	"github.com/tab-dance/tabdance-test/internal/inspect"
	"github.com/tab-dance/tabdance-test/internal/tape"
	"github.com/tab-dance/tabdance-test/internal/terminal"
	"github.com/tab-dance/tabdance-test/internal/wait"
)

func (s *Service) PlayTape(id, tapePath string, timeout time.Duration) error {
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	executor := tape.NewTapeExecutor(mt, timeout)
	return executor.Execute(tapePath)
}

// Wait polls until the pattern matches. A timeout returns an error
// carrying the rendered screen and recent PTY traffic.
func (s *Service) Wait(id string, opts wait.WaitForOptions) (*wait.WaitForResult, error) {
	if opts.Timeout <= 0 {
		opts.Timeout = 10 * time.Second
	}
	if opts.Format == "" {
		opts.Format = "plain"
	}
	mt, err := s.terminals.Get(id)
	if err != nil {
		return nil, err
	}

	result, err := wait.WaitFor(mt, mt.Terminal, opts)
	if err != nil {
		return nil, err
	}
	if !result.Found {
		return result, fmt.Errorf("%s", s.waitForTimeoutDetail(mt, opts.Pattern, opts.Timeout, opts.Format, opts.Trim, opts.Unwrap))
	}
	return result, nil
}

func (s *Service) waitForTimeoutDetail(mt *terminal.ManagedTerminal, pattern string, timeout time.Duration, format string, trim, unwrap bool) string {
	mt.Lock()
	view, viewErr := inspect.GetView(mt.Terminal, format, trim, unwrap)
	mt.Unlock()

	msg := fmt.Sprintf("pattern %q not found within %v", pattern, timeout)
	if viewErr != nil {
		return fmt.Sprintf("%s\n(could not render screen: %v)", msg, viewErr)
	}
	recent := mt.RecentOutput(2048)
	detail := msg + "\n--- screen at timeout ---\n" + view.Content
	if len(recent) > 0 {
		detail += fmt.Sprintf("\n--- recent output (last %d bytes) ---\n%s", len(recent), wait.Repr(recent))
	}
	return detail
}

// VerifyResult is the call result of golden verification.
type VerifyOutcome struct {
	Result *golden.VerifyResult
}

func (s *Service) Verify(id, goldenPath, format string, trim, unwrap bool) (*golden.VerifyResult, error) {
	if format == "" {
		format = "plain"
	}
	mt, err := s.terminals.Get(id)
	if err != nil {
		return nil, err
	}
	mt.Lock()
	defer mt.Unlock()
	return golden.VerifyOutput(mt.Terminal, goldenPath, format, trim, unwrap)
}

func (s *Service) UpdateGolden(id, goldenPath, format string, trim, unwrap bool) error {
	if format == "" {
		format = "plain"
	}
	mt, err := s.terminals.Get(id)
	if err != nil {
		return err
	}
	mt.Lock()
	defer mt.Unlock()
	return golden.UpdateGolden(mt.Terminal, goldenPath, format, trim, unwrap)
}
