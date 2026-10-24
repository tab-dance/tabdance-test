package daemon

import (
	"fmt"
	"time"

	"github.com/tab-dance/tabdance-test/internal/inspect"
	"github.com/tab-dance/tabdance-test/internal/tape"
	"github.com/tab-dance/tabdance-test/internal/terminal"
	"github.com/tab-dance/tabdance-test/internal/wait"
)

func (d *Daemon) handleResize(req Request) Response {
	id := req.GetString("id")
	cols := uint16(req.GetInt("cols"))
	rows := uint16(req.GetInt("rows"))

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	if err := mt.Resize(cols, rows); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(fmt.Sprintf("resized to %dx%d", cols, rows))
}

func (d *Daemon) handleTape(req Request) Response {
	id := req.GetString("id")
	tapePath := req.GetString("tape_path")
	timeout := req.GetFloat("timeout")
	if timeout == 0 {
		timeout = 30
	}

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	executor := tape.NewTapeExecutor(mt, time.Duration(timeout*float64(time.Second)))
	if err := executor.Execute(tapePath); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("tape executed")
}

func (d *Daemon) handleWait(req Request) Response {
	id := req.GetString("id")
	pattern := req.GetString("pattern")
	scope := req.GetString("scope")
	if scope == "" {
		scope = "screen"
	}
	timeout := req.GetFloat("timeout")
	if timeout == 0 {
		timeout = 10
	}
	format := req.GetString("format")
	if format == "" {
		format = "plain"
	}
	trim := req.GetBoolDefault("trim", true)
	unwrap := req.GetBoolDefault("unwrap", true)

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	result, err := wait.WaitFor(mt, mt.Terminal, wait.WaitForOptions{
		Pattern: pattern,
		Scope:   scope,
		Timeout: time.Duration(timeout * float64(time.Second)),
		Format:  format,
		Trim:    trim,
		Unwrap:  unwrap,
	})
	if err != nil {
		return ErrorResponse(err.Error())
	}
	if !result.Found {
		return ErrorResponse(waitTimeoutDetail(pattern, time.Duration(timeout*float64(time.Second)), mt, format, trim, unwrap))
	}
	return SuccessResponse(result)
}

// waitTimeoutDetail builds a diagnostic timeout error: the rendered screen
// plus a repr of the last PTY traffic.
func waitTimeoutDetail(pattern string, timeout time.Duration, mt *terminal.ManagedTerminal, format string, trim, unwrap bool) string {
	mt.Lock()
	view, err := inspect.GetView(mt.Terminal, format, trim, unwrap)
	mt.Unlock()

	msg := fmt.Sprintf("pattern %q not found within %v", pattern, timeout)
	if err != nil {
		return fmt.Sprintf("%s\n(could not render screen: %v)", msg, err)
	}
	recent := mt.RecentOutput(2048)
	detail := msg + "\n--- screen at timeout ---\n" + view.Content
	if len(recent) > 0 {
		detail += fmt.Sprintf("\n--- recent output (last %d bytes) ---\n%s", len(recent), wait.Repr(recent))
	}
	return detail
}
