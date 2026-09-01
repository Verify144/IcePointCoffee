package server

// APIEndpoint API 端点定义
type APIEndpoint struct {
	Method      string      `json:"method"`
	Path        string      `json:"path"`
	Description string      `json:"description"`
	Request     interface{} `json:"request,omitempty"`
	Response    interface{} `json:"response,omitempty"`
	Example     string      `json:"example,omitempty"`
}

// APIDoc API 文档
type APIDoc struct {
	Version  string        `json:"version"`
	Title    string        `json:"title"`
	BaseURL  string        `json:"base_url"`
	Endpoints []APIEndpoint `json:"endpoints"`
}

// GetAPIDoc 返回 API 文档
func GetAPIDoc() *APIDoc {
	return &APIDoc{
		Version: "1.0",
		Title:   "IcePoint Coffee API",
		BaseURL: "/api/v1",
		Endpoints: []APIEndpoint{
			{
				Method:      "GET",
				Path:        "/health",
				Description: "健康检查",
				Example:     `curl http://localhost:8080/health`,
			},
			{
				Method:      "GET",
				Path:        "/status",
				Description: "系统状态",
			},
			{
				Method:      "GET",
				Path:        "/events",
				Description: "SSE 实时事件流",
				Example:     `curl -N http://localhost:8080/api/v1/events`,
			},
			{
				Method:      "POST",
				Path:        "/ai/chat",
				Description: "AI 对话",
				Example: `curl -X POST http://localhost:8080/api/v1/ai/chat \\
  -H "Content-Type: application/json" \\
  -d '{"message":"hello","use_tools":true}'`,
			},
			{
				Method:      "GET",
				Path:        "/ai/tools",
				Description: "工具列表（OpenAI 格式）",
			},
			{
				Method:      "POST",
				Path:        "/commands",
				Description: "发送 MC 命令",
				Example: `curl -X POST http://localhost:8080/api/v1/commands \\
  -H "Content-Type: application/json" \\
  -d '{"command":"say Hello"}'`,
			},
			{
				Method:      "POST",
				Path:        "/build",
				Description: "生成建筑",
				Example: `curl -X POST http://localhost:8080/api/v1/build \\
  -H "Content-Type: application/json" \\
  -d '{"type":"house","size":10,"x":0,"y":64,"z":0}'`,
			},
			{
				Method:      "GET",
				Path:        "/plugins",
				Description: "插件列表",
			},
			{
				Method:      "POST",
				Path:        "/plugins/register",
				Description: "注册插件",
			},
			{
				Method:      "POST",
				Path:        "/admin/reload",
				Description: "重载配置（管理员）",
			},
			{
				Method:      "POST",
				Path:        "/admin/restart",
				Description: "重启服务（管理员）",
			},
		},
	}
}
