package daemon

func (d *Daemon) handleView(req Request) Response {
	format := req.GetString("format")
	trim := req.GetBoolDefault("trim", true)
	unwrap := req.GetBoolDefault("unwrap", true)

	result, err := d.svc.View(req.GetString("id"), format, trim, unwrap)
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleCursor(req Request) Response {
	result, err := d.svc.Cursor(req.GetString("id"))
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleSize(req Request) Response {
	result, err := d.svc.Size(req.GetString("id"))
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleCell(req Request) Response {
	result, err := d.svc.Cell(req.GetString("id"), req.GetInt("x"), req.GetInt("y"))
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleOutput(req Request) Response {
	n := req.GetInt("bytes")
	if n <= 0 {
		n = 4096
	}
	out, err := d.svc.RecentOutput(req.GetString("id"), n)
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(map[string]any{"output": out})
}
