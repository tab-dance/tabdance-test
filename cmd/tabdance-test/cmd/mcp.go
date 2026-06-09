package cmd

import (
	"github.com/spf13/cobra"
	"github.com/tab-dance/tabdance-test/internal/server"
)

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start MCP server (stdio)",
	RunE: func(cmd *cobra.Command, args []string) error {
		return server.New().Run()
	},
}

func init() {
	rootCmd.AddCommand(mcpCmd)
}
