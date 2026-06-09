package main

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/cmd/tabdance-test/cmd"
	"github.com/tab-dance/tabdance-test/internal/daemon"
)

func main() {
	rootCmd := cmd.NewRootCmd()

	rootCmd.Flags().Bool("daemon", false, "run as daemon (internal)")
	rootCmd.Flags().MarkHidden("daemon")

	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		if daemonMode, _ := cmd.Flags().GetBool("daemon"); daemonMode {
			d := daemon.New()
			if err := d.Run(); err != nil {
				os.Exit(1)
			}
			return
		}
		cmd.Help()
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
