package cmd

import (
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "tt",
	Short: "Headless terminal testing",
}

func init() {
	// Completions are carapace-space, not cobra space: the generated
	// "completion" command would advertise a static-script generator that
	// cannot serve the dynamic terminal-ID actions at all. Drop it and the
	// auto-added help command (still available, just hidden; --help/-h
	// keep working) to keep the subcommand surface honest.
	rootCmd.CompletionOptions = cobra.CompletionOptions{
		DisableDefaultCmd: true,
	}
	rootCmd.SetHelpCommand(&cobra.Command{
		Use:    "help [command]",
		Short:  "Help about any command",
		Hidden: true,
		RunE: func(c *cobra.Command, args []string) error {
			cmd, _, err := rootCmd.Find(args)
			if cmd == nil || err != nil {
				_ = rootCmd.Help()
				return nil
			}
			cmd.InitDefaultHelpFlag()
			return cmd.Help()
		},
	})
	// Cobra's usage template special-cases the help command and lists it
	// even when hidden; without the special case the surface is honest.
	rootCmd.SetUsageTemplate(strings.Replace(
		rootCmd.UsageTemplate(),
		`{{if (or .IsAvailableCommand (eq .Name "help"))}}`,
		`{{if .IsAvailableCommand}}`,
		1,
	))

	carapace.Gen(rootCmd)
}

func NewRootCmd() *cobra.Command {
	return rootCmd
}
