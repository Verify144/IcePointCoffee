package mcp

import "encoding/json"

// ---- MCP 协议版本 ----

const ProtocolVersion = "2025-06-18"

// ---- JSON-RPC 2.0 基础 ----

// Request JSON-RPC 2.0 请求
type Request struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Method  string          `json:"method"`
	Params  json.RawMessage `json:"params,omitempty"`
}

// Response JSON-RPC 2.0 响应
type Response struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      json.RawMessage `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *Error          `json:"error,omitempty"`
}

// Error JSON-RPC 2.0 错误
type Error struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data,omitempty"`
}

// Error codes
const (
	ErrCodeParseError     = -32700
	ErrCodeInvalidRequest = -32600
	ErrCodeMethodNotFound = -32601
	ErrCodeInvalidParams  = -32602
	ErrCodeInternalError = -32603
)

// ---- MCP 方法名 ----

const (
	MethodInitialize              = "initialize"
	MethodNotificationsInitialized = "notifications/initialized"
	MethodToolsList              = "tools/list"
	MethodToolsCall             = "tools/call"
	MethodResourcesList         = "resources/list"
	MethodResourcesRead         = "resources/read"
	MethodResourcesSubscribe    = "resources/subscribe"
	MethodPing                  = "ping"
)

// ---- Initialize ----

// InitializeParams initialize 请求参数
type InitializeParams struct {
	ProtocolVersion string       `json:"protocolVersion"`
	Capabilities    ClientCap   `json:"capabilities"`
	ClientInfo     ClientInfo  `json:"clientInfo"`
}

// ClientCap 客户端能力
type ClientCap struct {
	Roots       *struct{}    `json:"roots,omitempty"`
	Sampling    *struct{}    `json:"sampling,omitempty"`
	Elicitation *struct{}    `json:"elicitation,omitempty"`
}

// ClientInfo 客户端信息
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// InitializeResult initialize 结果
type InitializeResult struct {
	ProtocolVersion string        `json:"protocolVersion"`
	Capabilities    ServerCap    `json:"capabilities"`
	ServerInfo     ServerInfo   `json:"serverInfo"`
	Instructions   string       `json:"instructions,omitempty"`
}

// ServerCap 服务端能力
type ServerCap struct {
	Tools       *ToolsCap       `json:"tools,omitempty"`
	Resources   *ResourcesCap   `json:"resources,omitempty"`
	Prompts    *struct{}      `json:"prompts,omitempty"`
	Logging    *struct{}      `json:"logging,omitempty"`
}

// ToolsCap 工具能力
type ToolsCap struct {
	ListChanged bool `json:"listChanged,omitempty"`
}

// ResourcesCap 资源能力
type ResourcesCap struct {
	Subscribe   bool `json:"subscribe,omitempty"`
	ListChanged bool `json:"listChanged,omitempty"`
}

// ServerInfo 服务信息
type ServerInfo struct {
	Name    string `json:"name"`
	Version string `json:"version,omitempty"`
}

// ---- Tools ----

// ToolCallParams tools/call 请求参数
type ToolCallParams struct {
	Name      string          `json:"name"`
	Arguments json.RawMessage `json:"arguments,omitempty"`
}

// ToolCallResult tools/call 结果
type ToolCallResult struct {
	Content       []ContentBlock `json:"content"`
	IsError      bool           `json:"isError,omitempty"`
}

// ContentBlock 内容块
type ContentBlock struct {
	Type     string `json:"type"` // text | image | audio | resource | embedded
	Text     string `json:"text,omitempty"`
	MimeType string `json:"mimeType,omitempty"`
	Data     string `json:"data,omitempty"`      // base64
	URI      string `json:"uri,omitempty"`      // for resource
	Name     string `json:"name,omitempty"`
}

// ToolListResult tools/list 结果
type ToolListResult struct {
	Tools      []Tool   `json:"tools"`
	NextCursor string   `json:"nextCursor,omitempty"`
}

// Tool MCP 工具定义
type Tool struct {
	Name        string          `json:"name"`
	Title       string          `json:"title,omitempty"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"inputSchema"`
}

// ---- Resources ----

// ResourceListResult resources/list 结果
type ResourceListResult struct {
	Resources   []Resource `json:"resources"`
	NextCursor string     `json:"nextCursor,omitempty"`
}

// Resource 资源定义
type Resource struct {
	URI         string            `json:"uri"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	MimeType    string            `json:"mimeType,omitempty"`
}

// ---- Ping ----

// PingResult ping 结果
type PingResult struct{}

// NewError 构造错误响应
func NewError(id json.RawMessage, code int, msg string) *Response {
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Error: &Error{
			Code:    code,
			Message: msg,
		},
	}
}

// NewResult 构造成功响应
func NewResult(id json.RawMessage, result interface{}) *Response {
	data, _ := json.Marshal(result)
	return &Response{
		JSONRPC: "2.0",
		ID:      id,
		Result:  data,
	}
}

// ParseID 将 json.RawMessage 解析为 int 或 string ID
func ParseID(id json.RawMessage) interface{} {
	// Try int first
	var i int
	if err := json.Unmarshal(id, &i); err == nil {
		return i
	}
	// Fall back to string
	var s string
	if err := json.Unmarshal(id, &s); err == nil {
		return s
	}
	return nil
}
