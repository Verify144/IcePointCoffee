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
支持的指令格式：
- setblock <x> <y> <z> <block_name>
- fill <x1> <y1> <z1> <x2> <y2> <z2> <block_name>
请以 JSON 格式回复：
{
  "type": "command|structure|import",
  "description": "对用户需求的描述",
  "commands": ["指令1", "指令2", ...]
}`,
	}
}

// Task 任务
type Task = db.Task

// Execute 执行任务
func (e *Engine) Execute(ctx context.Context, prompt string) (*Task, error) {
	task := &Task{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Prompt:    prompt,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	if err := e.taskStore.Create(task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	messages := []ai.Message{
		{Role: "system", Content: e.systemPrompt},
		{Role: "user", Content: prompt},
	}

	var content string
	var err error

	if e.aiClient != nil {
		content, err = e.aiClient.Chat(ctx, messages)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			e.taskStore.Update(task)
			return task, err
		}
	} else {
		// Mock AI
		content = mockAIResponse(prompt)
	}

	task.Result = content

	result, err := e.parseResult(content)
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
		e.taskStore.Update(task)
		return task, err
	}

	task.Status = "completed"
	task.Result = result
	e.taskStore.Update(task)
	return task, nil
}

// ExecuteWithTools 带工具的任务执行
func (e *Engine) ExecuteWithTools(ctx context.Context, prompt string, tools []ai.Tool) (*Task, error) {
	task := &Task{
		ID:        fmt.Sprintf("task_%d", time.Now().UnixNano()),
		Prompt:    prompt,
		Status:    "running",
		CreatedAt: time.Now(),
	}

	if err := e.taskStore.Create(task); err != nil {
		return nil, fmt.Errorf("创建任务失败: %w", err)
	}

	messages := []ai.Message{
		{Role: "system", Content: e.systemPrompt},
		{Role: "user", Content: prompt},
	}

	var content string
	var err error

	if e.aiClient != nil {
		content, err = e.aiClient.ChatWithTools(ctx, messages, tools)
		if err != nil {
			task.Status = "failed"
			task.Error = err.Error()
			e.taskStore.Update(task)
			return task, err
		}
	} else {
		content = mockAIResponse(prompt)
	}

	result, err := e.parseResult(content)
	if err != nil {
		task.Status = "failed"
		task.Error = err.Error()
	} else {
		task.Status = "completed"
	}

	task.Result = result
	e.taskStore.Update(task)
	return task, nil
}

// parseResult 解析 AI 结果
func (e *Engine) parseResult(content string) (string, error) {
	content = strings.TrimSpace(content)

	// 尝试解析 JSON
	var data struct {
		Type        string   `json:"type"`
		Description string   `json:"description"`
		Commands    []string `json:"commands"`
	}
	if err := json.Unmarshal([]byte(content), &data); err != nil {
		return content, nil // 返回原文
	}

	if data.Description != "" {
		return data.Description, nil
	}
	if len(data.Commands) > 0 {
		return fmt.Sprintf("执行 %d 条命令", len(data.Commands)), nil
	}
	return content, nil
}

// mockAIResponse Mock AI 回复
func mockAIResponse(prompt string) string {
	prompt = strings.ToLower(prompt)
	if strings.Contains(prompt, "house") || strings.Contains(prompt, "房子") {
		return `{"type":"structure","description":"建造一个 10x10 的橡木房子","commands":["fill 0 64 0 9 69 9 stone","fill 1 65 1 8 68 8 air"]}`
	}
	if strings.Contains(prompt, "tower") || strings.Contains(prompt, "塔") {
		return `{"type":"structure","description":"建造一个 5x5 的石砖塔","commands":["fill 0 64 0 4 84 4 stone_bricks"]}`
	}
	if strings.Contains(prompt, "circle") || strings.Contains(prompt, "圆") {
		return `{"type":"structure","description":"建造一个半径 5 的石头圆形平台","commands":["fill -5 64 -5 5 64 5 stone"]}`
	}
	return `{"type":"command","description":"收到请求","commands":[]}`
}

// GetTask 获取任务
func (e *Engine) GetTask(id string) *Task {
	t, _ := e.taskStore.GetByID(id)
	return t
}

// ListTasks 列出任务
func (e *Engine) ListTasks() []*Task {
	tasks, _ := e.taskStore.ListAll(50)
	return tasks
}
