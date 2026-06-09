package daemon

import (
	"github.com/tab-dance/tabdance-test/internal/golden"
)

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
