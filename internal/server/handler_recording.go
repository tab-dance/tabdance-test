package server

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
)

func (s *Server) handleStartRecording(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	path, err := req.RequireString("output_path")
	if err != nil {
		return handleParamError(err)
	}
	idle := req.GetFloat("idle_time_limit", 2.0)

	if err := s.svc.StartRecording(id, path, idle); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText("recording started"), nil
}

func (s *Server) handleStopRecording(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	if err := s.svc.StopRecording(id); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText("recording stopped"), nil
}

func (s *Server) handleAddMarker(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	label, err := req.RequireString("label")
	if err != nil {
		return handleParamError(err)
	}
	if err := s.svc.AddMarker(id, label); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText("marker added"), nil
}
