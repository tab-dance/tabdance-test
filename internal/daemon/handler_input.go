package daemon

import (
	"github.com/tab-dance/tabdance-test/internal/input"
)

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
