package ai

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestMockStream(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "hello"},
	}

	var chunks []StreamChunk
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := MockStream(ctx, messages, func(c StreamChunk) {
		chunks = append(chunks, c)
	})

	if err != nil {
		t.Errorf("MockStream should succeed: %v", err)
	}

	// 应该有内容片段 + done
	if len(chunks) < 2 {
		t.Errorf("Should have multiple chunks, got %d", len(chunks))
	}

	// 收集所有内容
	var content strings.Builder
	for _, c := range chunks {
		if c.Type == "content" {
			content.WriteString(c.Content)
		}
	}

	if !strings.Contains(content.String(), "你好") {
		t.Errorf("Content should contain greeting, got: %s", content.String())
	}

	// 最后一个应该是 done
	if chunks[len(chunks)-1].Type != "done" {
		t.Error("Last chunk should be done")
	}
}

func TestMockStreamContextCancel(t *testing.T) {
	messages := []Message{{Role: "user", Content: "hello world hello world hello world"}}
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	chunks := 0
	_ = MockStream(ctx, messages, func(c StreamChunk) {
		chunks++
	})

	// 应该被中断，少于无限制时
	if chunks == 0 {
		t.Error("Should have received some chunks")
	}
}

func TestStreamChunkStructure(t *testing.T) {
	c := StreamChunk{Type: "content", Content: "hello"}
	if c.Type != "content" {
		t.Error("Type mismatch")
	}
	if c.Content != "hello" {
		t.Error("Content mismatch")
	}

	c2 := StreamChunk{Type: "tool_call", ToolName: "echo", ToolArgs: `{"text":"x"}`}
	if c2.ToolName != "echo" {
		t.Error("ToolName mismatch")
	}

	c3 := StreamChunk{Type: "done", Done: true}
	if !c3.Done {
		t.Error("Done flag should be true")
	}
}

func TestParseSSE(t *testing.T) {
	input := `data: {"choices":[{"delta":{"content":"Hello"}}]}

data: {"choices":[{"delta":{"content":" World"}}]}

data: {"choices":[{"delta":{"content":"!"}}],"finish_reason":"stop"}

data: [DONE]
`
	var chunks []StreamChunk
	err := parseSSE(strings.NewReader(input), func(c StreamChunk) {
		chunks = append(chunks, c)
	})

	if err != nil {
		t.Errorf("parseSSE should succeed: %v", err)
	}

	var content strings.Builder
	for _, c := range chunks {
		content.WriteString(c.Content)
	}

	if content.String() != "Hello World!" {
		t.Errorf("Expected 'Hello World!', got '%s'", content.String())
	}
}

func TestParseSSEWithToolCall(t *testing.T) {
	input := `data: {"choices":[{"delta":{"tool_calls":[{"index":0,"id":"call_1","type":"function","function":{"name":"echo","arguments":""}}]}}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":"{\"text\""}}]}}]}

data: {"choices":[{"delta":{"tool_calls":[{"index":0,"function":{"arguments":":\"hi\"}"}}]}}]}

data: [DONE]
`
	var chunks []StreamChunk
	err := parseSSE(strings.NewReader(input), func(c StreamChunk) {
		chunks = append(chunks, c)
	})

	if err != nil {
		t.Errorf("parseSSE should succeed: %v", err)
	}

	toolChunks := 0
	for _, c := range chunks {
		if c.Type == "tool_call" && c.ToolName == "echo" {
			toolChunks++
		}
	}

	if toolChunks == 0 {
		t.Error("Should have tool_call chunks")
	}
}
