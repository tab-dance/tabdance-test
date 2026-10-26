package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List active terminals in this project (--all for every project)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		infos, err := c.Ls(lsAllOutput)
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
		if lsAllOutput {
			fmt.Fprintln(w, "ID\tCOLS\tROWS\tRECORDING\tPROJECT\tCOMMAND")
		} else {
			fmt.Fprintln(w, "ID\tCOLS\tROWS\tRECORDING\tCOMMAND")
		}
		for _, info := range infos {
			rec := "-"
			if info.Recording {
				rec = "yes"
			}
			if lsAllOutput {
				fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\t%s\n", info.ID, info.Cols, info.Rows, rec, info.Project, info.Command)
			} else {
				fmt.Fprintf(w, "%s\t%d\t%d\t%s\t%s\n", info.ID, info.Cols, info.Rows, rec, info.Command)
			}
		}
		w.Flush()
		return nil
	},
}

var lsJsonOutput bool
var lsAllOutput bool

func init() {
	lsCmd.Flags().BoolVar(&lsJsonOutput, "json", false, "output as JSON")
	lsCmd.Flags().BoolVar(&lsAllOutput, "all", false, "list terminals of all projects")
	rootCmd.AddCommand(lsCmd)
}
