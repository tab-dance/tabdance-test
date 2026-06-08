package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func NewRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tabdance-test",
		Short: "Headless terminal testing using go-libghostty",
	}

	cmd.AddCommand(
		newMcpCmd(),
		newExecCmd(),
		newLsCmd(),
		newInspectCmd(),
		newAttachCmd(),
		newKillCmd(),
		newRmCmd(),
		newViewCmd(),
		newCursorCmd(),
		newSizeCmd(),
		newCellCmd(),
		newSendCmd(),
		newKeysCmd(),
		newMouseCmd(),
		newPasteCmd(),
		newFocusCmd(),
		newResizeCmd(),
		newTapeCmd(),
		newWaitCmd(),
		newVerifyCmd(),
		newSnapshotCmd(),
		newRecordCmd(),
		newStopCmd(),
		newMarkerCmd(),
		newRunCmd(),
	)

	carapace.Gen(cmd)

	return cmd
}
