package server

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/mark3labs/mcp-go/mcp"
)

func TestCreateAndCloseTerminal(t *testing.T) {
	s := New()

	result, err := s.handleCreateTerminal(context.Background(), makeRequest(map[string]any{
		"cols": 40,
		"rows": 10,
	}))
	if err != nil {
		t.Fatalf("create terminal: %v", err)
	}

	var createResp struct {
		TerminalID string `json:"terminal_id"`
	}
	if err := parseResult(result, &createResp); err != nil {
		t.Fatalf("parse create result: %v", err)
	}
	if createResp.TerminalID == "" {
		t.Fatal("expected terminal_id")
	}

	_, err = s.handleCloseTerminal(context.Background(), makeRequest(map[string]any{
		"terminal_id": createResp.TerminalID,
	}))
	if err != nil {
		t.Fatalf("close terminal: %v", err)
	}
}

func TestListTerminals(t *testing.T) {
	s := New()

	_, _ = s.handleCreateTerminal(context.Background(), makeRequest(map[string]any{
		"cols": 80, "rows": 24,
	}))
	_, _ = s.handleCreateTerminal(context.Background(), makeRequest(map[string]any{
		"cols": 40, "rows": 10,
	}))

	result, err := s.handleListTerminals(context.Background(), mcp.CallToolRequest{})
	if err != nil {
		t.Fatalf("list terminals: %v", err)
	}

	var listResp struct {
		Terminals []struct {
			ID string `json:"id"`
		} `json:"terminals"`
	}
	if err := parseResult(result, &listResp); err != nil {
		t.Fatalf("parse list result: %v", err)
	}
	if len(listResp.Terminals) != 2 {
		t.Fatalf("expected 2 terminals, got %d", len(listResp.Terminals))
	}
}

func TestGetView(t *testing.T) {
	s := New()

	result, _ := s.handleCreateTerminal(context.Background(), makeRequest(map[string]any{
		"cols": 80, "rows": 24,
	}))
	var createResp struct {
		TerminalID string `json:"terminal_id"`
	}
	parseResult(result, &createResp)

	viewResult, err := s.handleGetView(context.Background(), makeRequest(map[string]any{
		"terminal_id": createResp.TerminalID,
		"format":      "plain",
	}))
	if err != nil {
		t.Fatalf("get view: %v", err)
	}
	if viewResult == nil {
		t.Fatal("expected view result")
	}
}

func TestGetCursor(t *testing.T) {
	s := New()

	result, _ := s.handleCreateTerminal(context.Background(), makeRequest(map[string]any{
		"cols": 80, "rows": 24,
	}))
	var createResp struct {
		TerminalID string `json:"terminal_id"`
	}
	parseResult(result, &createResp)

	cursorResult, err := s.handleGetCursor(context.Background(), makeRequest(map[string]any{
		"terminal_id": createResp.TerminalID,
	}))
	if err != nil {
		t.Fatalf("get cursor: %v", err)
	}
	if cursorResult == nil {
		t.Fatal("expected cursor result")
	}
}

func TestGetSize(t *testing.T) {
	s := New()

	result, _ := s.handleCreateTerminal(context.Background(), makeRequest(map[string]any{
		"cols": 120, "rows": 30,
	}))
	var createResp struct {
		TerminalID string `json:"terminal_id"`
	}
	parseResult(result, &createResp)

	sizeResult, err := s.handleGetSize(context.Background(), makeRequest(map[string]any{
		"terminal_id": createResp.TerminalID,
	}))
	if err != nil {
		t.Fatalf("get size: %v", err)
	}

	var sizeResp struct {
		Cols uint16 `json:"cols"`
		Rows uint16 `json:"rows"`
	}
	if err := parseResult(sizeResult, &sizeResp); err != nil {
		t.Fatalf("parse size result: %v", err)
	}
	if sizeResp.Cols != 120 || sizeResp.Rows != 30 {
		t.Fatalf("expected 120x30, got %dx%d", sizeResp.Cols, sizeResp.Rows)
	}
}

func TestResize(t *testing.T) {
	s := New()

	result, _ := s.handleCreateTerminal(context.Background(), makeRequest(map[string]any{
		"cols": 80, "rows": 24,
	}))
	var createResp struct {
		TerminalID string `json:"terminal_id"`
	}
	parseResult(result, &createResp)

	_, err := s.handleResize(context.Background(), makeRequest(map[string]any{
		"terminal_id": createResp.TerminalID,
		"cols":        100,
		"rows":        40,
	}))
	if err != nil {
		t.Fatalf("resize: %v", err)
	}

	sizeResult, _ := s.handleGetSize(context.Background(), makeRequest(map[string]any{
		"terminal_id": createResp.TerminalID,
	}))
	var sizeResp struct {
		Cols uint16 `json:"cols"`
		Rows uint16 `json:"rows"`
	}
	parseResult(sizeResult, &sizeResp)
	if sizeResp.Cols != 100 || sizeResp.Rows != 40 {
		t.Fatalf("expected 100x40 after resize, got %dx%d", sizeResp.Cols, sizeResp.Rows)
	}
}

func TestInvalidTerminalID(t *testing.T) {
	s := New()

	result, err := s.handleGetView(context.Background(), makeRequest(map[string]any{
		"terminal_id": "nonexistent",
	}))
	if err != nil {
		t.Fatalf("unexpected error type: %v", err)
	}
	if len(result.Content) > 0 {
		if tc, ok := result.Content[0].(mcp.TextContent); ok {
			if tc.Text == "" || result.IsError {
				return
			}
		}
	}
	t.Fatal("expected error result for invalid terminal_id")
}

func TestVerifyOutputNoGolden(t *testing.T) {
	s := New()

	result, _ := s.handleCreateTerminal(context.Background(), makeRequest(map[string]any{
		"cols": 80, "rows": 24,
	}))
	var createResp struct {
		TerminalID string `json:"terminal_id"`
	}
	parseResult(result, &createResp)

	verifyResult, err := s.handleVerifyOutput(context.Background(), makeRequest(map[string]any{
		"terminal_id": createResp.TerminalID,
		"golden_path": "/tmp/tabdance-test-nonexistent-golden.txt",
	}))
	if err != nil {
		t.Fatalf("verify output: %v", err)
	}

	var verifyResp struct {
		Match bool `json:"match"`
	}
	if err := parseResult(verifyResult, &verifyResp); err != nil {
		t.Fatalf("parse verify result: %v", err)
	}
	if verifyResp.Match {
		t.Fatal("expected no match for nonexistent golden file")
	}
}

func makeRequest(args map[string]any) mcp.CallToolRequest {
	return mcp.CallToolRequest{
		Params: mcp.CallToolParams{
			Arguments: args,
		},
	}
}

func parseResult(result *mcp.CallToolResult, v any) error {
	if len(result.Content) == 0 {
		return nil
	}
	tc, ok := result.Content[0].(mcp.TextContent)
	if !ok {
		return nil
	}
	return json.Unmarshal([]byte(tc.Text), v)
}
