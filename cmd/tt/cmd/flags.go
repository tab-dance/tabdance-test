package cmd

import (
	"strings"

	"github.com/carapace-sh/carapace"
	"github.com/spf13/cobra"
)

var (
	outFormat string
	outTrim   bool
	outUnwrap bool
)

var allFormats = []string{"plain", "vt", "html", "svg", "json"}
var textFormats = []string{"plain", "vt", "html"}

// addOutputFlags declares the shared --format/--trim/--unwrap flag group
// on a command and wires carapace completion for the allowed formats.
func addOutputFlags(cmd *cobra.Command, formats []string) {
	cmd.Flags().StringVar(&outFormat, "format", "plain", "output format: "+strings.Join(formats, ", "))
	cmd.Flags().BoolVar(&outTrim, "trim", true, "trim trailing whitespace")
	cmd.Flags().BoolVar(&outUnwrap, "unwrap", true, "unwrap soft-wrapped lines")

	carapace.Gen(cmd).FlagCompletion(carapace.ActionMap{
		"format": carapace.ActionValues(formats...),
	})
}
