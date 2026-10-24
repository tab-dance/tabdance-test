package server

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/tab-dance/tabdance-test/internal/input"
)

func (s *Server) handleSendKeys(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}

	var keyInputs []input.KeyInput
	target := struct {
		Keys []input.KeyInput `json:"keys"`
	}{}
	if err := req.BindArguments(&target); err == nil {
		keyInputs = target.Keys
	} else {
		args := req.GetArguments()
		if raw, ok := args["keys"]; ok {
			data, _ := json.Marshal(raw)
			if err := json.Unmarshal(data, &keyInputs); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("parse keys: %v", err)), nil
			}
		}
	}

	if err := s.svc.SendKeys(id, keyInputs); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText("keys sent"), nil
}

func (s *Server) handleSendText(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	text, err := req.RequireString("text")
	if err != nil {
		return handleParamError(err)
	}
	if err := s.svc.SendText(id, text); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText("text sent"), nil
}

func (s *Server) handleSendMouse(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	action, err := req.RequireString("action")
	if err != nil {
		return handleParamError(err)
	}
	button := req.GetString("button", "left")
	x := float32(req.GetFloat("x", 0))
	y := float32(req.GetFloat("y", 0))
	mods := req.GetString("mods", "")

	if err := s.svc.SendMouse(id, input.MouseInput{
		Action: action,
		Button: button,
		X:      x,
		Y:      y,
		Mods:   mods,
	}); err != nil {
		return handleParamError(err)
	}
	if input.IsScrollAction(action) {
		return mcp.NewToolResultText("scroll event sent"), nil
	}
	return mcp.NewToolResultText("mouse event sent"), nil
}

func (s *Server) handleSendFocus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	event, err := req.RequireString("event")
	if err != nil {
		return handleParamError(err)
	}
	if err := s.svc.SendFocus(id, event); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText("focus event sent"), nil
}

func (s *Server) handleSendPaste(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return handleParamError(err)
	}
	text, err := req.RequireString("text")
	if err != nil {
		return handleParamError(err)
	}
	if err := s.svc.SendPaste(id, text); err != nil {
		return handleParamError(err)
	}
	return mcp.NewToolResultText("paste sent"), nil
}
