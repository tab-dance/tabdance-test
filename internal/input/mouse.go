package input

import (
	"fmt"
	"io"

	libghostty "go.mitchellh.com/libghostty"
)

type MouseInput struct {
	Action string  `json:"action"`
	Button string  `json:"button"`
	X      float32 `json:"x"`
	Y      float32 `json:"y"`
	Mods   string  `json:"mods,omitempty"`
}

func SendMouse(enc *libghostty.MouseEncoder, ptmxWriter io.Writer, input MouseInput) error {
	event, err := libghostty.NewMouseEvent()
	if err != nil {
		return fmt.Errorf("create mouse event: %w", err)
	}

	action, err := parseMouseAction(input.Action)
	if err != nil {
		return err
	}
	event.SetAction(action)

	button, err := libghostty.ParseMouseButton(input.Button)
	if err != nil {
		return fmt.Errorf("parse button %q: %w", input.Button, err)
	}
	event.SetButton(button)

	if input.Mods != "" {
		mods, err := libghostty.ParseMods(input.Mods)
		if err != nil {
			return fmt.Errorf("parse mods %q: %w", input.Mods, err)
		}
		event.SetMods(mods)
	}

	event.SetPosition(libghostty.MousePosition{X: input.X, Y: input.Y})

	seq, err := enc.Encode(event)
	event.Close()
	if err != nil {
		return fmt.Errorf("encode mouse: %w", err)
	}

	if seq != nil {
		if _, err := ptmxWriter.Write(seq); err != nil {
			return fmt.Errorf("write mouse seq: %w", err)
		}
	}
	return nil
}

func parseMouseAction(s string) (libghostty.MouseAction, error) {
	switch s {
	case "press":
		return libghostty.MouseActionPress, nil
	case "release":
		return libghostty.MouseActionRelease, nil
	case "move", "motion":
		return libghostty.MouseActionMotion, nil
	default:
		return 0, fmt.Errorf("unknown mouse action %q", s)
	}
}
