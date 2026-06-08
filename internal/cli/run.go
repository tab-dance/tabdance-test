package cli

import (
	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

func newRunCmd() *cobra.Command {
	var goldenPath string
	var update bool
	var cols, rows int
	var timeout float64
	var recordPath string
	var env []string
	var workingDir string

	cmd := &cobra.Command{
		Use:   "run <tape>",
		Short: "Play a tape file, verify against golden, exit 0/1 (CI-friendly)",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			return runRun(args[0], runOptions{
				goldenPath: goldenPath,
				update:    update,
				cols:      cols,
				rows:      rows,
				timeout:   timeout,
				recordPath: recordPath,
				env:       env,
				workingDir: workingDir,
			})
		},
	}

	cmd.Flags().StringVar(&goldenPath, "golden", "", "golden file to verify against")
	cmd.Flags().BoolVar(&update, "update", false, "update golden file instead of verifying")
	cmd.Flags().IntVar(&cols, "cols", 108, "terminal width in columns")
	cmd.Flags().IntVar(&rows, "rows", 24, "terminal height in rows")
	cmd.Flags().Float64Var(&timeout, "timeout", 30, "timeout in seconds for Wait commands")
	cmd.Flags().StringVar(&recordPath, "record", "", "record session to .cast file")
	cmd.Flags().StringArrayVar(&env, "env", nil, "environment variables (key=value, repeatable)")
	cmd.Flags().StringVar(&workingDir, "working-directory", "", "working directory for the child process")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"golden": carapace.ActionFiles(),
		"record": carapace.ActionFiles(".cast"),
	})

	carapace.Gen(cmd).PositionalCompletion(
		carapace.ActionFiles(".tape"),
	)

	return cmd
}

type runOptions struct {
	goldenPath string
	update     bool
	cols       int
	rows       int
	timeout    float64
	recordPath string
	env        []string
	workingDir string
}
