package server

import (
	"github.com/mark3labs/mcp-go/server"
	"github.com/tab-dance/tabdance-test/internal/service"
	"github.com/tab-dance/tabdance-test/internal/terminal"
)

type Server struct {
	mcp       *server.MCPServer
	terminals *terminal.TerminalManager
	svc       *service.Service
	version   string
}

func New(version string) *Server {
	manager := terminal.NewTerminalManager()
	s := &Server{
		version:   version,
		terminals: manager,
		svc:       service.NewWithManager(manager),
	}

	s.mcp = server.NewMCPServer(
		"tabdance-test",
		version,
		server.WithToolCapabilities(false),
	)

	s.registerTools()

	return s
}

func (s *Server) Run() error {
	defer s.terminals.CloseAll()
	return server.ServeStdio(s.mcp)
}
