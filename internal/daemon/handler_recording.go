package daemon

import (
	"encoding/json"
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

	for data := range sub.Channel() {
		if _, err := conn.Write(data); err != nil {
			return
		}
	}
}
