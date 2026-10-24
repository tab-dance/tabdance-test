package daemon

import (
	"encoding/json"
	"fmt"
	"net"
)

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

	sub := mt.Subscribe()
	defer mt.Unsubscribe(sub)

	var lastDropped int64 = 0
	for data := range sub.Channel() {
		if d := sub.Dropped(); d != lastDropped {
			loss := d - lastDropped
			lastDropped = d
			notice := []byte(fmt.Sprintf("\n[[tabdance-test: %d bytes of output dropped during attach]]\n", loss))
			if _, err := conn.Write(notice); err != nil {
				return
			}
		}
		if _, err := conn.Write(data); err != nil {
			return
		}
	}
}

// handleOutput returns the terminal's recent raw PTY output. Control
// bytes are rendered visibly so agents can read escape sequences.
func (d *Daemon) handleOutput(req Request) Response {
	id := req.GetString("id")
	bytes := req.GetInt("bytes")
	if bytes <= 0 {
		bytes = 4096
	}

	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	return SuccessResponse(map[string]any{
		"output": string(mt.RecentOutput(bytes)),
	})
}
