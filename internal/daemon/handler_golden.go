package daemon

func (d *Daemon) handleVerify(req Request) Response {
	format := req.GetString("format")
	trim := req.GetBoolDefault("trim", true)
	unwrap := req.GetBoolDefault("unwrap", true)

	result, err := d.svc.Verify(req.GetString("id"), req.GetString("golden_path"), format, trim, unwrap)
	if err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse(result)
}

func (d *Daemon) handleSnapshot(req Request) Response {
	format := req.GetString("format")
	trim := req.GetBoolDefault("trim", true)
	unwrap := req.GetBoolDefault("unwrap", true)

	if err := d.svc.UpdateGolden(req.GetString("id"), req.GetString("golden_path"), format, trim, unwrap); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("golden file \"" + req.GetString("golden_path") + "\" updated")
}
