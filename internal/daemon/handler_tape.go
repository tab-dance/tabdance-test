package daemon

import (
	"time"

	"github.com/tab-dance/tabdance-test/internal/wait"
)

func (d *Daemon) handleTape(req Request) Response {
	timeout := req.GetFloat("timeout")
	if timeout == 0 {
		timeout = 30
	}
	if err := d.svc.PlayTape(req.GetString("id"), req.GetString("tape_path"), time.Duration(timeout*float64(time.Second))); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("tape executed")
}

func (d *Daemon) handleWait(req Request) Response {
	format := req.GetString("format")
	trim := req.GetBoolDefault("trim", true)
	unwrap := req.GetBoolDefault("unwrap", true)

	scope := req.GetString("scope")
	if scope == "" {
		scope = "screen"
	}
	timeout := req.GetFloat("timeout")
	if timeout == 0 {
		timeout = 10
	}

	result, err := d.svc.Wait(req.GetString("id"), wait.WaitForOptions{
		Pattern: req.GetString("pattern"),
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
