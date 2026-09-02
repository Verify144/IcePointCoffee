package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestStdioTransportRoundTrip(t *testing.T) {
	srv := NewServer("Test", "1.0")
	srv.SetToolSource(func() []ToolExecutor {
		return []ToolExecutor{
			&mockTool{
				name:   "greet",
				desc:   "Greet someone",
				params: map[string]interface{}{"type": "object"},
				output: map[string]string{"greeting": "Hello, world!"},
			},
		}
	})

	// 内存 stdin/stdout
	initReq := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18","clientInfo":{"name":"TestClient"}}}`
	listReq := `{"jsonrpc":"2.0","id":2,"method":"tools/list"}`
	callReq := `{"jsonrpc":"2.0","id":3,"method":"tools/call","params":{"name":"greet","arguments":{"name":"Alice"}}}`
	content := initReq + "\n" + listReq + "\n" + callReq + "\n"

	stdinR := bytes.NewReader([]byte(content))
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	transport := NewStdioTransport(srv)
	transport.UseStdinStdout(stdinR, stdout, stderr)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = transport.Run(ctx)

	output := stdout.String()
	if !strings.Contains(output, `"protocolVersion":"2025-06-18"`) {
		t.Errorf("Missing protocol version:\n%s", output)
	}
	if !strings.Contains(output, `"greet"`) {
		t.Errorf("Missing greet tool:\n%s", output)
	}
	if !strings.Contains(output, `greeting`) {
		t.Errorf("Missing greeting:\n%s", output)
	}
}

func TestStdioTransportError(t *testing.T) {
	srv := NewServer("Test", "1.0")
	srv.SetToolSource(func() []ToolExecutor { return nil })

	badReq := `{"jsonrpc":"2.0","id":1,"method":"tools/call","params":{"name":"nonexistent","arguments":{}}}`
	stdinR := bytes.NewReader([]byte(badReq + "\n"))
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	transport := NewStdioTransport(srv)
	transport.UseStdinStdout(stdinR, stdout, stderr)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = transport.Run(ctx)

	output := stdout.String()
	if !strings.Contains(output, `"error"`) {
		t.Errorf("Expected error:\n%s", output)
	}
}

func TestStdioJSONFormat(t *testing.T) {
	srv := NewServer("Test", "1.0")
	srv.SetToolSource(func() []ToolExecutor { return nil })

	req := `{"jsonrpc":"2.0","id":42,"method":"ping"}`
	stdinR := bytes.NewReader([]byte(req + "\n"))
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	transport := NewStdioTransport(srv)
	transport.UseStdinStdout(stdinR, stdout, stderr)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = transport.Run(ctx)

	line := strings.TrimSpace(strings.Split(stdout.String(), "\n")[0])
	var resp Response
	if err := json.Unmarshal([]byte(line), &resp); err != nil {
		t.Fatalf("Response is not valid JSON: %s", err)
	}
	var id int
	if err := json.Unmarshal(resp.ID, &id); err != nil {
		t.Fatalf("ID is not int: %s", err)
	}
	if id != 42 {
		t.Errorf("Expected ID 42, got %d", id)
	}
}

func TestStdioNotification(t *testing.T) {
	srv := NewServer("Test", "1.0")
	srv.SetToolSource(func() []ToolExecutor { return nil })

	// 只发 notification，不发请求
	notif := `{"jsonrpc":"2.0","method":"notifications/initialized"}`
	stdinR := bytes.NewReader([]byte(notif + "\n"))
	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}

	transport := NewStdioTransport(srv)
	transport.UseStdinStdout(stdinR, stdout, stderr)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = transport.Run(ctx)

	// stdout 应该是空的（无响应要求）
	if stdout.Len() != 0 {
		t.Errorf("stdout should be empty for notification, got: %s", stdout.String())
	}
	if !srv.initialized {
		t.Error("Server should be initialized after notification")
	}
}
