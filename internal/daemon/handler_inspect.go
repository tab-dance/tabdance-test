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
	trim := req.GetBool("trim")
	if _, ok := req.Params["trim"]; !ok {
		trim = true
	}
	unwrap := req.GetBool("unwrap")
	if _, ok := req.Params["unwrap"]; !ok {
		unwrap = true
	}
	outputPath := req.GetString("output_path")

	mt.Lock()
	defer mt.Unlock()

	result, err := inspect.GetView(mt.Terminal, format, trim, unwrap, outputPath)
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
