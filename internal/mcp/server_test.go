package mcp

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

// mockTool 模拟 AI 工具
type mockTool struct {
	name   string
	desc   string
	params map[string]interface{}
	output interface{}
}

func (t *mockTool) Name() string                       { return t.name }
func (t *mockTool) Description() string                { return t.desc }
func (t *mockTool) Parameters() map[string]interface{} { return t.params }
func (t *mockTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	return t.output, nil
}

func TestServerInitialize(t *testing.T) {
	s := NewServer("TestServer", "1.0.0")
	s.SetToolSource(func() []ToolExecutor { return nil })

	req := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"test"}}}`
	resp := s.Handle(context.Background(), []byte(req))

	if !strings.Contains(string(resp), `"protocolVersion":"2025-06-18"`) {
		t.Errorf("Missing protocol version: %s", resp)
	}
	if !strings.Contains(string(resp), `"name":"TestServer"`) {
		t.Errorf("Missing server name: %s", resp)
	}
	if !strings.Contains(string(resp), `"tools"`) {
		t.Errorf("Missing tools capability: %s", resp)
	}
}

func TestServerToolsList(t *testing.T) {
	s := NewServer("TestServer", "1.0.0")
	s.SetToolSource(func() []ToolExecutor {
		return []ToolExecutor{
			&mockTool{
				name:   "test_tool",
				desc:   "A test tool",
				params: map[string]interface{}{"type": "object"},
				output: "ok",
			},
		}
	})

	req := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	resp := s.Handle(context.Background(), []byte(req))

	if !strings.Contains(string(resp), `"test_tool"`) {
		t.Errorf("Missing test_tool: %s", resp)
	}
	if !strings.Contains(string(resp), `"inputSchema"`) {
		t.Errorf("Missing inputSchema: %s", resp)
	}
}

func TestServerToolsCall(t *testing.T) {
	s := NewServer("TestServer", "1.0.0")
	s.SetToolSource(func() []ToolExecutor {
		return []ToolExecutor{
			&mockTool{
				name:   "echo",
				desc:   "Echo",
				params: map[string]interface{}{"type": "object"},
				output: map[string]string{"echo": "hello"},
			},
		}
	})

	req := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"echo","arguments":{"text":"hello"}}}`
	resp := s.Handle(context.Background(), []byte(req))

	if !strings.Contains(string(resp), `"content"`) {
		t.Errorf("Missing content: %s", resp)
	}
	if !strings.Contains(string(resp), `echo`) {
		t.Errorf("Missing echo in response: %s", resp)
	}
}

func TestServerToolsCallUnknown(t *testing.T) {
	s := NewServer("TestServer", "1.0.0")
	s.SetToolSource(func() []ToolExecutor { return nil })

	req := `{"jsonrpc":"2.0","id":4,"method":"tools/call","params":{"name":"nonexistent","arguments":{}}}`
	resp := s.Handle(context.Background(), []byte(req))

	if !strings.Contains(string(resp), `"error"`) {
		t.Errorf("Expected error: %s", resp)
	}
	if !strings.Contains(string(resp), `"code":-32602`) {
		t.Errorf("Expected -32602: %s", resp)
	}
}

func TestServerNotification(t *testing.T) {
	s := NewServer("TestServer", "1.0.0")
	// Notification 没有 ID，不应该返回响应
	req := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	resp := s.Handle(context.Background(), []byte(req))

	if resp != nil {
		t.Errorf("Notification should return nil, got: %s", resp)
	}
	if !s.initialized {
		t.Error("Server should be initialized")
	}
}

func TestServerMethodNotFound(t *testing.T) {
	s := NewServer("TestServer", "1.0.0")

	req := `{"jsonrpc":"2.0","id":5,"method":"unknown/method"}`
	resp := s.Handle(context.Background(), []byte(req))

	if !strings.Contains(string(resp), `"code":-32601`) {
		t.Errorf("Expected method not found error: %s", resp)
	}
}

func TestServerPing(t *testing.T) {
	s := NewServer("TestServer", "1.0.0")
	req := `{"jsonrpc":"2.0","id":6,"method":"ping"}`
	resp := s.Handle(context.Background(), []byte(req))

	if !strings.Contains(string(resp), `"result"`) {
		t.Errorf("Expected ping result: %s", resp)
	}
}

func TestServerParseError(t *testing.T) {
	s := NewServer("TestServer", "1.0.0")
	resp := s.Handle(context.Background(), []byte(`{not json`))

	if !strings.Contains(string(resp), `"code":-32700`) {
		t.Errorf("Expected parse error: %s", resp)
	}
}

func TestServerResourcesList(t *testing.T) {
	s := NewServer("TestServer", "1.0.0")
	req := `{"jsonrpc":"2.0","id":7,"method":"resources/list"}`
	resp := s.Handle(context.Background(), []byte(req))

	if !strings.Contains(string(resp), `"icepoint://server/info"`) {
		t.Errorf("Missing server info resource: %s", resp)
	}
}
