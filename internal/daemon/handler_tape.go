package daemon

import (
	"fmt"
	"time"

	"github.com/tab-dance/tabdance-test/internal/tape"
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
	return SuccessResponse(result)
}
