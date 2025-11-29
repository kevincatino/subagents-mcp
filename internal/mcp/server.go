package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"go.uber.org/zap"

	"subagents-mcp/internal/agents"
	"subagents-mcp/internal/runner"
)

// Server handles MCP requests over stdio.
type Server struct {
	logger   *zap.Logger
	handlers *Handlers
}

func NewServer(logger *zap.Logger, repo agents.Repository, r runner.AgentRunner) *Server {
	return &Server{
		logger:   logger,
		handlers: NewHandlers(repo, r, logger),
	}
}

func (s *Server) Serve(ctx context.Context, r io.Reader, w io.Writer) error {
	dec := json.NewDecoder(bufio.NewReader(r))
	enc := json.NewEncoder(w)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var req Request
		if err := dec.Decode(&req); err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("decode request: %w", err)
		}

		resp, ok := s.handle(ctx, req)
		if !ok {
			continue
		}
		if err := enc.Encode(resp); err != nil {
			return fmt.Errorf("encode response: %w", err)
		}
	}
}

func (s *Server) handle(ctx context.Context, req Request) (Response, bool) {
	if req.JSONRPC != "2.0" {
		return errorResponse(req.ID, ErrCodeInvalidRequest, "jsonrpc must be 2.0"), true
	}

	switch req.Method {
	case "initialize":
		var params InitializeParams
		if len(req.Params) > 0 {
			if err := json.Unmarshal(req.Params, &params); err != nil {
				return errorResponse(req.ID, ErrCodeInvalidParams, "invalid initialize params"), true
			}
		}

		result := InitializeResult{
			ProtocolVersion: "2024-11-05",
			Capabilities:    map[string]any{"tools": map[string]any{}},
			ServerInfo:      ServerInfo{Name: "codex-subagents", Version: "0.1.0"},
			ClientInfo:      params.ClientInfo,
		}
		return Response{JSONRPC: "2.0", ID: req.ID, Result: result}, true
	case "notifications/initialized":
		return Response{}, false
	case "tools/list":
		return s.listTools(req.ID), true
	case "tools/call":
		return s.callTool(ctx, req), true
	default:
		return errorResponse(req.ID, ErrCodeMethodNotFound, "method not found"), true
	}
}

func (s *Server) listTools(id any) Response {
	tools := []Tool{
		{
			Name:        "list_agents",
			Description: "List all available agents with name and description.",
			InputSchema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
				"required":   []string{},
			},
		},
		{
			Name:        "delegate_task",
			Description: "Delegate a task to a specific agent with a working directory.",
			InputSchema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"agent":             map[string]any{"type": "string", "description": "Agent name to delegate to"},
					"task":              map[string]any{"type": "string", "description": "Task to be executed"},
					"working_directory": map[string]any{"type": "string", "description": "Absolute workspace path for execution"},
				},
				"required": []string{"agent", "task", "working_directory"},
			},
		},
	}
	return Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  ToolsListResult{Tools: tools},
	}
}

func (s *Server) callTool(ctx context.Context, req Request) Response {
	var params ToolsCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return errorResponse(req.ID, ErrCodeInvalidParams, "invalid params")
	}

	switch params.Name {
	case "list_agents":
		result, err := s.handlers.ListAgents(ctx)
		if err != nil {
			s.logger.Error("list_agents failed", zap.Error(err))
			return errorResponse(req.ID, ErrCodeInternal, err.Error())
		}
		return Response{JSONRPC: "2.0", ID: req.ID, Result: result}
	case "delegate_task":
		args, err := decodeArgs[delegateArgs](params.Arguments)
		if err != nil {
			return errorResponse(req.ID, ErrCodeInvalidParams, "invalid delegate_task arguments")
		}
		result, err := s.handlers.DelegateTask(ctx, args)
		if err != nil {
			s.logger.Error("delegate_task failed", zap.Error(err))
			return errorResponse(req.ID, ErrCodeInternal, err.Error())
		}
		return Response{JSONRPC: "2.0", ID: req.ID, Result: result}
	default:
		return errorResponse(req.ID, ErrCodeMethodNotFound, "tool not found")
	}
}

// NewlineDelimitedCodec ensures JSON-RPC messages remain line separated for stdio transports.
func NewlineDelimitedCodec(enc *json.Encoder) *json.Encoder {
	enc.SetEscapeHTML(false)
	return enc
}

// Debug helper to log raw requests (kept simple to avoid verbose output).
func (s *Server) logRequest(method string, params string) {
	s.logger.Debug("incoming request", zap.String("method", method), zap.String("params", strings.TrimSpace(params)))
}
