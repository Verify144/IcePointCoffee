package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
)

// StdioTransport STDIO 传输层（MCP 标准传输）
// 读取 stdin（JSON-RPC 请求）→ 写入 stdout（JSON-RPC 响应）
type StdioTransport struct {
	server  *Server
	stdin  io.Reader
	stdout io.Writer
	stderr io.Writer
	stopCh chan struct{}
	wg     sync.WaitGroup
}

// NewStdioTransport 创建 STDIO 传输
func NewStdioTransport(server *Server) *StdioTransport {
	return &StdioTransport{
		server:  server,
		stdin:  os.Stdin,
		stdout: os.Stdout,
		stderr: os.Stderr,
		stopCh: make(chan struct{}),
	}
}

// Run 启动 STDIO 服务器
// 协议：每个 JSON-RPC 请求以 '\n' 分隔，响应也以 '\n' 分隔
func (t *StdioTransport) Run(ctx context.Context) error {
	t.wg.Add(1)
	defer t.wg.Done()

	// 写入版本头到 stderr（调试用，Claude Desktop 不会读取 stderr）
	fmt.Fprintf(t.stderr, "[IcePointCoffee MCP] Server starting (protocol %s)\n", ProtocolVersion)

	// 使用 bufio.Scanner 读取每行 JSON-RPC
	scanner := bufio.NewScanner(t.stdin)
	// 默认 64KB，JSON-RPC 请求可能较大
	scanner.Buffer(make([]byte, 0), 1024*1024)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue // 跳过空行
		}

		// 检查 stop 信号
		select {
		case <-t.stopCh:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// 处理请求
		response := t.server.Handle(ctx, []byte(line))
		if response == nil {
			// Notification，不需要响应
			continue
		}

		// 写入响应（JSON + newline）
		if _, err := fmt.Fprintf(t.stdout, "%s\n", response); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// broken pipe 通常是客户端关闭
			return nil
		}
		// 尝试 Sync（仅对 *os.File 有效）
		if f, ok := t.stdout.(*os.File); ok {
			f.Sync()
		}
	}

	if err := scanner.Err(); err != nil {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		return fmt.Errorf("stdin scanner error: %w", err)
	}
	return nil
}

// Stop 停止服务器
func (t *StdioTransport) Stop() {
	close(t.stopCh)
}

// WriteNotification 发送通知（无响应要求）
func (t *StdioTransport) WriteNotification(method string, params interface{}) error {
	notif := struct {
		JSONRPC string `json:"jsonrpc"`
		Method  string `json:"method"`
		Params  interface{} `json:"params,omitempty"`
	}{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
	}
	data, err := json.Marshal(notif)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(t.stdout, "%s\n", data)
	return err
}

// UseStdinStdout 测试用：自定义 io
func (t *StdioTransport) UseStdinStdout(stdin io.Reader, stdout, stderr io.Writer) {
	t.stdin = stdin
	t.stdout = stdout
	t.stderr = stderr
}
