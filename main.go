// IcePoint Coffee - Minecraft 建筑助手
// MIT License
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/agent"
	"github.com/Verify144/IcePointCoffee/internal/ai"
	"github.com/Verify144/IcePointCoffee/internal/builder"
	"github.com/Verify144/IcePointCoffee/internal/config"
	"github.com/Verify144/IcePointCoffee/internal/db"
	"github.com/Verify144/IcePointCoffee/internal/importer"
	"github.com/Verify144/IcePointCoffee/internal/mc"

	netherite_mc "github.com/Verify144/IcePointCoffee/internal/netherite/mc"
	"github.com/Verify144/IcePointCoffee/internal/netherite/protocol"
	"github.com/Verify144/IcePointCoffee/internal/plugin"
	"github.com/Verify144/IcePointCoffee/internal/server"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置错误: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n请创建配置文件 ~/.icepoint/config.yaml")
		os.Exit(1)
	}

	// 展开路径中的 ~
	dbPath := expandHome(cfg.DB.Path)
	if err := os.MkdirAll(dir(dbPath), 0755); err != nil {
		fmt.Fprintf(os.Stderr, "无法创建数据目录: %v\n", err)
		os.Exit(1)
	}

	// 初始化数据库
	database, err := db.New(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "数据库错误: %v\n", err)
		os.Exit(1)
	}
	defer database.Close()

	// 初始化组件
	taskStore := db.NewTaskStore(database)

	var aiClient *ai.Client
	if cfg.AI.BaseURL != "" && cfg.AI.APIKey != "" {
		aiClient = ai.NewClient(ai.Config{
			APIURL:  cfg.AI.BaseURL,
			APIKey:  cfg.AI.APIKey,
			Model:   cfg.AI.Model,
			Timeout: 30 * time.Second,
		})
	}

	agentEngine := agent.NewEngine(aiClient, taskStore)
	buildEngine := builder.New()
	pluginManager := plugin.NewManager(cfg.Plugin.Dir)

	// 初始化 MC 客户端
	mcClient, err := netherite_mc.NewClient(&netherite_mc.Options{
		FBToken:    cfg.Server.FBToken,
		ServerCode: cfg.Server.Address,
		ServerPass: "",
		PlayerName: cfg.Server.PlayerName,
		AuthServer: "",
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "MC 客户端初始化失败: %v\n", err)
		mcClient = nil
	} else {
		eventProcessor := netherite_mc.NewEventProcessor(mcClient)
		eventProcessor.Start()

		asyncExecutor := netherite_mc.NewAsyncExecutor(mcClient, 4)
		asyncExecutor.Start()
		defer asyncExecutor.Stop()

		eventProcessor.RegisterHandler(protocol.IDText, func(data []byte) error {
			text, _ := protocol.DecodeText(data)
			fmt.Printf("[聊天] %s\n", text.Message)
			return nil
		})
		eventProcessor.RegisterHandler(protocol.IDDisconnect, func(data []byte) error {
			fmt.Println("[MC] 服务器断开连接")
			return nil
		})
	}

	mcAdapter := mc.NewAdapter()
	mcAdapter.SetClient(mcClient)
	cmdExecutor := importer.NewExecutor(mcClient)
	_ = cmdExecutor

	// 启动插件
	plugins, _ := pluginManager.Scan()
	if len(plugins) > 0 {
		fmt.Printf("已扫描到 %d 个插件:\n", len(plugins))
		for _, p := range plugins {
			if err := pluginManager.Start(p.ID); err != nil {
				fmt.Fprintf(os.Stderr, "  启动 %s 失败: %v\n", p.Name, err)
			} else {
				fmt.Printf("  ✓ %s\n", p.Name)
			}
		}
	}

	// 启动 HTTP RPC 服务
	if cfg.Plugin.HTTPPort > 0 {
		// AI RPC 服务（支持 AI 工具调用）
		rpcServer := server.NewServerWithDB(int(cfg.Plugin.HTTPPort), database.SQLDB())
		mcAdapter := mc.NewAdapter()
		mcAdapter.SetClient(mcClient) // 注入真实客户端
		rpcServer.SetMCAdapter(mcAdapter)
		if err := rpcServer.Start(); err != nil {
			log.Fatalf("HTTP Server 启动失败: %v", err)
		}
		defer rpcServer.Stop()
		fmt.Printf("📊 Dashboard: http://127.0.0.1:%d/\n", cfg.Plugin.HTTPPort)
		fmt.Printf("🤖 AI Chat: POST http://127.0.0.1:%d/api/v1/ai/chat\n", cfg.Plugin.HTTPPort)
	}

	// 连接服务器
	ctx := context.Background()
	if mcClient != nil {
		fmt.Println("正在连接租赁服...")
		if err := mcClient.Connect(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "连接服务器失败: %v\n", err)
			fmt.Fprintln(os.Stderr, "提示: 请检查 server.address 和 server.fb_token 配置")
		} else {
			fmt.Printf("已连接到服务器: %s\n", mcClient.RemoteAddr())
		}
	}

	// 主界面
	fmt.Println()
	printBanner()
	fmt.Printf("冰点咖啡 v1.0.0\n")
	if cfg.AI.BaseURL != "" {
		fmt.Printf("模型: %s @ %s\n", cfg.AI.Model, cfg.AI.BaseURL)
	} else {
		fmt.Println("AI: Mock 模式")
	}
	if mcClient != nil {
		fmt.Printf("服务器: %s (%s)\n", mcClient.RemoteAddr(), cfg.Server.PlayerName)
	}
	if cfg.Plugin.HTTPPort > 0 {
		fmt.Printf("Dashboard: http://127.0.0.1:%d/\n", cfg.Plugin.HTTPPort)
	}
	fmt.Println()
	fmt.Println("输入你的建筑需求，例如：")
	fmt.Println("  build house width:10 height:5")
	fmt.Println("  build tower height:30")
	fmt.Println()
	fmt.Println("命令: /tasks  /plugins  /connect  /disconnect  /quit")

	for {
		fmt.Print("\n❄ > ")
		var input string
		if _, err := fmt.Fscan(os.Stdin, &input); err != nil {
			break
		}
		input = trimSpace(input)
		if input == "" {
			continue
		}

		switch {
		case input == "/quit", input == "/exit":
			fmt.Println("再见!")
			if mcClient != nil {
				mcClient.Close()
			}
			return
		case input == "/tasks":
			listTasks(taskStore)
			continue
		case input == "/plugins":
			listPlugins(pluginManager)
			continue
		case input == "/help":
			printHelp()
			continue
		case input == "/connect":
			if mcClient == nil {
				fmt.Println("MC 客户端未初始化")
				continue
			}
			if err := mcClient.Connect(ctx); err != nil {
				fmt.Fprintf(os.Stderr, "连接失败: %v\n", err)
			} else {
				fmt.Println("连接成功!")
			}
			continue
		case input == "/disconnect":
			if mcClient != nil {
				mcClient.Close()
				fmt.Println("已断开")
			}
			continue
		}

		// 处理建筑请求
		handleRequest(ctx, input, agentEngine, buildEngine, mcClient)
	}
}

func handleRequest(ctx context.Context, prompt string,
	agentEngine *agent.Engine,
	buildEngine *builder.Builder,
	mcClient *netherite_mc.Client) {

	fmt.Println("正在分析需求...")

	task, err := agentEngine.Execute(ctx, prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "处理失败: %v\n", err)
		return
	}

	fmt.Printf("任务 ID: %s\n", task.ID)
	fmt.Printf("结果: %s\n", task.Result)
	fmt.Printf("状态: %s\n", task.Status)
	fmt.Println()
}

func listTasks(store *db.TaskStore) {
	tasks, err := store.ListAll(10)
	if err != nil {
		fmt.Fprintf(os.Stderr, "读取任务失败: %v\n", err)
		return
	}
	if len(tasks) == 0 {
		fmt.Println("暂无任务")
		return
	}
	for _, t := range tasks {
		status := t.Status
		if t.Error != "" {
			status = t.Error
		}
		fmt.Printf("[%s] %s | %s | %s\n",
			t.ID[:12], t.Prompt, status, t.CreatedAt.Format("01-02 15:04"))
	}
}

func listPlugins(m *plugin.Manager) {
	plugins := m.List()
	if len(plugins) == 0 {
		fmt.Println("暂无插件")
		return
	}
	for _, p := range plugins {
		status := "停止"
		if p.Cmd != nil && p.Cmd.Process != nil {
			status = "运行中"
		}
		fmt.Printf("[%s] %s | %s | %s\n", p.ID, p.Name, status, p.Description)
	}
}

func printBanner() {
	fmt.Println()
	fmt.Println("    ░█▀▀░█░█░█▀▀░░░█░░░█▀▀░█▀▄░█▄█░█▀█░█▀▄░░░█▀▀░█▀▀░█▀▄░█▀▀░█▀▄")
	fmt.Println("    ░█▀▀░█░█░▀▀█░░░█░░░█░░░█▀▄░█░█░█░█░█▀▄░░░█░░░█▀▀░█▀▄░▀▀█░█▀▄")
	fmt.Println("    ░▀▀▀░▀▀▀░▀▀▀░░░▀░░░▀▀▀░▀░▀░▀░▀░▀░░░▀░░░░░▀▀▀░▀░░░▀░▀░▀░░")
	fmt.Println()
}

func printHelp() {
	fmt.Println("冰点咖啡命令:")
	fmt.Println("  /tasks     查看最近任务")
	fmt.Println("  /plugins   查看插件")
	fmt.Println("  /status    查看连接状态")
	fmt.Println("  /connect   连接服务器")
	fmt.Println("  /disconnect 断开连接")
	fmt.Println("  /help      显示帮助")
	fmt.Println("  /quit      退出")
	fmt.Println()
	fmt.Println("建筑请求格式:")
	fmt.Println("  build type:block_name width:N height:N depth:N radius:N")
	fmt.Println("  示例: build house width:10 height:5")
}

func expandHome(path string) string {
	if len(path) > 0 && path[0] == '~' {
		home, _ := os.UserHomeDir()
		if home != "" {
			return home + path[1:]
		}
	}
	return path
}

func dir(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}

func trimSpace(s string) string {
	for len(s) > 0 && (s[0] == ' ' || s[0] == '\t' || s[0] == '\n' || s[0] == '\r') {
		s = s[1:]
	}
	for len(s) > 0 && (s[len(s)-1] == ' ' || s[len(s)-1] == '\t' || s[len(s)-1] == '\n' || s[len(s)-1] == '\r') {
		s = s[:len(s)-1]
	}
	return s
}

func init() {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		log.Println("Shutting down...")
		os.Exit(0)
	}()
}
