package cmd

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/spf13/cobra"
)

var lsCmd = &cobra.Command{
	Use:   "ls",
	Short: "List active terminals",
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()
		infos, err := c.Ls()
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
		fmt.Fprintln(w, "ID\tCOLS\tROWS\tCOMMAND")
		for _, info := range infos {
			fmt.Fprintf(w, "%s\t%d\t%d\t%s\n", info.ID, info.Cols, info.Rows, info.Command)
		}
		w.Flush()
		return nil
	},
}

var lsJsonOutput bool

func init() {
	lsCmd.Flags().BoolVar(&lsJsonOutput, "json", false, "output as JSON")
	rootCmd.AddCommand(lsCmd)
}
