package daemon

import (
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/tab-dance/tabdance-test/internal/inspect"
)

// followDebounce coalesces bursts of PTY output into one rendered frame
// per tick, so follow tracks the screen at a steady cadence instead of
// rendering per output byte.
const followDebounce = 50 * time.Millisecond

// handleFollow streams live-view frames to the connection: the full
// screen (modes, style, cursor) re-rendered after each output burst.
// Unlike attach this is stateful by construction - every frame carries
// the complete display state - so a mid-session attach never garbles.
// It is transport-specific (streaming) and stays out of the service
// beyond the per-frame render, which FollowFrame owns.
func (d *Daemon) handleFollow(req Request, conn net.Conn) {
	id := req.GetString("id")
	trim := req.GetBoolDefault("trim", true)
	unwrap := req.GetBoolDefault("unwrap", true)

	mt, err := d.manager.Get(id)
	if err != nil {
		json.NewEncoder(conn).Encode(ErrorResponse(err.Error()))
		return
	}

	sub := mt.Subscribe()
	defer mt.Unsubscribe(sub)

	state := &inspect.FrameState{}
	renderAndSend := func() bool {
		frame, changed, err := d.svc.FollowFrame(id, state, trim, unwrap)
		if err != nil {
			json.NewEncoder(conn).Encode(ErrorResponse(err.Error()))
			return false
		}
		if !changed {
			return true
		}
		payload, err := json.Marshal(map[string]string{"content": frame})
		if err != nil {
			return false
		}
		response, err := json.Marshal(SuccessResponse(json.RawMessage(payload)))
		if err != nil {
			return false
		}
		_, err = fmt.Fprintf(conn, "%s\n", response)
		return err == nil
	}

	// First frame straight away (this is the mid-session "initial view"),
	// then render once per debounce window that saw output. Idle
	// terminals render nothing, and any screen change (including
	// alternate-screen transitions) is always preceded by output, so a
	// fixed cadence is enough for consistency.
	if !renderAndSend() {
		return
	}
	ticker := time.NewTicker(followDebounce)
	defer ticker.Stop()
	pending := false
	for {
		select {
		case _, ok := <-sub.Channel():
			if !ok {
				return
			}
			pending = true
		case <-ticker.C:
		}
		if !pending {
			continue
		}
		pending = false
		if !renderAndSend() {
			return
		}
	}
}
