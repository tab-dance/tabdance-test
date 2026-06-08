package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Server struct {
	mcp *server.MCPServer
}

func New() *Server {
	s := &Server{}

	s.mcp = server.NewMCPServer(
		"tabdance-test",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	s.registerTools()

	return s
}

func (s *Server) Run() error {
	return server.ServeStdio(s.mcp)
}

func (s *Server) registerTools() {
	createTerminalTool := mcp.NewTool("create_terminal",
		mcp.WithDescription("Create a new headless terminal, optionally running a child process"),
		mcp.WithNumber("cols",
			mcp.Description("Terminal width in columns"),
		),
		mcp.WithNumber("rows",
			mcp.Description("Terminal height in rows"),
		),
		mcp.WithNumber("scrollback",
			mcp.Description("Max scrollback lines"),
		),
		mcp.WithString("command",
			mcp.Description("Command to run in the terminal"),
		),
		mcp.WithObject("env",
			mcp.Description("Environment variables (key-value pairs)"),
		),
		mcp.WithString("working_directory",
			mcp.Description("Working directory for the child process"),
		),
	)

	s.mcp.AddTool(createTerminalTool, s.handleCreateTerminal)
}

func (s *Server) handleCreateTerminal(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return mcp.NewToolResultText("not implemented"), nil
}
