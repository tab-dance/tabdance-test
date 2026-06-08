package input

import (
	"fmt"
	"io"

	libghostty "go.mitchellh.com/libghostty"
)

type KeyInput struct {
	Key  string `json:"key"`
	Mods string `json:"mods,omitempty"`
	Text string `json:"text,omitempty"`
}

func SendKeys(enc *libghostty.KeyEncoder, ptmxWriter io.Writer, term *libghostty.Terminal, keys []KeyInput) error {
	for _, k := range keys {
		if k.Text != "" {
			if _, err := ptmxWriter.Write([]byte(k.Text)); err != nil {
				return fmt.Errorf("write text: %w", err)
			}
			continue
		}

		key, err := libghostty.ParseKey(k.Key)
		if err != nil {
			return fmt.Errorf("parse key %q: %w", k.Key, err)
		}

		var mods libghostty.Mods
		if k.Mods != "" {
			mods, err = libghostty.ParseMods(k.Mods)
			if err != nil {
				return fmt.Errorf("parse mods %q: %w", k.Mods, err)
			}
		}

		event, err := libghostty.NewKeyEvent()
		if err != nil {
			return fmt.Errorf("create key event: %w", err)
		}

		event.SetAction(libghostty.KeyActionPress)
		event.SetKey(key)
		event.SetMods(mods)

		seq, err := enc.Encode(event)
		event.Close()
		if err != nil {
			return fmt.Errorf("encode key: %w", err)
		}

		if seq != nil {
			if _, err := ptmxWriter.Write(seq); err != nil {
				return fmt.Errorf("write key seq: %w", err)
			}
		}

		kkf, _ := term.KittyKeyboardFlags()
		if kkf != libghostty.KittyKeyDisabled {
			releaseEvent, err := libghostty.NewKeyEvent()
			if err != nil {
				return fmt.Errorf("create release event: %w", err)
			}
			releaseEvent.SetAction(libghostty.KeyActionRelease)
			releaseEvent.SetKey(key)
			releaseEvent.SetMods(mods)

			seq, err = enc.Encode(releaseEvent)
			releaseEvent.Close()
			if err != nil {
				return fmt.Errorf("encode release: %w", err)
			}
			if seq != nil {
				if _, err := ptmxWriter.Write(seq); err != nil {
					return fmt.Errorf("write release seq: %w", err)
				}
			}
		}
	}
	return nil
}
