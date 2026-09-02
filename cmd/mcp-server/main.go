// mcp-server 是 IcePointCoffee 的 MCP（Model Context Protocol）服务器
//
// 用法：
//   icepoint mcp-server
//
// Claude Desktop 配置（~/.config/Claude/claude_desktop_config.json）：
//   {
//     "mcpServers": {
//       "IcePointCoffee": {
//         "command": "/path/to/icepoint",
//         "args": ["mcp-server"],
//         "env": {
//           "VERIFY_SERVER": "hk-8.qwq.fan:19132",
//           "VERIFY_TOKEN": "your-fb-token"
//         }
//       }
//     }
//   }
//
// Cursor / 其他 AI 客户端配置：HTTP 模式见 server 端 /mcp/info
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Verify144/IcePointCoffee/internal/ai"
	"github.com/Verify144/IcePointCoffee/internal/mcp"
	"github.com/Verify144/IcePointCoffee/internal/mc"
)

func main() {
	// 启动日志输出到 stderr（stdout 留给 MCP 协议）
	fmt.Fprintln(os.Stderr, "[IcePointCoffee MCP] Starting MCP server v0.1.0")

	// 创建工具注册表（不连接 MC 客户端，等待 AI 框架触发）
	registry := ai.NewToolRegistry()
	mcCtrl := ai.RegisterMCTools(registry)

	// 尝试注入 MC 客户端（如果有 MC 配置）
	if err := tryConnectMC(mcCtrl); err != nil {
		fmt.Fprintf(os.Stderr, "[IcePointCoffee MCP] MC not connected: %v\n", err)
		fmt.Fprintln(os.Stderr, "[IcePointCoffee MCP] Running with mock tools (use env vars to connect real MC server)")
		mcCtrl.Inject(mc.NewMock(true))
	}

	// 创建 MCP 服务器
	server := mcp.NewServer("IcePointCoffee", "0.1.0")
	server.SetToolSource(func() []mcp.ToolExecutor {
		tools := registry.List()
		out := make([]mcp.ToolExecutor, 0, len(tools))
		for _, t := range tools {
			out = append(out, &aiToolAdapter{tool: t})
		}
		return out
	})

	// 创建 STDIO 传输
	transport := mcp.NewStdioTransport(server)

	// 处理 Ctrl+C / SIGTERM
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigCh
		fmt.Fprintln(os.Stderr, "[IcePointCoffee MCP] Shutting down")
		cancel()
		transport.Stop()
	}()

	// 运行
	if err := transport.Run(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "[IcePointCoffee MCP] Error: %v\n", err)
		os.Exit(1)
	}
}

// tryConnectMC 尝试连接 MC 服务器（从环境变量）
func tryConnectMC(ctrl *ai.MCController) error {
	// 当前：仅打印提示，MC 客户端连接需要 MC 配置
	// TODO: 后续集成 MC Config → netherite.Client → mc.Adapter
	addr := os.Getenv("VERIFY_SERVER")
	token := os.Getenv("VERIFY_TOKEN")
	cookie := os.Getenv("VERIFY_COOKIE")
	if addr == "" || token == "" {
		return fmt.Errorf("VERIFY_SERVER and VERIFY_TOKEN not set")
	}
	_ = cookie
	return fmt.Errorf("MC client creation not yet implemented in mcp-server (use HTTP mode)")
}

// aiToolAdapter 把 ai.Tool 转成 mcp.ToolExecutor
type aiToolAdapter struct {
	tool ai.Tool
}

func (a *aiToolAdapter) Name() string        { return a.tool.Name() }
func (a *aiToolAdapter) Description() string { return a.tool.Description() }
func (a *aiToolAdapter) Parameters() map[string]interface{} {
	return a.tool.Parameters()
}
func (a *aiToolAdapter) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	return a.tool.Execute(ctx, args)
}
