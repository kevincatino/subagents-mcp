# API (MCP)

Protocol: JSON-RPC 2.0, MCP version `2024-11-05`.

## Methods
- `initialize`
  - Request: `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"clientInfo":{"name":"my-client","version":"1.0.0"}}}`
  - Result: `{"protocolVersion":"2024-11-05","capabilities":{"tools":{}},"serverInfo":{"name":"codex-subagents","version":"0.1.0"},"clientInfo":{"name":"my-client","version":"1.0.0"}}`
- `tools/list`
  - Request: `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
  - Result: tools array with schemas for `list_agents` and `delegate_task`; optional `nextCursor` not used.
- `tools/call`
  - Params: `{"name": string, "arguments"?: object}`
  - Result: varies by tool; errors returned as JSON-RPC errors (no `isError` field is used).

## Tools
- `list_agents`
  - Input schema: `{ "type": "object", "properties": {} }`
  - Call example: `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"list_agents"}}`
  - Success result: `{"content":[{"type":"text","text":"{\"agents\":[{\"name\":\"docs-fetcher\",\"description\":\"Docs excerpt fetcher\"}]}"}]}`
- `delegate_task`
  - Input schema: object with required `agent`, `task`, `working_directory` (strings).
  - Call example:
    ```json
    {
      "jsonrpc": "2.0",
      "id": 4,
      "method": "tools/call",
      "params": {
        "name": "delegate_task",
        "arguments": {
          "agent": "docs-fetcher",
          "task": "summarize latest release notes",
          "working_directory": "/abs/workspace"
        }
      }
    }
    ```
  - Success result: `{"content":[{"type":"text","text":"<final output from runner>"}]}`

## Errors
- Protocol/validation errors return JSON-RPC `error` with codes:
  - `-32600` invalid request (e.g., wrong `jsonrpc`)
  - `-32601` method not found (unknown method or tool)
  - `-32602` invalid params (bad arguments)
  - `-32603` internal error (handler failures)
- Tool execution errors surface as JSON-RPC errors; tool-level `isError` is not used by this server.
