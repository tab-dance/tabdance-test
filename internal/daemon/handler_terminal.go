package daemon

import (
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

func (d *Daemon) handleExec(req Request) Response {
	result, err := d.svc.Exec(createConfigFrom(req), req.GetString("record"), zeroOrTwo(req.GetFloat("idle_time_limit")))
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func createConfigFrom(req Request) terminal.CreateConfig {
	cols := uint16(req.GetInt("cols"))
	if cols == 0 {
		cols = 108
	}
	rows := uint16(req.GetInt("rows"))
	if rows == 0 {
		rows = 24
	}
	scrollback := uint(req.GetInt("scrollback"))
	if scrollback == 0 {
		scrollback = 10000
	}
	project := req.GetString("project")
	if project == "" {
		project = SessionKey()
	}
	return terminal.CreateConfig{
		Cols:             cols,
		Rows:             rows,
		Scrollback:       scrollback,
		Command:          req.GetString("command"),
		Env:              envMap(req),
		WorkingDirectory: req.GetString("working_directory"),
		ProjectKey:       project,
	}
}

func zeroOrTwo(v float64) float64 {
	if v == 0 {
		return 2.0
	}
	return v
}

func (d *Daemon) handleLs(req Request) Response {
	return SuccessResponse(d.svc.List(req.GetString("project"), req.GetBool("all")))
}

func (d *Daemon) handleInspect(req Request) Response {
	result, err := d.svc.Inspect(req.GetString("id"))
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleRm(req Request) Response {
	id := req.GetString("id")
	if err := d.svc.Delete(id); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("terminal \"" + id + "\" removed")
}
