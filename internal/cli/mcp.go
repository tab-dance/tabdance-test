package cli

import (
	"github.com/spf13/cobra"
)

func newMcpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mcp",
		Short: "Start MCP server (stdio)",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runMcp()
		},
	}
	return cmd
}
