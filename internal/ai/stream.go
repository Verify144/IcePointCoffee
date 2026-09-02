package ai

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// StreamChunk 流式响应块
type StreamChunk struct {
	Type     string `json:"type"` // content / tool_call / tool_result / done / error / session / start
	Content  string `json:"content,omitempty"`
	ToolName string `json:"tool_name,omitempty"`
	ToolArgs string `json:"tool_args,omitempty"`
	ToolID   string `json:"tool_id,omitempty"`
	ToolResult string `json:"tool_result,omitempty"` // 工具执行结果（前端回传后回显）
	ToolSuccess bool   `json:"tool_success,omitempty"` // 工具是否执行成功
	ToolError string `json:"tool_error,omitempty"`   // 工具错误信息
	Round     int    `json:"round,omitempty"`        // 第几轮工具调用
	Error     string `json:"error,omitempty"`
	Done      bool   `json:"done,omitempty"`
}

// Stream 流式聊天
func (c *Client) Stream(ctx context.Context, messages []Message, tools []Tool, callback func(StreamChunk)) error {
	openAITools := convertTools(tools)
	req := ChatRequest{
		Model:    c.model,
		Messages: messages,
		Tools:    openAITools,
		Stream:   true,
	}
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.apiURL+"/chat/completions", bytes.NewReader(data))
	if err != nil {
		return err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "text/event-stream")
	if c.apiKey != "" {
		httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("AI API error: %d %s", resp.StatusCode, string(body))
	}

	return parseSSE(resp.Body, callback)
}

// parseSSE 解析 SSE 流
func parseSSE(reader io.Reader, callback func(StreamChunk)) error {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			callback(StreamChunk{Type: "done", Done: true})
			return nil
		}

		var chunk struct {
			Choices []struct {
				Delta struct {
					Role    string `json:"role,omitempty"`
					Content string `json:"content,omitempty"`
					ToolCalls []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Type     string `json:"type"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls,omitempty"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) == 0 {
			continue
		}

		delta := chunk.Choices[0].Delta

		// 内容片段
		if delta.Content != "" {
			callback(StreamChunk{Type: "content", Content: delta.Content})
		}

		// 工具调用
		for _, tc := range delta.ToolCalls {
			if tc.Function.Name != "" || tc.Function.Arguments != "" {
				callback(StreamChunk{
					Type:     "tool_call",
					ToolID:   tc.ID,
					ToolName: tc.Function.Name,
					ToolArgs: tc.Function.Arguments,
				})
			}
		}

		// 完成
		if chunk.Choices[0].FinishReason != "" {
			callback(StreamChunk{Type: "done", Done: true})
			return nil
		}
	}

	if err := scanner.Err(); err != nil {
		callback(StreamChunk{Type: "error", Error: err.Error()})
		return err
	}
	return nil
}

// MockStream Mock 流式输出（用于无 API 测试）
func MockStream(ctx context.Context, messages []Message, callback func(StreamChunk)) error {
	last := messages[len(messages)-1]
	prompt := last.Content

	// Mock 工具演示：若消息包含关键词则触发工具气泡
	toolDemo := []struct {
		keyword string
		tool    string
		args    string
		result  string
	}{
		{"fill", "mc_fill", `{"x1":0,"y1":64,"z1":0,"x2":10,"y2":70,"z2":10,"block":"stone"}`, "填充完成：110 个方块（10×7×10）"},
		{"give", "mc_give", `{"target":"@s","item":"diamond","count":64}`, "已给予 Steve 64 钻石"},
		{"spawn", "mc_spawn_entity", `{"entity_type":"zombie","x":0,"y":64,"z":0}`, "已召唤 zombie 于 (0, 64, 0)"},
		{"sound", "mc_play_sound", `{"sound_id":"random.levelup","x":0,"y":64,"z":0}`, "已播放 random.levelup"},
		{"particle", "mc_particle", `{"particle_id":12,"x":0,"y":64,"z":0}`, "已发射心形粒子效果"},
	}
	lower := strings.ToLower(prompt)

	for i, demo := range toolDemo {
		if strings.Contains(lower, demo.keyword) {
			callback(StreamChunk{Type: "tool_call", ToolName: demo.tool, ToolArgs: demo.args, ToolID: fmt.Sprintf("mock_%d", i), Round: 1})
			time.Sleep(600 * time.Millisecond)
			callback(StreamChunk{
				Type:        "tool_result",
				ToolName:    demo.tool,
				ToolID:      fmt.Sprintf("mock_%d", i),
				ToolResult:  demo.result,
				ToolSuccess: true,
				Round:       1,
			})
			time.Sleep(300 * time.Millisecond)
		}
	}

	var response string
	switch {
	case strings.Contains(lower, "hello") || strings.Contains(lower, "hi"):
		response = "你好！我是 IcePoint Coffee AI 助手。有什么可以帮你的吗？"
	case strings.Contains(lower, "house") || strings.Contains(lower, "房子"):
		response = "好的，我来帮你建造一个房子。我会使用橡木和玻璃。先铺地基，然后搭建四面墙，最后封顶。"
	case strings.Contains(lower, "time") || strings.Contains(lower, "时间"):
		response = fmt.Sprintf("现在时间是 %s", time.Now().Format("2006-01-02 15:04:05"))
	case strings.Contains(lower, "status"):
		response = "系统运行正常！所有组件都在工作。"
	case strings.Contains(lower, "fill") || strings.Contains(lower, "give") || strings.Contains(lower, "spawn"):
		response = "我已经执行了相应的操作。以上是执行结果的实时反馈。"
	default:
		response = "我已收到你的请求：" + prompt + "\n\n当前是 Mock 模式，配置 AI API 可获得真实回复。"
	}

	// 逐字发送（模拟流式）
	for _, ch := range response {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		callback(StreamChunk{Type: "content", Content: string(ch)})
		time.Sleep(20 * time.Millisecond)
	}
	callback(StreamChunk{Type: "done", Done: true})
	return nil
}
