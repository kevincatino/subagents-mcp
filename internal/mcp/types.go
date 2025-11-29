package mcp

import "encoding/json"

type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      any             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params"`
}

type Response struct {
	JSONRPC string         `json:"jsonrpc"`
	ID      any            `json:"id"`
	Result  any            `json:"result,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Tool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

type ToolsListResult struct {
	Tools []Tool `json:"tools"`
}

type ToolsCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type InitializeParams struct {
	ProtocolVersion string     `json:"protocolVersion,omitempty"`
	ClientInfo      ClientInfo `json:"clientInfo,omitempty"`
}

type InitializeResult struct {
	ProtocolVersion string         `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ServerInfo      ServerInfo     `json:"serverInfo"`
	ClientInfo      ClientInfo     `json:"clientInfo,omitempty"`
}

type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}
