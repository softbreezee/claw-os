// Package mcpserver provides a generic MCP (Model Context Protocol) stdio
// server framework. Callers register tools via Register() then call
// Serve() to run the read-dispatch-write loop on stdin/stdout.
//
// The server implements the JSON-RPC 2.0 transport used by MCP:
//   - initialize  → returns protocol version + server capabilities
//   - tools/list   → returns all registered tool definitions
//   - tools/call   → executes a named tool and returns content
//
// This package is separate from internal/mcp (which is the client side)
// to avoid import cycles — the MCP server may want to import pawnix's
// config/logging, while the MCP client is imported by the agent loop.
package mcpserver

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sort"
)

// ── MCP JSON-RPC types (mirrors internal/mcp for the server side) ──

type jsonRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

type jsonRPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int             `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *jsonRPCError   `json:"error,omitempty"`
}

type jsonRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type initializeResult struct {
	ProtocolVersion string             `json:"protocolVersion"`
	Capabilities    serverCapabilities `json:"capabilities"`
	ServerInfo      serverInfo         `json:"serverInfo"`
}

type serverCapabilities struct {
	Tools struct{} `json:"tools,omitempty"`
}

type serverInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

type toolDef struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema interface{} `json:"inputSchema"`
}

type toolsListResult struct {
	Tools []toolDef `json:"tools"`
}

type toolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments"`
}

type toolCallResult struct {
	Content []toolContent `json:"content"`
}

type toolContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

// ToolFunc is the callback that executes a tool. It receives raw JSON
// arguments and returns a text result. If it returns a non-nil error,
// the error text is included in the tool response as content (MCP
// convention: tools don't return RPC errors for business failures).
type ToolFunc func(args json.RawMessage) (string, error)

type registeredTool struct {
	def toolDef
	fn  ToolFunc
}

// Server is an MCP stdio server. Create one with NewServer, register
// tools with Register, then call Serve to start the loop.
type Server struct {
	name    string
	version string
	tools   map[string]registeredTool
}

// NewServer creates a new MCP server with the given identity.
func NewServer(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
		tools:   make(map[string]registeredTool),
	}
}

// Register adds a tool to the server. Name must be unique; duplicate
// registration overwrites silently (last-wins).
func (s *Server) Register(name, description string, inputSchema interface{}, fn ToolFunc) {
	s.tools[name] = registeredTool{
		def: toolDef{
			Name:        name,
			Description: description,
			InputSchema: inputSchema,
		},
		fn: fn,
	}
}

// Serve runs the stdio read-dispatch-write loop. It reads one JSON-RPC
// request per line from stdin, dispatches to the matching handler, and
// writes the response to stdout. The loop exits when stdin is closed or
// an unrecoverable I/O error occurs.
//
// Stderr is reserved for structured logging (via slog) so that log
// output doesn't interfere with the JSON-RPC wire protocol.
func (s *Server) Serve() error {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelWarn,
	})))

	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1 MB buffer
	out := os.Stdout

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			slog.Warn("mcp-server: ignoring non-JSON line", "err", err)
			continue
		}

		resp := s.dispatch(req)
		encoded, err := json.Marshal(resp)
		if err != nil {
			slog.Error("mcp-server: marshal response", "err", err)
			continue
		}

		if _, err := fmt.Fprintf(out, "%s\n", encoded); err != nil {
			return fmt.Errorf("write response: %w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	return nil
}

func (s *Server) dispatch(req jsonRPCRequest) jsonRPCResponse {
	switch req.Method {
	case "initialize":
		return s.handleInitialize(req)
	case "tools/list":
		return s.handleToolsList(req)
	case "tools/call":
		return s.handleToolsCall(req)
	default:
		return s.errorResponse(req.ID, -32601, fmt.Sprintf("unknown method: %s", req.Method))
	}
}

func (s *Server) handleInitialize(req jsonRPCRequest) jsonRPCResponse {
	result := initializeResult{
		ProtocolVersion: "2024-11-05",
		Capabilities:    serverCapabilities{Tools: struct{}{}},
		ServerInfo:      serverInfo{Name: s.name, Version: s.version},
	}
	encoded, _ := json.Marshal(result)
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  encoded,
	}
}

func (s *Server) handleToolsList(req jsonRPCRequest) jsonRPCResponse {
	defs := make([]toolDef, 0, len(s.tools))
	for _, t := range s.tools {
		defs = append(defs, t.def)
	}
	// Stable order so tool list is deterministic across calls.
	sort.Slice(defs, func(i, j int) bool { return defs[i].Name < defs[j].Name })

	result := toolsListResult{Tools: defs}
	encoded, _ := json.Marshal(result)
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  encoded,
	}
}

func (s *Server) handleToolsCall(req jsonRPCRequest) jsonRPCResponse {
	var params toolCallParams
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return s.errorResponse(req.ID, -32602, fmt.Sprintf("invalid params: %v", err))
	}

	reg, ok := s.tools[params.Name]
	if !ok {
		return s.errorResponse(req.ID, -32602, fmt.Sprintf("unknown tool: %s", params.Name))
	}

	text, err := reg.fn(params.Arguments)
	if err != nil {
		// MCP convention: tool business errors → content with error text,
		// not an RPC-level error. The LLM reads the text and decides.
		text = fmt.Sprintf("Error: %v", err)
	}

	result := toolCallResult{
		Content: []toolContent{{Type: "text", Text: text}},
	}
	encoded, _ := json.Marshal(result)
	return jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      req.ID,
		Result:  encoded,
	}
}

func (s *Server) errorResponse(id int, code int, msg string) jsonRPCResponse {
	errResp := jsonRPCResponse{
		JSONRPC: "2.0",
		ID:      id,
		Error:   &jsonRPCError{Code: code, Message: msg},
	}
	return errResp
}

// StdioReader is exposed for testing: it lets tests inject lines
// without reading from real stdin. The default Serve() always uses
// os.Stdin; test code can call this method to simulate the loop.
func (s *Server) serveReader(in io.Reader, out io.Writer) error {
	scanner := bufio.NewScanner(in)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}
		var req jsonRPCRequest
		if err := json.Unmarshal(line, &req); err != nil {
			continue
		}
		resp := s.dispatch(req)
		encoded, _ := json.Marshal(resp)
		fmt.Fprintf(out, "%s\n", encoded)
	}
	return scanner.Err()
}
