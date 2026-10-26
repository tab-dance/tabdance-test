package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/pkg/actions"
)

var viewCmd = &cobra.Command{
	Use:   "view <id>",
	Short: "Print terminal screen content (--follow to watch it live)",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c := newClient()

		if viewFollow {
			// Follow streams the live screen as complete frames rendered by
			// the daemon; trim/unwrap apply per frame. The interactive vs
			// piped format choice does not apply: the daemon always renders
			// frames in VT format.
			return c.Follow(args[0], outTrim, outUnwrap)
		}

		format := effectiveViewFormat(cmd)

		result, err := c.View(args[0], format, outTrim, outUnwrap)
		if err != nil {
			return err
		}

		if viewJsonOutput {
			data, _ := json.Marshal(result)
			fmt.Println(string(data))
			return nil
		}

		fmt.Print(result.Content)
		return nil
	},
}

// effectiveViewFormat selects the default view format per surface: a
// piped stdout gets plain (deterministic; agents, tests, CI), while an
// interactive terminal replays the real screen via vt. An explicit
// --format always wins.
func effectiveViewFormat(cmd *cobra.Command) string {
	if cmd.Flags().Changed("format") || viewJsonOutput {
		return outFormat
	}
	if isTerminal(os.Stdout) {
		return "vt"
	}
	return "plain"
}

// isTerminal reports whether the file is a character device (tty, or
// unobservable sinks like /dev/null, which is fine).
func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

var viewJsonOutput bool
var viewFollow bool

func init() {
	viewCmd.Flags().BoolVar(&viewJsonOutput, "json", false, "output as JSON")
	viewCmd.Flags().BoolVar(&viewFollow, "follow", false, "watch the screen live (re-rendered complete frames; Ctrl+C to detach)")
	addOutputFlags(viewCmd, allFormats)
	viewCmd.Flags().Lookup("format").Usage = "output format: plain (piped default), vt (interactive default), html, svg, json"
	viewCmd.MarkFlagsMutuallyExclusive("follow", "json")
	viewCmd.MarkFlagsMutuallyExclusive("follow", "format")

	rootCmd.AddCommand(viewCmd)
	carapace.Gen(viewCmd).PositionalCompletion(
		actions.ActionTerminalIDs(),
	)
}
