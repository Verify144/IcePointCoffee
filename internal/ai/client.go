// Package ai 提供 OpenAI 兼容协议的客户端。
package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/config"
)

// Client OpenAI 兼容 AI 客户端。
type Client struct {
	baseURL    string
	apiKey     string
	model      string
	httpClient *http.Client
}

// NewClient 创建 AI 客户端。
func NewClient(cfg config.AIConfig) *Client {
	return &Client{
		baseURL: cfg.BaseURL,
		apiKey:  cfg.APIKey,
		model:   cfg.Model,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// ChatRequest 聊天请求。
type ChatRequest struct {
	Model       string         `json:"model"`
	Messages    []ChatMessage  `json:"messages"`
	Temperature float64        `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// ChatMessage 聊天消息。
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

// ChatResponse 聊天响应。
type ChatResponse struct {
	ID      string   `json:"id"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 选项。
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 用量。
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// Chat 发送聊天请求。
func (c *Client) Chat(ctx context.Context, messages []ChatMessage) (*ChatResponse, error) {
	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.7,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("AI 返回错误 %d: %s", resp.StatusCode, string(respBody))
	}

	var result ChatResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return &result, nil
}

// ChatStream 发送流式聊天请求。
func (c *Client) ChatStream(ctx context.Context, messages []ChatMessage, handler func(content string)) error {
	req := ChatRequest{
		Model:       c.model,
		Messages:    messages,
		Temperature: 0.7,
		Stream:      true,
	}

	body, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求失败: %w", err)
	}

	url := fmt.Sprintf("%s/chat/completions", c.baseURL)
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("创建请求失败: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI 返回错误 %d: %s", resp.StatusCode, string(respBody))
	}

	// 解析 SSE 流
	buf := make([]byte, 4096)
	reader := resp.Body
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			content := parseSSE(buf[:n])
			if content != "" {
				handler(content)
			}
		}
		if err != nil {
			break
		}
	}

	return nil
}

// parseSSE 从 SSE 数据中提取 content。
func parseSSE(data []byte) string {
	text := string(data)
	for {
		idx := -1
		for i := 0; i < len(text)-6; i++ {
			if text[i] == 'd' && text[i+1] == 'a' && text[i+2] == 't' && text[i+3] == 'a' && text[i+4] == ':' {
				idx = i + 5
				break
			}
		}
		if idx < 0 {
			return ""
		}
		text = text[idx:]
		// 解析 JSON
		var chunk struct {
			Choices []struct {
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
			} `json:"choices"`
		}
		if err := json.Unmarshal([]byte(text), &chunk); err != nil {
			return ""
		}
		if len(chunk.Choices) > 0 {
			return chunk.Choices[0].Delta.Content
		}
		return ""
	}
}
