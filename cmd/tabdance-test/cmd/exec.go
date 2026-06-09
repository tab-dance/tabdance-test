package cmd

import (
	"fmt"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var execCmd = &cobra.Command{
	Use:   "exec command [args...]",
	Short: "Create a terminal running a command",
	Args:  cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		env := make(map[string]string)
		for _, e := range execEnv {
			k, v, _ := parseKeyValue(e)
			env[k] = v
		}

		id, err := c.Exec(args[0], execCols, execRows, execScrollback, env, execWorkingDir, execRecordPath)
		if err != nil {
			return err
		}

		if execDetach {
			fmt.Println(id)
			return nil
		}

		return c.Attach(id)
	},
}

var (
	execCols       int
	execRows       int
	execScrollback int
	execEnv        []string
	execWorkingDir string
	execRecordPath string
	execDetach     bool
)

func init() {
	execCmd.Flags().IntVar(&execCols, "cols", 108, "terminal width in columns")
	execCmd.Flags().IntVar(&execRows, "rows", 24, "terminal height in rows")
	execCmd.Flags().IntVar(&execScrollback, "scrollback", 10000, "max scrollback lines")
	execCmd.Flags().StringArrayVar(&execEnv, "env", nil, "environment variables (key=value, repeatable)")
	execCmd.Flags().StringVar(&execWorkingDir, "working-directory", "", "working directory for the child process")
	execCmd.Flags().StringVar(&execRecordPath, "record", "", "start asciinema recording to path")
	execCmd.Flags().BoolVarP(&execDetach, "detach", "d", true, "detach from terminal after creation")

	rootCmd.AddCommand(execCmd)

	carapace.Gen(execCmd).FlagCompletion(carapace.ActionMap{
		"record": carapace.ActionFiles(".cast"),
	})
	carapace.Gen(execCmd).PositionalCompletion(
		carapace.ActionExecutables(),
	)
}
