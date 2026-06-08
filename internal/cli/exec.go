package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newExecCmd() *cobra.Command {
	var cols, rows int
	var scrollback int
	var env []string
	var workingDir string
	var recordPath string
	var detach bool

	cmd := &cobra.Command{
		Use:   "exec command [args...]",
		Short: "Create a terminal running a command",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runExec(execOptions{
				command:   args[0],
				cols:      cols,
				rows:      rows,
				scrollback: scrollback,
				env:       env,
				workingDir: workingDir,
				recordPath: recordPath,
				detach:    detach,
			})
		},
	}

	cmd.Flags().IntVar(&cols, "cols", 108, "terminal width in columns")
	cmd.Flags().IntVar(&rows, "rows", 24, "terminal height in rows")
	cmd.Flags().IntVar(&scrollback, "scrollback", 10000, "max scrollback lines")
	cmd.Flags().StringArrayVar(&env, "env", nil, "environment variables (key=value, repeatable)")
	cmd.Flags().StringVar(&workingDir, "working-directory", "", "working directory for the child process")
	cmd.Flags().StringVar(&recordPath, "record", "", "start asciinema recording to path")
	cmd.Flags().BoolVarP(&detach, "detach", "d", true, "detach from terminal after creation")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"record": carapace.ActionFiles(".cast"),
	})

	carapace.Gen(cmd).PositionalCompletion(
		carapace.ActionExecutables(),
	)

	return cmd
}

type execOptions struct {
	command    string
	cols       int
	rows       int
	scrollback int
	env        []string
	workingDir string
	recordPath string
	detach     bool
}
