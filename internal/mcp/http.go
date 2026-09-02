package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"

	"github.com/Verify144/IcePointCoffee/internal/ai"
)

// HTTPHandler HTTP+SSE 端点（可选）
// Server 获取底层 MCP Server
func (h *HTTPHandler) Server() *Server { return h.server }
// 模式 1：POST /mcp         传统 HTTP 模式（请求-响应）
// 模式 2：GET  /mcp/stream  SSE 流（双向通信）
type HTTPHandler struct {
	server *Server
	// SSE 流客户端
	mu       sync.RWMutex
	clients  map[chan []byte]bool
}

// NewHTTPHandler 创建 HTTP handler
func NewHTTPHandler(server *Server) *HTTPHandler {
	return &HTTPHandler{
		server:  server,
		clients: make(map[chan []byte]bool),
	}
}

// RegisterRoutes 注册到 mux
func (h *HTTPHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/mcp", h.handleRequest)
	mux.HandleFunc("/mcp/info", h.handleInfo)
	mux.HandleFunc("/mcp/stream", h.handleStream)
	mux.HandleFunc("/mcp/notify", h.handleNotify)
}

// handleInfo 简单 GET 返回服务器信息
func (h *HTTPHandler) handleInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"name":            h.server.name,
		"version":         h.server.version,
		"protocolVersion": ProtocolVersion,
		"capabilities": ServerCap{
			Tools:     &ToolsCap{ListChanged: true},
			Resources: &ResourcesCap{ListChanged: true},
		},
		"endpoints": map[string]string{
			"info":   "GET  /mcp/info",
			"tools":  "POST /mcp (method: tools/list or tools/call)",
			"stream": "GET  /mcp/stream (SSE)",
			"notify": "POST /mcp/notify (server notifications)",
		},
	})
}

// handleRequest 主入口：POST /mcp
func (h *HTTPHandler) handleRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, MCP-Protocol-Version")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS")

	// OPTIONS CORS
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Read error", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	ctx := r.Context()
	// MCP 协议版本头
	if v := r.Header.Get("MCP-Protocol-Version"); v != "" {
		// 校验版本（略）
		_ = v
	}

	resp := h.server.Handle(ctx, body)
	if resp == nil {
		// Notification
		w.WriteHeader(http.StatusNoContent)
		return
	}
	w.Write(resp)
}

// handleStream SSE 双向流
func (h *HTTPHandler) handleStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "GET only", http.StatusMethodNotAllowed)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "SSE not supported", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ch := make(chan []byte, 16)
	h.mu.Lock()
	h.clients[ch] = true
	h.mu.Unlock()

	defer func() {
		h.mu.Lock()
		delete(h.clients, ch)
		h.mu.Unlock()
		close(ch)
	}()

	for {
		select {
		case <-r.Context().Done():
			return
		case data := <-ch:
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
		}
	}
}

// handleNotify 触发服务器主动通知（用于测试 / 集成）
func (h *HTTPHandler) handleNotify(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}
	body, _ := io.ReadAll(r.Body)
	defer r.Body.Close()
	h.Broadcast(body)
	w.WriteHeader(http.StatusOK)
}

// Broadcast 广播消息到所有 SSE 客户端
func (h *HTTPHandler) Broadcast(data []byte) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	for ch := range h.clients {
		select {
		case ch <- data:
		default:
			// 客户端太慢，丢弃
		}
	}
}

// 工具源桥接：把 ai.ToolRegistry 转成 ToolExecutor
type RegistrySource struct {
	Registry *ai.ToolRegistry
}

func (s *RegistrySource) List() []ToolExecutor {
	tools := s.Registry.List()
	out := make([]ToolExecutor, 0, len(tools))
	for _, t := range tools {
		out = append(out, &toolAdapter{tool: t})
	}
	return out
}

// toolAdapter 适配 ai.Tool → ToolExecutor
type toolAdapter struct {
	tool ai.Tool
}

func (a *toolAdapter) Name() string { return a.tool.Name() }
func (a *toolAdapter) Description() string { return a.tool.Description() }
func (a *toolAdapter) Parameters() map[string]interface{} { return a.tool.Parameters() }
func (a *toolAdapter) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	return a.tool.Execute(ctx, args)
}
