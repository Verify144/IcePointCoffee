// Package ai 提供 AI Agent 与工具调用。
package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"
)

// Tool 工具接口
type Tool interface {
	Name() string
	Description() string
	Parameters() map[string]interface{}
	Execute(ctx context.Context, args json.RawMessage) (interface{}, error)
}

// ToolRegistry 工具注册表
type ToolRegistry struct {
	mu       sync.RWMutex
	tools    map[string]Tool
	observer func(name string, args json.RawMessage, result *CallResult, durationMs int64)
}

// NewToolRegistry 创建工具注册表
func NewToolRegistry() *ToolRegistry {
	return &ToolRegistry{tools: make(map[string]Tool)}
}

// Register 注册工具
func (r *ToolRegistry) Register(tool Tool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.tools[tool.Name()] = tool
}

// Get 获取工具
func (r *ToolRegistry) Get(name string) (Tool, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.tools[name]
	return t, ok
}

// List 列出所有工具
func (r *ToolRegistry) List() []Tool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	tools := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		tools = append(tools, t)
	}
	return tools
}

// CallResult 工具调用结果
type CallResult struct {
	ToolName string      `json:"tool_name"`
	Args     interface{} `json:"args"`
	Result   interface{} `json:"result"`
	Error    string      `json:"error,omitempty"`
	Time     time.Time   `json:"time"`
}

// CallTool 调用工具
func (r *ToolRegistry) CallTool(ctx context.Context, name string, args json.RawMessage) *CallResult {
	start := time.Now()
	tool, ok := r.Get(name)
	result := &CallResult{ToolName: name, Time: start}
	if !ok {
		result.Error = fmt.Sprintf("tool not found: %s", name)
		return result
	}

	var parsedArgs interface{}
	if err := json.Unmarshal(args, &parsedArgs); err != nil {
		result.Error = fmt.Sprintf("invalid args: %v", err)
		return result
	}
	result.Args = parsedArgs

	res, err := tool.Execute(ctx, args)
	durationMs := time.Since(start).Milliseconds()
	if err != nil {
		result.Error = err.Error()
	} else {
		result.Result = res
	}

	// 通知 observer（用于记录历史/SSE推送）
	if observer := r.getObserver(); observer != nil {
		observer(name, args, result, durationMs)
	}

	return result
}

// SetObserver 设置工具调用观察者（用于历史记录）
func (r *ToolRegistry) SetObserver(fn func(name string, args json.RawMessage, result *CallResult, durationMs int64)) {
	r.mu.Lock()
	r.observer = fn
	r.mu.Unlock()
}

func (r *ToolRegistry) getObserver() func(string, json.RawMessage, *CallResult, int64) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.observer
}

// OpenAITool OpenAI 兼容工具格式
type OpenAITool struct {
	Type     string                 `json:"type"`
	Function OpenAIToolFunction     `json:"function"`
}

// OpenAIToolFunction 工具函数描述
type OpenAIToolFunction struct {
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// ToOpenAI 转换为 OpenAI 格式
func (r *ToolRegistry) ToOpenAI() []OpenAITool {
	tools := r.List()
	result := make([]OpenAITool, 0, len(tools))
	for _, t := range tools {
		result = append(result, OpenAITool{
			Type: "function",
			Function: OpenAIToolFunction{
				Name:        t.Name(),
				Description: t.Description(),
				Parameters:  t.Parameters(),
			},
		})
	}
	return result
}
