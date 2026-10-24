package daemon

import (
	"github.com/tab-dance/tabdance-test/internal/inspect"
)

func (d *Daemon) handleView(req Request) Response {
	id := req.GetString("id")
	mt, err := d.manager.Get(id)
	if err != nil {
		return ErrorResponse(err.Error())
	}

	format := req.GetString("format")
	if format == "" {
		format = "plain"
	}
	trim := req.GetBoolDefault("trim", true)
	unwrap := req.GetBoolDefault("unwrap", true)

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
