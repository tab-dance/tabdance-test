package cmd

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tt",
	Short: "Headless terminal testing using go-libghostty",
}

func init() {
	carapace.Gen(rootCmd)
}

func NewRootCmd() *cobra.Command {
	return rootCmd
}
