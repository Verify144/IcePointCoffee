package ai

import (
	"context"
	"encoding/json"
	"testing"
)

func TestToolRegistry(t *testing.T) {
	r := NewToolRegistry()
	r.Register(&EchoTool{})
	r.Register(&GetTimeTool{})

	if r.List()[0].Name() == "" {
		t.Error("Should have tools")
	}

	// 调用工具
	args := json.RawMessage(`{"text": "hello"}`)
	result := r.CallTool(context.Background(), "echo", args)
	if result.Error != "" {
		t.Errorf("Echo should succeed: %s", result.Error)
	}
	if result.Result != "hello" {
		t.Errorf("Expected 'hello', got %v", result.Result)
	}
}

func TestMemory(t *testing.T) {
	m := NewMemory(10)
	m.Add(Message{Role: "user", Content: "hi"})
	m.Add(Message{Role: "assistant", Content: "hello"})
	m.Add(Message{Role: "user", Content: "how are you?"})

	if m.Size() != 3 {
		t.Errorf("Expected 3 messages, got %d", m.Size())
	}

	last, ok := m.LastAssistant()
	if !ok || last.Content != "hello" {
		t.Errorf("Last assistant should be 'hello', got %v", last)
	}

	// 测试限制
	for i := 0; i < 20; i++ {
		m.Add(Message{Role: "user", Content: "x"})
	}
	if m.Size() > 10 {
		t.Errorf("Memory size should be capped, got %d", m.Size())
	}
}

func TestCalculate(t *testing.T) {
	tests := []struct {
		expr     string
		expected float64
	}{
		{"1+1", 2},
		{"5-2", 3},
		{"2*3", 6},
		{"10/2", 5},
		{"1+2*3", 7},
		{"10-2*3", 4},
	}
	for _, tt := range tests {
		got, err := simpleCalc(tt.expr)
		if err != nil {
			t.Errorf("simpleCalc(%s) error: %v", tt.expr, err)
			continue
		}
		if got != tt.expected {
			t.Errorf("simpleCalc(%s) = %f, want %f", tt.expr, got, tt.expected)
		}
	}
}

func TestToolToOpenAI(t *testing.T) {
	r := NewToolRegistry()
	r.Register(&EchoTool{})

	tools := r.ToOpenAI()
	if len(tools) != 1 {
		t.Fatalf("Expected 1 tool, got %d", len(tools))
	}

	if tools[0].Type != "function" {
		t.Errorf("Type should be 'function'")
	}
	if tools[0].Function.Name != "echo" {
		t.Errorf("Name should be 'echo'")
	}
	if tools[0].Function.Description == "" {
		t.Error("Description should not be empty")
	}
}

func TestCallUnknownTool(t *testing.T) {
	r := NewToolRegistry()
	result := r.CallTool(context.Background(), "unknown", nil)
	if result.Error == "" {
		t.Error("Should return error for unknown tool")
	}
}
