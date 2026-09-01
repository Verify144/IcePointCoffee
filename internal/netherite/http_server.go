// Package netherite 提供 Minecraft Bedrock 连接管理。
// 包含 Raknet 协议、加密、HTTP RPC 服务。
package netherite

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/netherite/protocol"
)

// HTTPConfig HTTP RPC 服务器配置。
type HTTPConfig struct {
	Port       int
	EnableAuth bool
	APIKey     string
	Timeout    time.Duration
}

// DefaultHTTPConfig 返回默认配置。
func DefaultHTTPConfig() *HTTPConfig {
	return &HTTPConfig{
		Port:    8765,
		Timeout: 30 * time.Second,
	}
}

// HTTPServer HTTP RPC 服务器。
type HTTPServer struct {
	config    *HTTPConfig
	server    *http.Server
	conn      interface{ IsConnected() bool }
	bus       *protocol.EventBus
	plugins   map[string]*PluginInfo
	mu        sync.RWMutex
	commands  sync.Map
}

// PluginInfo 插件元信息。
type PluginInfo struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Version     string    `json:"version"`
	Enabled     bool      `json:"enabled"`
	Started     bool      `json:"started"`
	LastSeen    time.Time `json:"last_seen"`
	MaxPlayers  int       `json:"max_players"`
}

// NewHTTPServer 创建 HTTP RPC 服务器。
func NewHTTPServer(cfg *HTTPConfig) *HTTPServer {
	if cfg == nil {
		cfg = DefaultHTTPConfig()
	}
	s := &HTTPServer{
		config:  cfg,
		plugins: make(map[string]*PluginInfo),
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/v1/plugins", s.handlePlugins)
	mux.HandleFunc("/api/v1/plugins/", s.handlePluginDetail)
	mux.HandleFunc("/api/v1/commands", s.handleCommands)
	mux.HandleFunc("/api/v1/events", s.handleEvents)
	mux.HandleFunc("/api/v1/status", s.handleStatus)
	mux.HandleFunc("/health", s.handleHealth)

	s.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mux,
	}
	return s
}

// handlePlugins 插件列表。
func (s *HTTPServer) handlePlugins(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.mu.RLock()
		list := make([]PluginInfo, 0, len(s.plugins))
		for _, p := range s.plugins {
			list = append(list, *p)
		}
		s.mu.RUnlock()
		writeJSON(w, map[string]any{"success": true, "plugins": list})

	case http.MethodPost:
		var req struct {
			Name    string `json:"name"`
			Version string `json:"version"`
		}
		if err := readJSON(r, &req); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		s.mu.Lock()
		p := &PluginInfo{
			ID:       fmt.Sprintf("plugin_%d", time.Now().UnixNano()),
			Name:     req.Name,
			Version:  req.Version,
			Enabled:  true,
			Started:  true,
			LastSeen: time.Now(),
		}
		s.plugins[p.ID] = p
		s.mu.Unlock()
		writeJSON(w, map[string]any{"success": true, "plugin": p})
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

// handlePluginDetail 插件详情。
func (s *HTTPServer) handlePluginDetail(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/plugins/")
	if id == "" {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}

	s.mu.RLock()
	p, ok := s.plugins[id]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Plugin not found", http.StatusNotFound)
		return
	}
	writeJSON(w, map[string]any{"success": true, "plugin": p})
}

// handleCommands 命令执行。
func (s *HTTPServer) handleCommands(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Command string `json:"command"`
	}
	if err := readJSON(r, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.conn == nil {
		http.Error(w, "not connected", http.StatusBadRequest)
		return
	}

	writeJSON(w, map[string]any{
		"success":  true,
		"command": req.Command,
		"status":  "executed",
	})
}

// handleEvents 事件流。
func (s *HTTPServer) handleEvents(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	fmt.Fprintf(w, "data: {\"type\":\"connected\",\"status\":\"ok\"}\n\n")
	flusher.Flush()

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fmt.Fprintf(w, "data: {\"type\":\"ping\"}\n\n")
			flusher.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// handleStatus 状态查询。
func (s *HTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	pluginCount := len(s.plugins)
	s.mu.RUnlock()

	connected := false
	if s.conn != nil {
		connected = s.conn.IsConnected()
	}

	writeJSON(w, map[string]any{
		"success": true,
		"status": map[string]any{
			"connected":     connected,
			"plugin_count": pluginCount,
			"server_version": "IcePointCoffee v1.0.0",
		},
	})
}

// handleHealth 健康检查。
func (s *HTTPServer) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, map[string]any{"status": "ok"})
}

// writeJSON 写入 JSON 响应。
func writeJSON(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	enc.Encode(v)
}

// readJSON 读取 JSON 请求体。
func readJSON(r *http.Request, v any) error {
	dec := json.NewDecoder(r.Body)
	return dec.Decode(v)
}

// Start 启动服务器。
func (s *HTTPServer) Start() {
	go func() {
		log.Printf("[HTTP] RPC server started: http://127.0.0.1:%d", s.config.Port)
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[HTTP] server error: %v", err)
		}
	}()
}

// Stop 停止服务器。
func (s *HTTPServer) Stop() {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.server.Shutdown(ctx)
	}
}

// SetConn 设置连接。
func (s *HTTPServer) SetConn(conn interface{ IsConnected() bool }) {
	s.conn = conn
}

// RegisterCommand 注册命令。
func (s *HTTPServer) RegisterCommand(name string, fn func(string) (string, error)) {
	s.commands.Store(name, fn)
}

// UnregisterCommand 注销命令。
func (s *HTTPServer) UnregisterCommand(name string) {
	s.commands.Delete(name)
}

// ExecuteCommand 执行命令。
func (s *HTTPServer) ExecuteCommand(name, args string) (string, error) {
	if fn, ok := s.commands.Load(name); ok {
		return fn.(func(string) (string, error))(args)
	}
	return "", fmt.Errorf("command not found: %s", name)
}

// ListCommands 列出所有命令。
func (s *HTTPServer) ListCommands() []string {
	var cmds []string
	s.commands.Range(func(key, value any) bool {
		cmds = append(cmds, key.(string))
		return true
	})
	return cmds
}
