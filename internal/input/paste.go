package input

import (
	"fmt"
	"io"

	libghostty "go.mitchellh.com/libghostty"
)

func SendPaste(ptmxWriter io.Writer, term *libghostty.Terminal, text string) error {
	bracketed, _ := term.Mode(libghostty.ModeBracketedPaste)

	encoded, err := libghostty.PasteEncode([]byte(text), bracketed)
	if err != nil {
		return fmt.Errorf("encode paste: %w", err)
	}

	if encoded != nil {
		if _, err := ptmxWriter.Write(encoded); err != nil {
			return fmt.Errorf("write paste: %w", err)
		}
	}
	return nil
}
