// Package agent 提供 AI Agent 引擎。
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/ai"
	"github.com/Verify144/IcePointCoffee/internal/db"
)

// Engine Agent 引擎。
type Engine struct {
	aiClient     *ai.Client
	taskStore    *db.TaskStore
	systemPrompt string
}

// NewEngine 创建 Agent 引擎。
func NewEngine(aiClient *ai.Client, taskStore *db.TaskStore) *Engine {
	return &Engine{
		aiClient:  aiClient,
		taskStore: taskStore,
		systemPrompt: `你是一个 Minecraft 建筑助手。用户会提出建筑需求（如"建一个 10x10 的房子"、"做一个圆形喷泉"等）。
你需要将需求转化为具体的 Minecraft 指令或建筑数据。
支持的指令格式：
- setblock <x> <y> <z> <block_name>
- fill <x1> <y1> <z1> <x2> <y2> <z2> <block_name>
- structure load <name> <x> <y> <z>

请以 JSON 格式回复，包含以下字段：
{
  "type": "command|structure|import",
  "description": "对用户需求的描述",
  "commands": ["指令1", "指令2", ...],
  "structure_data": { /* 可选，结构数据 */ }
}

如果用户想要的是建筑导入，请设置 type 为 "import"，并提供建筑文件路径或结构数据。
如果需要多步指令，请拆分为多个单独命令。
只回复 JSON，不要有其他内容。`,
	}
}

// Task 任务。直接使用 db.Task。
type Task = db.Task

// Handle 处理用户需求。
func (e *Engine) Handle(ctx context.Context, userID, prompt string) (*Task, error) {
	task := &db.Task{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		UserID:    userID,
		Prompt:    prompt,
		Type:      "command",
		Status:    "running",
		CreatedAt: time.Now(),
	}

	if err := e.taskStore.Create(task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	messages := []ai.ChatMessage{
		{Role: "system", Content: e.systemPrompt},
		{Role: "user", Content: prompt},
	}

	resp, err := e.aiClient.Chat(ctx, messages)
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		e.taskStore.Update(task)
		return task, err
	}

	if len(resp.Choices) == 0 {
		task.Status = "failed"
		task.Error = "AI 未返回有效内容"
		e.taskStore.Update(task)
		return task, fmt.Errorf("AI 未返回有效内容")
	}

	content := resp.Choices[0].Message.Content
	task.Result = content

	result, err := e.parseResult(content)
	if err != nil {
		task.Status = "failed"
		task.Error = fmt.Sprintf("解析 AI 返回失败: %v | 原始内容: %s", err, content)
		e.taskStore.Update(task)
		return task, err
	}

	task.Type = result.Type
	task.Description = result.Description
	task.Commands = result.Commands
	task.Status = "done"
	task.DoneAt = time.Now()
	e.taskStore.Update(task)

	return task, nil
}

// HandleStream 流式处理用户需求。
func (e *Engine) HandleStream(ctx context.Context, userID, prompt string, handler func(content string)) (*Task, error) {
	task := &db.Task{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		UserID:    userID,
		Prompt:    prompt,
		Type:      "command",
		Status:    "running",
		CreatedAt: time.Now(),
	}

	if err := e.taskStore.Create(task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	messages := []ai.ChatMessage{
		{Role: "system", Content: e.systemPrompt},
		{Role: "user", Content: prompt},
	}

	var fullContent strings.Builder
	err := e.aiClient.ChatStream(ctx, messages, func(content string) {
		fullContent.WriteString(content)
		handler(content)
	})
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		e.taskStore.Update(task)
		return task, err
	}

	task.Result = fullContent.String()

	result, err := e.parseResult(task.Result)
	if err != nil {
		task.Status = "failed"
		task.Error = fmt.Sprintf("解析 AI 返回失败: %v", err)
		e.taskStore.Update(task)
		return task, err
	}

	task.Type = result.Type
	task.Description = result.Description
	task.Commands = result.Commands
	task.Status = "done"
	task.DoneAt = time.Now()
	e.taskStore.Update(task)

	return task, nil
}

// parseResult 解析 AI 返回的 JSON 结果。
func (e *Engine) parseResult(content string) (*ParseResult, error) {
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start < 0 || end < 0 || end <= start {
		return nil, fmt.Errorf("无法从内容中提取 JSON")
	}
	jsonStr := content[start : end+1]

	var result ParseResult
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("JSON 解析失败: %w", err)
	}

	return &result, nil
}

// ParseResult AI 返回的解析结果。
type ParseResult struct {
	Type          string   `json:"type"`
	Description   string   `json:"description"`
	Commands      []string `json:"commands,omitempty"`
	StructureData any      `json:"structure_data,omitempty"`
}