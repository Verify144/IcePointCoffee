package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/ai"
	"github.com/Verify144/IcePointCoffee/internal/builder"
	"github.com/Verify144/IcePointCoffee/internal/cron"
	"github.com/Verify144/IcePointCoffee/internal/metrics"
	"github.com/Verify144/IcePointCoffee/internal/task"
	"github.com/Verify144/IcePointCoffee/internal/template"
)

// Server HTTP RPC 服务
type Server struct {
	mu        sync.RWMutex
	port      int
	mux       *http.ServeMux
	server    *http.Server
	aiClient  *ai.Client
	registry  *ai.ToolRegistry
	memory    *ai.Memory
	builder   *builder.Builder
	plugins   map[string]interface{}
	commands  []string
	rateLimit *RateLimiter
	taskManager *task.Manager
	events    []Event
	maxEvents int

	// 模板
	templateStore *template.Store

	// Cron
	cronStore *cron.Store
	cronSched *cron.Scheduler

	// 流式会话管理（用于 AI 流式取消）
	streamCancels map[string]context.CancelFunc
	streamMu      sync.RWMutex
}

// NewServer 创建服务器（使用内存存储，无认证/模板/Cron）
func NewServer(port int) *Server {
	return NewServerWithStore(port, nil)
}

// NewServerWithStore 创建服务器，使用指定的 task store（nil=内存存储）
func NewServerWithStore(port int, store task.Store) *Server {
	if store == nil {
		store = task.NewListStore()
	}
	taskMgr := task.NewManager(store, 4)

	s := &Server{
		port:          port,
		mux:           http.NewServeMux(),
		registry:      ai.NewToolRegistry(),
		memory:        ai.NewMemory(50),
		builder:       builder.New(),
		plugins:       make(map[string]interface{}),
		maxEvents:     100,
		rateLimit:     NewRateLimiter(100, time.Minute),
		taskManager:   taskMgr,
		streamCancels: make(map[string]context.CancelFunc),
	}
	taskMgr.Start()
	s.SetupRoutes()
	s.registerHandlers()
	return s
}

// NewServerWithDB 创建带完整功能的服务器（认证+模板+Cron）
func NewServerWithDB(port int, db *sql.DB) *Server {
	s := NewServerWithStore(port, nil)
	if db != nil {
		s.initFullFeatures(db)
		// 重新设置路由以包含 auth/templates/cron
		s.SetupRoutes()
	}
	return s
}

func (s *Server) initFullFeatures(db *sql.DB) {
	// 模板
	s.templateStore = template.NewStore(db)
	s.templateStore.SeedPublicTemplates() // 幂等初始化公开模板

	// Cron
	s.cronStore = cron.NewStore(db)
	s.cronSched = cron.NewScheduler(s.cronStore, s.taskManager)
	s.cronSched.Start()
}

// authTokenHandlers 占位（已删除认证）
func (s *Server) authTokenHandlers() http.Handler {
	return http.NewServeMux()
}

func (s *Server) registerHandlers() {

	// 注册内置任务 handler
	s.taskManager.Register("echo", func(ctx context.Context, t *task.Task) error {
		if msg, ok := t.Payload["message"].(string); ok {
			t.Result = "Echo: " + msg
		}
		return nil
	})
	s.taskManager.Register("delay", func(ctx context.Context, t *task.Task) error {
		seconds := 1
		if v, ok := t.Payload["seconds"].(float64); ok {
			seconds = int(v)
		}
		for i := 0; i < seconds; i++ {
			select {
			case <-ctx.Done():
				return task.ErrCancelled
			case <-time.After(time.Second):
			}
			t.Progress = (i + 1) * 100 / seconds
		}
		t.Result = fmt.Sprintf("等待 %d 秒完成", seconds)
		return nil
	})
}

// SetupRoutes 设置路由（幂等）
func (s *Server) SetupRoutes() {
	// 如果已设置过，替换 mux 以避免冲突
	s.mu.Lock()
	s.mux = http.NewServeMux()
	s.mu.Unlock()

	// 健康检查
	s.mux.HandleFunc("/health", s.handleHealth)

	// Prometheus Metrics
	s.mux.Handle("/metrics", s.MetricsHandler())

	// API 文档
	s.mux.HandleFunc("/api/docs", s.handleAPIDocs)

	// ==== API v1 ====
	api := http.NewServeMux()
	api.Handle("/ai/chat", s.rateLimit.Middleware(http.HandlerFunc(s.handleAIChat)))
	// AI 流式对话
	api.Handle("/ai/chat/stream", s.rateLimit.Middleware(http.HandlerFunc(s.handleAIChatStream)))
	// AI 流式取消
	api.HandleFunc("/ai/chat/stream/", s.handleAIStreamCancel)
	api.HandleFunc("/ai/tools", s.handleAITools)
	api.HandleFunc("/ai/memory", s.handleAIMemory)

	// 任务管理
	// 同时注册 /tasks 和 /tasks/ 以支持子路径匹配
	// /tasks 用于精确匹配（列表/提交）
	// /tasks/ 用于子路径匹配（详情/取消/统计）
	api.HandleFunc("/tasks", s.handleTasksRoot)
	api.HandleFunc("/tasks/", s.handleTasksRoot)

	api.Handle("/commands", s.rateLimit.Middleware(http.HandlerFunc(s.handleCommands)))
	api.HandleFunc("/events", s.handleEvents)
	api.HandleFunc("/plugins", s.handlePlugins)
	api.HandleFunc("/plugins/register", s.handlePluginRegister)
	api.Handle("/build", s.rateLimit.Middleware(http.HandlerFunc(s.handleBuild)))
	api.HandleFunc("/status", s.handleStatus)
	s.mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

	// 模板/Cron 路由（仅在 initFullFeatures 之后可用）
	// 认证已删除，冰点咖啡作为本地工具不需要用户系统
	if s.templateStore != nil {
		s.handleTemplateRoutes(api)
	}
	if s.cronStore != nil {
		s.handleCronRoutes(api)
	}

	// 管理员接口
	admin := http.NewServeMux()
	admin.HandleFunc("/reload", s.handleReload)
	admin.HandleFunc("/restart", s.handleRestart)
	admin.HandleFunc("/stats", s.handleStats)
	s.mux.Handle("/api/admin/", http.StripPrefix("/api/admin", admin))

	// 静态文件（Dashboard）
	s.mux.HandleFunc("/", s.handleDashboard)
}

// Start 启动服务器
func (s *Server) Start() error {
	s.SetupRoutes()

	// 初始化 metrics
	metrics.InitDefault()

	// 注册内置工具
	s.registry.Register(&ai.EchoTool{})
	s.registry.Register(&ai.GetTimeTool{})
	s.registry.Register(&ai.CalculateTool{})

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.MetricsMiddleware(s.withCORS(s.mux)),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("💻 HTTP Server 启动于 :%d", s.port)
	log.Printf("📊 Dashboard: http://localhost:%d/", s.port)
	log.Printf("🤖 AI Chat: POST http://localhost:%d/api/v1/ai/chat", s.port)
	log.Printf("📡 Events: GET http://localhost:%d/api/v1/events", s.port)
	log.Printf("📈 Metrics: http://localhost:%d/metrics", s.port)

	return s.server.ListenAndServe()
}

// Stop 停止服务器
func (s *Server) Stop() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// ==== Handlers ====

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"status":  "ok",
		"service": "IcePointCoffee",
		"version": "1.0",
		"time":    time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleAPIDocs(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(GetAPIDoc())
}

func (s *Server) handleAIChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	defer func() {
		metrics.DefaultBusiness.IncAIChat(true, 0)
	}()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	var req struct {
		Message   string `json:"message"`
		System    string `json:"system"`
		UseTools  bool   `json:"use_tools"`
		SessionID string `json:"session_id"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.memory.Add(ai.Message{Role: "user", Content: req.Message})

	var response string
	if s.aiClient != nil {
		msgs := s.memory.Get()
		if req.UseTools {
			res, err := s.aiClient.ChatWithTools(context.Background(), msgs, s.registry.List())
			if err != nil {
				response = fmt.Sprintf("AI 错误: %v", err)
			} else {
				response = res
			}
		} else {
			res, err := s.aiClient.Chat(context.Background(), msgs)
			if err != nil {
				response = fmt.Sprintf("AI 错误: %v", err)
			} else {
				response = res
			}
		}
	} else {
		response = mockAIResponse(req.Message)
	}

	s.memory.Add(ai.Message{Role: "assistant", Content: response})
	s.addEvent("ai_chat", map[string]string{"message": req.Message, "response": response})

	// AI metrics
	aiSuccess := !strings.HasPrefix(response, "AI 错误")
	metrics.DefaultBusiness.IncAIChat(aiSuccess, len(response))

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"response": response,
		"tools":    s.registry.ToOpenAI(),
	})
}

func (s *Server) handleAITools(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"tools":  s.registry.List(),
		"openai": s.registry.ToOpenAI(),
	})
}

func (s *Server) handleAIMemory(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		s.memory.Clear()
		jsonResponse(w, http.StatusOK, map[string]string{"status": "cleared"})
		return
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"messages": s.memory.Get(),
		"size":     s.memory.Size(),
	})
}

func (s *Server) handleCommands(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		cmds := s.commands
		s.mu.RUnlock()
		jsonResponse(w, http.StatusOK, map[string]interface{}{
			"commands": cmds,
			"count":    len(cmds),
		})
	case http.MethodPost:
		var req struct{ Command string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		s.commands = append(s.commands, req.Command)
		s.mu.Unlock()
		s.addEvent("command", map[string]string{"cmd": req.Command})
		metrics.DefaultBusiness.IncCommand(true)
		jsonResponse(w, http.StatusOK, map[string]string{"status": "queued", "command": req.Command})
	default:
		http.Error(w, "GET or POST", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	metrics.DefaultRegistry.Inc("icepoint_event_stream_connections", 1)
	defer metrics.DefaultRegistry.Inc("icepoint_event_stream_connections", -1)

	ctx := r.Context()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	fmt.Fprintf(w, "data: {\"type\":\"ping\",\"time\":\"%s\"}\n\n", time.Now().Format(time.RFC3339))
	flusher.Flush()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.mu.RLock()
			events := make([]Event, len(s.events))
			copy(events, s.events)
			s.mu.RUnlock()
			for _, e := range events {
				data, _ := json.Marshal(e)
				fmt.Fprintf(w, "data: %s\n\n", data)
				metrics.DefaultRegistry.Inc("icepoint_event_stream_messages_total", 1)
			}
			flusher.Flush()
		}
	}
}

func (s *Server) handlePlugins(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	plugins := make([]map[string]interface{}, 0, len(s.plugins))
	for name := range s.plugins {
		plugins = append(plugins, map[string]interface{}{
			"name":    name,
			"enabled": true,
		})
	}
	jsonResponse(w, http.StatusOK, map[string]interface{}{"plugins": plugins})
}

func (s *Server) handlePluginRegister(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Name string `json:"name"`
		Type string `json:"type"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.mu.Lock()
	s.plugins[req.Name] = nil
	s.mu.Unlock()

	s.addEvent("plugin_register", map[string]string{"name": req.Name})
	jsonResponse(w, http.StatusOK, map[string]string{"status": "registered", "name": req.Name})
}

func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type   string                 `json:"type"`
		Size   int                    `json:"size"`
		X      int                    `json:"x"`
		Y      int                    `json:"y"`
		Z      int                    `json:"z"`
		Blocks map[string]interface{} `json:"blocks"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result, err := s.builder.Build(req.Type, map[string]interface{}{
		"size":   req.Size,
		"x":      req.X,
		"y":      req.Y,
		"z":      req.Z,
		"blocks": req.Blocks,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	blockCount := estimateBlockCount(result)
	metrics.DefaultBusiness.IncBuild(req.Type, blockCount)

	s.addEvent("build", map[string]string{"type": req.Type, "result": result})
	metrics.DefaultBusiness.IncBuild(req.Type, estimateBlockCount(result))
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"type":   req.Type,
		"result": result,
		"blocks": result,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// 更新 Gauge 指标
	metrics.DefaultBusiness.SetPluginCount(len(s.plugins))
	metrics.DefaultBusiness.SetMemorySize(s.memory.Size())

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"uptime":      time.Since(time.Now().Add(-1 * time.Hour)).String(),
		"connections": 0,
		"commands":    len(s.commands),
		"events":      len(s.events),
		"plugins":     len(s.plugins),
		"ai_enabled":  s.aiClient != nil,
		"memory_size": s.memory.Size(),
		"tools_count": len(s.registry.List()),
	})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	log.Println("🔄 Reload requested")
	s.addEvent("reload", nil)
	jsonResponse(w, http.StatusOK, map[string]string{"status": "reloading"})
}

func (s *Server) handleRestart(w http.ResponseWriter, r *http.Request) {
	log.Println("🔄 Restart requested")
	s.addEvent("restart", nil)
	jsonResponse(w, http.StatusOK, map[string]string{"status": "restarting"})
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(dashboardHTMLV2))
}

// ==== Middleware ====

func (s *Server) withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}


func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func (s *Server) addEvent(eventType string, data interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, Event{
		Type:      eventType,
		Data:      data,
		Timestamp: time.Now(),
	})
	if len(s.events) > s.maxEvents {
		s.events = s.events[len(s.events)-s.maxEvents:]
	}
}

// estimateBlockCount 从建筑结果字符串估算方块数
func estimateBlockCount(result string) int {
	count := 0
	for _, line := range strings.Split(result, "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "✅") {
			var n int
			if _, err := fmt.Sscanf(line, "✅ 生成 %d", &n); err == nil {
				count = n
			}
		}
	}
	if count == 0 {
		count = 100
	}
	return count
}

// mockAIResponse Mock AI 回复
func mockAIResponse(msg string) string {
	msg = strings.ToLower(msg)
	switch {
	case strings.Contains(msg, "hello") || strings.Contains(msg, "hi"):
		return "你好！我是 IcePointCoffee AI 助手。有什么我可以帮你的吗？"
	case strings.Contains(msg, "help"):
		return "我可以帮你：\n1. 生成建筑 (house/tower/circle/sphere)\n2. 执行命令\n3. 计算数学\n4. 查询时间\n5. 回答问题"
	case strings.Contains(msg, "status"):
		return "服务器运行正常！"
	default:
		return fmt.Sprintf("收到消息: %s\n(当前为 Mock 模式，配置 AI API 可获得真实回复)", msg)
	}
}

// Event 事件
type Event struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}


// ==== JSON helpers ====

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeJSONError(w http.ResponseWriter, status int, err, msg string) {
	writeJSON(w, status, map[string]interface{}{
		"error":   err,
		"message": msg,
	})
}
