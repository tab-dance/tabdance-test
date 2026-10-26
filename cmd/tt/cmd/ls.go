package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List active terminals of all projects (--mine for this shell's project only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		infos, err := c.Ls(lsMineOutput)
		if err != nil {
			return err
		}

		if lsJsonOutput {
			data, _ := json.Marshal(infos)
			fmt.Println(string(data))
			return nil
		}

		if len(infos) == 0 {
			fmt.Println("No active terminals")
			return nil
		}

		w := tabwriter.NewWriter(cmd.OutOrStdout(), 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tCOLS\tROWS\tRECORDING\tPROJECT\tCOMMAND")
		for _, info := range infos {
			rec := "-"
			if info.Recording {
				rec = "yes"
			}
			fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\t%s\n", info.ID, info.Cols, info.Rows, rec, info.Project, info.Command)
		}
		w.Flush()
		return nil
	},
}

var lsJsonOutput bool
var lsMineOutput bool

func init() {
	lsCmd.Flags().BoolVar(&lsJsonOutput, "json", false, "output as JSON")
	lsCmd.Flags().BoolVar(&lsMineOutput, "mine", false, "list only terminals of this shell's project")
	rootCmd.AddCommand(lsCmd)
}
