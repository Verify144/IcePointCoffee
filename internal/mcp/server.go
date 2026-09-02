package mcp

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Verify144/IcePointCoffee/internal/ai"
)

// ToolExecutor 工具执行接口（兼容 ai.Tool）
type ToolExecutor interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(ctx context.Context, args json.RawMessage) (interface{}, error)
}

// Server MCP 服务器
type Server struct {
	name        string
	version     string
	toolSource  func() []ToolExecutor
	resources   []Resource
	initialized bool
}

// NewServer 创建 MCP 服务器
func NewServer(name, version string) *Server {
	return &Server{
		name:    name,
		version: version,
		resources: []Resource{
			{
				URI:      "icepoint://server/info",
				Name:     "Server Information",
				Description: "IcePointCoffee server connection info",
				MimeType: "application/json",
			},
		},
	}
}

// SetToolSource 设置工具来源（可动态注入）
func (s *Server) SetToolSource(src func() []ToolExecutor) {
	s.toolSource = src
}

// Handle 处理 JSON-RPC 请求
func (s *Server) Handle(ctx context.Context, data []byte) []byte {
	var req Request
	if err := json.Unmarshal(data, &req); err != nil {
		return marshalError(nil, ErrCodeParseError, "Parse error: "+err.Error())
	}

	// Notification (无 ID)：不需要响应
	if len(req.ID) == 0 || string(req.ID) == "null" {
		s.handleNotification(ctx, &req)
		return nil
	}

	switch req.Method {
	case MethodInitialize:
		return s.handleInitialize(ctx, &req)
	case MethodToolsList:
		return s.handleToolsList(ctx, &req)
	case MethodToolsCall:
		return s.handleToolsCall(ctx, &req)
	case MethodResourcesList:
		return s.handleResourcesList(ctx, &req)
	case MethodPing:
		return marshalResult(req.ID, PingResult{})
	default:
		return marshalError(req.ID, ErrCodeMethodNotFound, "Method not found: "+req.Method)
	}
}

// handleNotification 处理 notification（无响应）
func (s *Server) handleNotification(ctx context.Context, req *Request) {
	switch req.Method {
	case MethodNotificationsInitialized:
		s.initialized = true
	}
}

// handleInitialize 处理 initialize
func (s *Server) handleInitialize(ctx context.Context, req *Request) []byte {
	var params InitializeParams
	if len(req.Params) > 0 {
		if err := json.Unmarshal(req.Params, &params); err != nil {
			return marshalError(req.ID, ErrCodeInvalidParams, "Invalid params: "+err.Error())
		}
	}

	result := InitializeResult{
		ProtocolVersion: ProtocolVersion,
		Capabilities: ServerCap{
			Tools:     &ToolsCap{ListChanged: true},
			Resources: &ResourcesCap{ListChanged: true},
		},
		ServerInfo: ServerInfo{
			Name:    s.name,
			Version: s.version,
		},
		Instructions: "IcePointCoffee MCP server. Provides MC server control tools (28 tools: mc_command, mc_perceive, mc_teleport, etc.)",
	}
	return marshalResult(req.ID, result)
}

// handleToolsList 处理 tools/list
func (s *Server) handleToolsList(ctx context.Context, req *Request) []byte {
	if s.toolSource == nil {
		return marshalResult(req.ID, ToolListResult{Tools: []Tool{}})
	}

	executors := s.toolSource()
	tools := make([]Tool, 0, len(executors))
	for _, t := range executors {
		params := t.Parameters()
		paramsJSON, _ := json.Marshal(params)
		tools = append(tools, Tool{
			Name:        t.Name(),
			Title:       t.Name(),
			Description: t.Description(),
			InputSchema: paramsJSON,
		})
	}
	return marshalResult(req.ID, ToolListResult{Tools: tools})
}

// handleToolsCall 处理 tools/call
func (s *Server) handleToolsCall(ctx context.Context, req *Request) []byte {
	var params ToolCallParams
	if len(req.Params) == 0 {
		return marshalError(req.ID, ErrCodeInvalidParams, "Missing params")
	}
	if err := json.Unmarshal(req.Params, &params); err != nil {
		return marshalError(req.ID, ErrCodeInvalidParams, "Invalid params: "+err.Error())
	}

	if s.toolSource == nil {
		return marshalError(req.ID, ErrCodeInternalError, "No tools available")
	}

	// 找工具
	var found ToolExecutor
	for _, t := range s.toolSource() {
		if t.Name() == params.Name {
			found = t
			break
		}
	}
	if found == nil {
		return marshalError(req.ID, ErrCodeInvalidParams, "Unknown tool: "+params.Name)
	}

	// 执行
	result, err := found.Execute(ctx, params.Arguments)
	if err != nil {
		text, _ := json.Marshal(result)
		if len(text) == 0 {
			text = []byte("Tool error: " + err.Error())
		}
		return marshalResult(req.ID, ToolCallResult{
			Content: []ContentBlock{{
				Type: "text",
				Text: string(text),
			}},
			IsError: true,
		})
	}

	// 成功
	text, _ := json.MarshalIndent(result, "", "  ")
	return marshalResult(req.ID, ToolCallResult{
		Content: []ContentBlock{{
			Type: "text",
			Text: string(text),
		}},
	})
}

// handleResourcesList 处理 resources/list
func (s *Server) handleResourcesList(ctx context.Context, req *Request) []byte {
	return marshalResult(req.ID, ResourceListResult{Resources: s.resources})
}

// --- helpers ---

func marshalResult(id json.RawMessage, result interface{}) []byte {
	resp := NewResult(id, result)
	data, _ := json.Marshal(resp)
	return data
}

func marshalError(id json.RawMessage, code int, msg string) []byte {
	resp := NewError(id, code, msg)
	data, _ := json.Marshal(resp)
	return data
}

// 编译器断言
var _ = fmt.Sprintf
var _ = ai.StreamChunk{} // 避免 import 警告
