package server

import (
	"bytes"
	"strings"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/ai"
	"github.com/Verify144/IcePointCoffee/internal/builder"
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
	events    []Event
	maxEvents int
}

// NewServer 创建服务器
func NewServer(port int) *Server {
	return &Server{
		port:      port,
		mux:       http.NewServeMux(),
		registry:  ai.NewToolRegistry(),
		memory:    ai.NewMemory(50),
		builder:   builder.New(),
		plugins:   make(map[string]interface{}),
		maxEvents: 100,
		rateLimit: NewRateLimiter(100, time.Minute), // 每分钟 100 请求
	}
}

// SetupRoutes 设置路由
func (s *Server) SetupRoutes() {
	// 健康检查
	s.mux.HandleFunc("/health", s.handleHealth)

	// API 文档
	s.mux.HandleFunc("/api/docs", s.handleAPIDocs)

	// ==== API v1 ====
	api := http.NewServeMux()
	
	// AI 对话（限流）
	api.Handle("/ai/chat", s.rateLimit.Middleware(http.HandlerFunc(s.handleAIChat)))
	api.HandleFunc("/ai/tools", s.handleAITools)
	api.HandleFunc("/ai/memory", s.handleAIMemory)

	// 命令（限流）
	api.Handle("/commands", s.rateLimit.Middleware(http.HandlerFunc(s.handleCommands)))

	// 事件流 (SSE) - 不限流
	api.HandleFunc("/events", s.handleEvents)

	// 插件管理
	api.HandleFunc("/plugins", s.handlePlugins)
	api.HandleFunc("/plugins/register", s.handlePluginRegister)

	// 建筑（限流）
	api.Handle("/build", s.rateLimit.Middleware(http.HandlerFunc(s.handleBuild)))

	// 状态
	api.HandleFunc("/status", s.handleStatus)

	// 嵌入 /api 前缀
	s.mux.Handle("/api/v1/", http.StripPrefix("/api/v1", api))

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

	// 注册内置工具
	s.registry.Register(&ai.EchoTool{})
	s.registry.Register(&ai.GetTimeTool{})
	s.registry.Register(&ai.CalculateTool{})

	s.server = &http.Server{
		Addr:         fmt.Sprintf(":%d", s.port),
		Handler:      s.withCORS(s.mux),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	log.Printf("💻 HTTP Server 启动于 :%d", s.port)
	log.Printf("📊 Dashboard: http://localhost:%d/", s.port)
	log.Printf("🤖 AI Chat: POST http://localhost:%d/api/v1/ai/chat", s.port)
	log.Printf("📡 Events: GET http://localhost:%d/api/v1/events", s.port)

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

	// 保存用户消息
	s.memory.Add(ai.Message{Role: "user", Content: req.Message})
	defer func() {
		s.memory.Add(ai.Message{Role: "assistant", Content: req.Message})
	}()

	// 使用 AI 或简单回复
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
		// Mock AI 回复
		response = mockAIResponse(req.Message)
	}

	// 添加工具调用结果
	s.memory.Add(ai.Message{Role: "assistant", Content: response})

	s.addEvent("ai_chat", map[string]string{"message": req.Message, "response": response})

	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"response": response,
		"tools":    s.registry.ToOpenAI(),
	})
}

func (s *Server) handleAITools(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"tools": s.registry.List(),
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
		jsonResponse(w, http.StatusOK, map[string]string{"status": "queued", "command": req.Command})
	default:
		http.Error(w, "GET or POST", http.StatusMethodNotAllowed)
	}
}

func (s *Server) handleEvents(w http.ResponseWriter, r *http.Request) {
	// SSE 事件流
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	ctx := r.Context()
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	// 发送初始 ping
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
	defer s.mu.Unlock()
	s.plugins[req.Name] = nil

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

	s.addEvent("build", map[string]string{"type": req.Type, "result": result})
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"type":   req.Type,
		"result": result,
		"blocks": result,
	})
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

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
	go func() {
		time.Sleep(500 * time.Millisecond)
	}()
}

func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"memory":    memStats(),
	})
}

// ==== Dashboard ====

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// 简单内联 Dashboard
	html := dashboardHTMLV2
	if r.URL.Path != "/" && r.URL.Path != "/index.html" {
		// SPA 回退
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
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

// ==== Helpers ====

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

// memStats 简单内存统计
func memStats() map[string]interface{} {
	return map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
	}
}

// mockAIResponse Mock AI 回复
func mockAIResponse(msg string) string {
	msg = strings.ToLower(msg)
	switch {
	case contains(msg, "hello") || contains(msg, "hi"):
		return "你好！我是 IcePointCoffee AI 助手。有什么我可以帮你的吗？"
	case contains(msg, "help"):
		return "我可以帮你：\n1. 生成建筑 (house/tower/circle/sphere)\n2. 执行命令\n3. 计算数学\n4. 查询时间\n5. 回答问题"
	case contains(msg, "status"):
		return "服务器运行正常！"
	default:
		return fmt.Sprintf("收到消息: %s\n(当前为 Mock 模式，配置 AI API 可获得真实回复)", msg)
	}
}

func contains(s, substr string) bool {
	return bytes.Contains([]byte(s), []byte(substr))
}

// Event 事件
type Event struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
}
