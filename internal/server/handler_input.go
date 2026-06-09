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
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
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

	mt.Lock()
	defer mt.Unlock()

	if len(keyInputs) == 0 {
		return mcp.NewToolResultError("no valid keys provided"), nil
	}

	if err := input.SendKeys(mt.KeyEncoder, mt, mt.Terminal, keyInputs); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("keys sent"), nil
}

func (s *Server) handleSendText(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	text, err := req.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	if _, err := mt.WritePty([]byte(text)); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("text sent"), nil
}

func (s *Server) handleSendMouse(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	action, err := req.RequireString("action")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	button, err := req.RequireString("button")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}
	x := float32(req.GetFloat("x", 0))
	y := float32(req.GetFloat("y", 0))
	mods := req.GetString("mods", "")

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	if input.IsScrollAction(action) {
		if err := input.SendScroll(mt.MouseEncoder, mt, action, x, y, mods); err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText("scroll event sent"), nil
	}

	if err := input.SendMouse(mt.MouseEncoder, mt, input.MouseInput{
		Action: action,
		Button: button,
		X:      x,
		Y:      y,
		Mods:   mods,
	}); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("mouse event sent"), nil
}

func (s *Server) handleSendFocus(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	event, err := req.RequireString("event")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	if err := input.SendFocus(mt, event); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("focus event sent"), nil
}

func (s *Server) handleSendPaste(_ context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	id, err := req.RequireString("terminal_id")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	text, err := req.RequireString("text")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt, err := s.terminals.Get(id)
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	mt.Lock()
	defer mt.Unlock()

	if err := input.SendPaste(mt, mt.Terminal, text); err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	return mcp.NewToolResultText("paste sent"), nil
}
