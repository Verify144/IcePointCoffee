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
	netherite "github.com/Verify144/IcePointCoffee/internal/netherite"
	"github.com/Verify144/IcePointCoffee/internal/netherite/mc"
	"github.com/Verify144/IcePointCoffee/internal/netherite/protocol"
	"github.com/Verify144/IcePointCoffee/internal/plugin"
)

func main() {
	// 加载配置
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "配置错误: %v\n", err)
		fmt.Fprintln(os.Stderr, "\n请创建配置文件 ~/.icepoint/config.yaml，参考 config.example.yaml")
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
	aiClient := ai.NewClient(cfg.AI)
	agentEngine := agent.NewEngine(aiClient, taskStore)
	buildEngine := builder.NewBuilder()
	pluginManager := plugin.NewManager(cfg.Plugin.Dir)

	// 初始化 MC 客户端
	mcClient, err := mc.NewClient(&mc.Options{
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
		// 初始化事件处理器和异步执行器
		eventProcessor := mc.NewEventProcessor(mcClient)
		eventProcessor.Start()

		asyncExecutor := mc.NewAsyncExecutor(mcClient, 4)
		asyncExecutor.Start()
		defer asyncExecutor.Stop()

		// 设置事件回调
		eventProcessor.RegisterHandler(protocol.IDText, func(data []byte) error {
			text, err := protocol.DecodeText(data)
			if err != nil {
				return err
			}
			fmt.Printf("[聊天] %s\n", text.Message)
			return nil
		})

		eventProcessor.RegisterHandler(protocol.IDCommandOutput, func(data []byte) error {
			output, err := protocol.DecodeCommandOutput(data)
			if err != nil {
				return err
			}
			for _, msg := range output.Messages {
				fmt.Printf("  > %s\n", msg.Message)
			}
			return nil
		})

		eventProcessor.RegisterHandler(protocol.IDDisconnect, func(data []byte) error {
			fmt.Println("[MC] 服务器断开连接")
			return nil
		})
	}

	cmdExecutor := importer.NewExecutor(mcClient)

	// 启动插件
	plugins, err := pluginManager.Scan()
	if err != nil {
		fmt.Fprintf(os.Stderr, "插件扫描错误: %v\n", err)
	} else if len(plugins) > 0 {
		fmt.Printf("已扫描到 %d 个插件:\n", len(plugins))
		for _, p := range plugins {
			if err := pluginManager.Start(p.ID); err != nil {
				fmt.Fprintf(os.Stderr, "  启动 %s 失败: %v\n", p.Name, err)
			} else {
				fmt.Printf("  ✓ %s\n", p.Name)
			}
		}
	}

	// 启动 HTTP RPC 插件服务
	if cfg.Plugin.HTTPPort > 0 {
		httpServer := netherite.NewHTTPServer(&netherite.HTTPConfig{
			Port:    cfg.Plugin.HTTPPort,
			Timeout: 30 * time.Second,
		})
		if mcClient != nil {
			httpServer.SetConn(mcClient)
		}
		httpServer.Start()
		defer httpServer.Stop()
		fmt.Printf("HTTP RPC 插件服务: http://127.0.0.1:%d\n", cfg.Plugin.HTTPPort)
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

	// 主菜单
	fmt.Println()
	printBanner()
	fmt.Printf("冰点咖啡 v1.0.0\n")
	fmt.Printf("模型: %s @ %s\n", cfg.AI.Model, cfg.AI.BaseURL)
	if mcClient != nil {
		fmt.Printf("服务器: %s (%s)\n", mcClient.RemoteAddr(), cfg.Server.PlayerName)
	}
	if cfg.Plugin.HTTPPort > 0 {
		fmt.Printf("插件 HTTP: http://127.0.0.1:%d\n", cfg.Plugin.HTTPPort)
	}
	fmt.Printf("插件数: %d\n", len(plugins))
	fmt.Println()
	fmt.Println("输入你的建筑需求，例如：")
	fmt.Println("  build house width:10 height:5 depth:8 block:oak_planks")
	fmt.Println("  build tower height:30 block:stone_bricks")
	fmt.Println("  描述: 做一个 5x5 的泥土农场")
	fmt.Println()
	fmt.Println("命令: /tasks  /plugins  /connect  /disconnect  /quit")

	for {
		fmt.Print("❄ > ")
		var input string
		if _, err := fmt.Fscan(os.Stdin, &input); err != nil {
			break
		}
		input = trimSpace(input)
		if input == "" {
			continue
		}

		switch {
		case input == "/quit", input == "/exit", input == "quit":
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
		case input == "/status":
			if mcClient != nil && mcClient.IsConnected() {
				fmt.Println("状态: 已连接")
			} else {
				fmt.Println("状态: 未连接")
			}
			continue
		}

		// 处理建筑请求
		handleRequest(ctx, input, agentEngine, buildEngine, cmdExecutor, mcClient)
	}
}

func handleRequest(ctx context.Context, prompt string,
	agentEngine *agent.Engine,
	buildEngine *builder.Builder,
	cmdExecutor *importer.Executor,
	mcClient *mc.Client) {

	fmt.Println("正在分析需求...")

	task, err := agentEngine.Handle(ctx, "cli_user", prompt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "处理失败: %v\n", err)
		return
	}

	fmt.Printf("任务 ID: %s\n", task.ID)
	fmt.Printf("描述: %s\n", task.Description)
	fmt.Printf("类型: %s\n", task.Type)
	fmt.Printf("状态: %s\n", task.Status)
	fmt.Println()

	if task.Type == "command" && len(task.Commands) > 0 {
		fmt.Printf("将执行 %d 条指令:\n", len(task.Commands))
		for i, cmd := range task.Commands {
			fmt.Printf("  [%d] %s\n", i+1, cmd)
		}
		fmt.Print("\n确认执行? (y/n): ")
		var confirm string
		fmt.Fscan(os.Stdin, &confirm)
		if confirm != "y" && confirm != "Y" {
			fmt.Println("已取消")
			return
		}

		// 执行指令
		for i, cmd := range task.Commands {
			fmt.Printf("  执行 [%d/%d]: %s\n", i+1, len(task.Commands), cmd)
			if mcClient != nil && mcClient.IsConnected() {
				result, err := mcClient.SendCommand(ctx, cmd)
				if err != nil {
					fmt.Fprintf(os.Stderr, "  执行失败: %v\n", err)
					break
				}
				fmt.Printf("  结果: %d 成功\n", result.SuccessCount)
			} else {
				fmt.Fprintf(os.Stderr, "  未连接到服务器\n")
				break
			}
		}
		fmt.Println("执行完成!")
	} else if task.Type == "structure" || task.Type == "import" {
		if len(task.Commands) > 0 && mcClient != nil && mcClient.IsConnected() {
			fmt.Println("使用 structure load 指令...")
			for _, cmd := range task.Commands {
				mcClient.SendCommand(ctx, cmd)
			}
		}
		fmt.Println("建筑已生成，待导入服务器")
	}
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
		fmt.Printf("[%s] %s | %s | %s | %s\n",
			t.ID[:12], t.Prompt, t.Type, t.Status, t.CreatedAt.Format("01-02 15:04"))
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
	fmt.Println("  示例: build house width:10 height:5 block:oak_planks")
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

// Init sets up signal handling.
func init() {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
		<-ch
		log.Println("Shutting down...")
		os.Exit(0)
	}()
}
