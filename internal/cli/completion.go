package cli

import (
	"github.com/carapace-sh/carapace"
)

func actionTerminalIDs() carapace.Action {
	return carapace.ActionCallback(func(c carapace.Context) carapace.Action {
		ids, err := listTerminalIDs()
		if err != nil {
			return carapace.ActionValues()
		}
		return carapace.ActionValues(ids...)
	})
}

func actionKeyNames() carapace.Action {
	return carapace.ActionValues(
		"Tab", "Enter", "Escape", "Backspace", "Delete", "Insert",
		"Home", "End", "PageUp", "PageDown", "Space",
		"ArrowUp", "ArrowDown", "ArrowLeft", "ArrowRight",
		"F1", "F2", "F3", "F4", "F5", "F6",
		"F7", "F8", "F9", "F10", "F11", "F12",
		"Ctrl+A", "Ctrl+B", "Ctrl+C", "Ctrl+D", "Ctrl+E",
		"Ctrl+F", "Ctrl+G", "Ctrl+H", "Ctrl+I", "Ctrl+J",
		"Ctrl+K", "Ctrl+L", "Ctrl+M", "Ctrl+N", "Ctrl+O",
		"Ctrl+P", "Ctrl+Q", "Ctrl+R", "Ctrl+S", "Ctrl+T",
		"Ctrl+U", "Ctrl+V", "Ctrl+W", "Ctrl+X", "Ctrl+Y",
		"Ctrl+Z",
	)
}
