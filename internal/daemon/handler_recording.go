package daemon

import (
	"encoding/json"
	"net"
)

func (d *Daemon) handleRecord(req Request) Response {
	if err := d.svc.StartRecording(req.GetString("id"), req.GetString("path"), req.GetFloat("idle_time_limit")); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("recording started")
}

func (d *Daemon) handleStop(req Request) Response {
	if err := d.svc.StopRecording(req.GetString("id")); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("recording stopped")
}

func (d *Daemon) handleMarker(req Request) Response {
	if err := d.svc.AddMarker(req.GetString("id"), req.GetString("label")); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("marker added")
}

// handleAttach streams the terminal's live raw PTY output to the
// connection (tail semantics: bytes produced from attach on). It is
// transport-specific (raw byte streaming) and stays at the terminal
// level: the subscriber broadcasts exactly the bytes the child produces.
// Watching a screen is view --follow instead, which renders complete
// frames rather than trusting the consumer to be in byte-lockstep.
func (d *Daemon) handleAttach(req Request, conn net.Conn) {
	id := req.GetString("id")

	mt, err := d.manager.Get(id)
	if err != nil {
		json.NewEncoder(conn).Encode(ErrorResponse(err.Error()))
		return
	}
	if mt.Ptmx == nil {
		json.NewEncoder(conn).Encode(ErrorResponse("terminal has no PTY"))
		return
	}

	sub := mt.Subscribe()
	defer mt.Unsubscribe(sub)

	for data := range sub.Channel() {
		if _, err := conn.Write(data); err != nil {
			return
		}
	}
}
