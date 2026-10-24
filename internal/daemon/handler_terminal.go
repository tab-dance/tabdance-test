package daemon

import (
	"fmt"

	"github.com/tab-dance/tabdance-test/internal/terminal"
)

func (d *Daemon) handleExec(req Request) Response {
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
	command := req.GetString("command")
	workingDir := req.GetString("working_directory")

	var env map[string]string
	if v, ok := req.Params["env"]; ok {
		if m, ok := v.(map[string]any); ok {
			env = make(map[string]string, len(m))
			for k, v := range m {
				if s, ok := v.(string); ok {
					env[k] = s
				}
			}
		}
	}

	mt, err := d.manager.Create(terminal.CreateConfig{
		Cols:             cols,
		Rows:             rows,
		Scrollback:       scrollback,
		Command:          command,
		Env:              env,
		WorkingDirectory: workingDir,
	})
	if err != nil {
		return ErrorResponse(err.Error())
	}

	recordPath := req.GetString("record")
	if recordPath != "" {
		idleLimit := req.GetFloat("idle_time_limit")
		if idleLimit == 0 {
			idleLimit = 2.0
		}
		if err := mt.StartRecording(recordPath, idleLimit); err != nil {
			d.manager.Delete(mt.ID)
			return ErrorResponse(fmt.Sprintf("start recording: %v", err))
		}
	}

	return SuccessResponse(map[string]any{
		"id":   mt.ID,
		"cols": mt.Cols,
		"rows": mt.Rows,
	})
}

func (d *Daemon) handleLs(_ Request) Response {
	return SuccessResponse(d.manager.List())
}

func (d *Daemon) handleInspect(req Request) Response {
	id := req.GetString("id")
	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	cols, rows, command, recording := mt.Cols, mt.Rows, mt.Command, mt.IsRecording()
	mt.Unlock()

	result := map[string]any{
		"id":        mt.ID,
		"cols":      cols,
		"rows":      rows,
		"command":   command,
		"recording": recording,
	}
	if exited, code, signal := mt.ExitState(); exited {
		result["exited"] = true
		result["exit_code"] = code
		if signal != "" {
			result["exit_signal"] = signal
		}
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleRm(req Request) Response {
	id := req.GetString("id")
	if err := d.manager.Delete(id); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(fmt.Sprintf("terminal %q removed", id))
}
