package input

import (
	"fmt"
	"io"

	libghostty "go.mitchellh.com/libghostty"
)

func SendFocus(ptmxWriter io.Writer, event string) error {
	fe, err := libghostty.ParseFocusEvent(event)
	if err != nil {
		return fmt.Errorf("parse focus event %q: %w", event, err)
	}

	seq, err := libghostty.FocusEncode(fe)
	if err != nil {
		return fmt.Errorf("encode focus: %w", err)
	}

	if seq != nil {
		if _, err := ptmxWriter.Write(seq); err != nil {
			return fmt.Errorf("write focus seq: %w", err)
		}
	}
	return nil
}
