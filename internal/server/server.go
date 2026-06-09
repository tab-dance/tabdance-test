package server

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

type Server struct {
	mcp       *server.MCPServer
	terminals *terminal.TerminalManager
}

func New() *Server {
	s := &Server{
		terminals: terminal.NewTerminalManager(),
	}

	s.mcp = server.NewMCPServer(
		"tabdance-test",
		"0.1.0",
		server.WithToolCapabilities(false),
	)

	s.registerTools()

	return s
}

func (s *Server) Run() error {
	defer s.terminals.CloseAll()
	return server.ServeStdio(s.mcp)
}
