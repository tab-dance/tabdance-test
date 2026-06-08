package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"strconv"
	"time"

	"github.com/tab-dance/tabdance-test/internal/golden"
	"github.com/tab-dance/tabdance-test/internal/input"
	"github.com/tab-dance/tabdance-test/internal/inspect"
	"github.com/tab-dance/tabdance-test/internal/tape"
	"github.com/tab-dance/tabdance-test/internal/terminal"
	"github.com/tab-dance/tabdance-test/internal/wait"
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
	defer mt.Unlock()

	return SuccessResponse(map[string]any{
		"id":         mt.ID,
		"cols":       mt.Cols,
		"rows":       mt.Rows,
		"command":    mt.Command,
		"recording":  mt.Recorder != nil,
	})
}

func (d *Daemon) handleKill(req Request) Response {
	id := req.GetString("id")
	if err := d.manager.Delete(id); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(fmt.Sprintf("terminal %q killed", id))
}

func (d *Daemon) handleRm(req Request) Response {
	id := req.GetString("id")
	if err := d.manager.Delete(id); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(fmt.Sprintf("terminal %q removed", id))
}

func (d *Daemon) handleView(req Request) Response {
	id := req.GetString("id")
	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	format := req.GetString("format")
	trim := req.GetBool("trim")
	if _, ok := req.Params["trim"]; !ok {
		trim = true
	}
	unwrap := req.GetBool("unwrap")
	if _, ok := req.Params["unwrap"]; !ok {
		unwrap = true
	}

	mt.Lock()
	defer mt.Unlock()

	result, err := inspect.GetView(mt.Terminal, format, trim, unwrap)
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleCursor(req Request) Response {
	id := req.GetString("id")
	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	defer mt.Unlock()

	result := inspect.GetCursor(mt.Terminal)
	return SuccessResponse(result)
}

func (d *Daemon) handleSize(req Request) Response {
	id := req.GetString("id")
	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	defer mt.Unlock()

	result := inspect.GetSize(mt.Terminal)
	return SuccessResponse(result)
}

func (d *Daemon) handleCell(req Request) Response {
	id := req.GetString("id")
	x := uint16(req.GetInt("x"))
	y := uint32(req.GetInt("y"))

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	defer mt.Unlock()

	result, err := inspect.GetCell(mt.Terminal, x, y)
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleSend(req Request) Response {
	id := req.GetString("id")
	text := req.GetString("text")

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	defer mt.Unlock()

	if _, err := mt.WritePty([]byte(text)); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("text sent")
}

func (d *Daemon) handleKeys(req Request) Response {
	id := req.GetString("id")
	keys := req.GetStringSlice("keys")

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	var keyInputs []input.KeyInput
	for _, k := range keys {
		keyInputs = append(keyInputs, input.KeyInput{Key: k})
	}

	if len(keyInputs) == 0 {
		return ErrorResponse("no keys provided")
	}

	mt.Lock()
	defer mt.Unlock()

	if err := input.SendKeys(mt.KeyEncoder, mt, mt.Terminal, keyInputs); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("keys sent")
}

func (d *Daemon) handleMouse(req Request) Response {
	id := req.GetString("id")
	action := req.GetString("action")
	button := req.GetString("button")
	x := float32(req.GetFloat("x"))
	y := float32(req.GetFloat("y"))
	mods := req.GetString("mods")

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	defer mt.Unlock()

	if input.IsScrollAction(action) {
		if err := input.SendScroll(mt.MouseEncoder, mt, action, x, y, mods); err != nil {
			return ErrorResponse(err.Error())
		}
		return SuccessResponse("scroll event sent")
	}

	if err := input.SendMouse(mt.MouseEncoder, mt, input.MouseInput{
		Action: action,
		Button: button,
		X:      x,
		Y:      y,
		Mods:   mods,
	}); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("mouse event sent")
}

func (d *Daemon) handlePaste(req Request) Response {
	id := req.GetString("id")
	text := req.GetString("text")

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	defer mt.Unlock()

	if err := input.SendPaste(mt, mt.Terminal, text); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("paste sent")
}

func (d *Daemon) handleFocus(req Request) Response {
	id := req.GetString("id")
	event := req.GetString("event")

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	defer mt.Unlock()

	if err := input.SendFocus(mt, event); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("focus event sent")
}

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
	trim := req.GetBool("trim")
	if _, ok := req.Params["trim"]; !ok {
		trim = true
	}
	unwrap := req.GetBool("unwrap")
	if _, ok := req.Params["unwrap"]; !ok {
		unwrap = true
	}

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	result, err := wait.WaitFor(mt, mt.Terminal, wait.WaitForOptions{
		Pattern: pattern,
		Scope:  scope,
		Timeout: time.Duration(timeout * float64(time.Second)),
		Format: format,
		Trim:   trim,
		Unwrap: unwrap,
	})
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleVerify(req Request) Response {
	id := req.GetString("id")
	goldenPath := req.GetString("golden_path")
	format := req.GetString("format")
	if format == "" {
		format = "plain"
	}
	trim := req.GetBool("trim")
	if _, ok := req.Params["trim"]; !ok {
		trim = true
	}
	unwrap := req.GetBool("unwrap")
	if _, ok := req.Params["unwrap"]; !ok {
		unwrap = true
	}

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	defer mt.Unlock()

	result, err := golden.VerifyOutput(mt.Terminal, goldenPath, format, trim, unwrap)
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleSnapshot(req Request) Response {
	id := req.GetString("id")
	goldenPath := req.GetString("golden_path")
	format := req.GetString("format")
	if format == "" {
		format = "plain"
	}
	trim := req.GetBool("trim")
	if _, ok := req.Params["trim"]; !ok {
		trim = true
	}
	unwrap := req.GetBool("unwrap")
	if _, ok := req.Params["unwrap"]; !ok {
		unwrap = true
	}

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	mt.Lock()
	defer mt.Unlock()

	if err := golden.UpdateGolden(mt.Terminal, goldenPath, format, trim, unwrap); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("golden file updated")
}

func (d *Daemon) handleRecord(req Request) Response {
	id := req.GetString("id")
	path := req.GetString("path")
	idleLimit := req.GetFloat("idle_time_limit")
	if idleLimit == 0 {
		idleLimit = 2.0
	}

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	if err := mt.StartRecording(path, idleLimit); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("recording started")
}

func (d *Daemon) handleStop(req Request) Response {
	id := req.GetString("id")

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	if err := mt.StopRecording(); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("recording stopped")
}

func (d *Daemon) handleMarker(req Request) Response {
	id := req.GetString("id")
	label := req.GetString("label")

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	if err := mt.AddMarker(label); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("marker added")
}

func atoi(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}

func (d *Daemon) handleAttach(req Request, conn net.Conn) {
	id := req.GetString("id")

	mt, err := d.manager.Get(id)
	if err != nil {
		enc := json.NewEncoder(conn)
		enc.Encode(ErrorResponse(err.Error()))
		return
	}

	if mt.Ptmx == nil {
		enc := json.NewEncoder(conn)
		enc.Encode(ErrorResponse("terminal has no PTY"))
		return
	}

	buf := make([]byte, 4096)
	for {
		n, err := mt.Ptmx.Read(buf)
		if err != nil {
			return
		}
		if n == 0 {
			continue
		}

		if _, err := conn.Write(buf[:n]); err != nil {
			return
		}
	}
}
