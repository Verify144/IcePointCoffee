package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client AI 客户端
type Client struct {
	apiURL  string
	apiKey  string
	model   string
	timeout time.Duration
	http    *http.Client
}

// Config AI 配置
type Config struct {
	APIURL  string
	APIKey  string
	Model   string
	Timeout time.Duration
}

// NewClient 创建 AI 客户端
func NewClient(cfg Config) *Client {
	if cfg.Timeout <= 0 {
		cfg.Timeout = 30 * time.Second
	}
	if cfg.Model == "" {
		cfg.Model = "gpt-3.5-turbo"
	}
	return &Client{
		apiURL:  cfg.APIURL,
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		timeout: cfg.Timeout,
		http:    &http.Client{Timeout: cfg.Timeout},
	}
}

// ChatRequest 聊天请求
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []OpenAITool `json:"tools,omitempty"`
	Stream   bool      `json:"stream"`
}

// ChatResponse 聊天响应
type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
}

// Choice 选择
type Choice struct {
	Message      ChoiceMessage `json:"message"`
	FinishReason string        `json:"finish_reason"`
}

// ChoiceMessage 选择消息
type ChoiceMessage struct {
	Role    string      `json:"role"`
	Content string      `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall 工具调用
type ToolCall struct {
	ID       string       `json:"id"`
	Type     string       `json:"type"`
	Function FunctionCall `json:"function"`
}

// FunctionCall 函数调用
type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// Chat 单轮对话
func (c *Client) Chat(ctx context.Context, messages []Message) (string, error) {
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}

// ChatWithTools 带工具的对话
func (c *Client) ChatWithTools(ctx context.Context, messages []Message, tools []Tool) (string, error) {
	// 转换 tools 为 OpenAI 格式
	openAITools := convertTools(tools)
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    openAITools,
	}
	resp, err := c.do(ctx, req)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response")
	}
	return resp.Choices[0].Message.Content, nil
}

func (c *Client) do(ctx context.Context, req ChatRequest) (*ChatResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI API error: %d %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, err
	}
	return &chatResp, nil
}

func convertTools(tools []Tool) []OpenAITool {
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
