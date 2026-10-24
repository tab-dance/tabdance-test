package daemon

import (
	"github.com/tab-dance/tabdance-test/internal/input"
)

func (d *Daemon) handleSend(req Request) Response {
	if err := d.svc.SendText(req.GetString("id"), req.GetString("text")); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("text sent")
}

func (d *Daemon) handleKeys(req Request) Response {
	keys := req.GetStringSlice("keys")
	if err := d.svc.SendKeyCombos(req.GetString("id"), keys, req.GetString("mods")); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("keys sent")
}

func (d *Daemon) handleMouse(req Request) Response {
	mi := input.MouseInput{
		Action: req.GetString("action"),
		Button: req.GetString("button"),
		X:      float32(req.GetFloat("x")),
		Y:      float32(req.GetFloat("y")),
		Mods:   req.GetString("mods"),
	}
	if err := d.svc.SendMouse(req.GetString("id"), mi); err != nil {
		return ErrorResponse(err.Error())
	}
	if input.IsScrollAction(mi.Action) {
		return SuccessResponse("scroll event sent")
	}
	return SuccessResponse("mouse event sent")
}

func (d *Daemon) handlePaste(req Request) Response {
	if err := d.svc.SendPaste(req.GetString("id"), req.GetString("text")); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("paste sent")
}

func (d *Daemon) handleFocus(req Request) Response {
	if err := d.svc.SendFocus(req.GetString("id"), req.GetString("event")); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("focus event sent")
}

func (d *Daemon) handleResize(req Request) Response {
	if err := d.svc.Resize(req.GetString("id"), req.GetInt("cols"), req.GetInt("rows")); err != nil {
		return ErrorResponse(err.Error())
	}
	return SuccessResponse("resized")
}
