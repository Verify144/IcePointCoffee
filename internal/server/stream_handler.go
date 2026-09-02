package server

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/ai"
	"github.com/Verify144/IcePointCoffee/internal/metrics"
	"github.com/Verify144/IcePointCoffee/internal/task"
)

// handleAIChatStream 流式 AI 对话（SSE）
func (s *Server) handleAIChatStream(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		SendError(w, http.StatusBadRequest, 400, "读取请求失败", err.Error())
		return
	}
	defer r.Body.Close()

	var req struct {
		Message   string `json:"message"`
		System    string `json:"system"`
		UseTools  bool   `json:"use_tools"`
		SessionID string `json:"session_id"` // 可选，客户端指定 session
	}
	if err := json.Unmarshal(body, &req); err != nil {
		SendError(w, http.StatusBadRequest, 400, "JSON 解析失败", err.Error())
		return
	}

	if req.Message == "" {
		SendError(w, http.StatusBadRequest, 400, "消息为空", "请提供 message 字段")
		return
	}

	// 生成或使用提供的 session ID
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = fmt.Sprintf("s_%d", time.Now().UnixNano())
	}

	// 设置 SSE 响应头
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("X-Accel-Buffering", "no")
	// 返回 session ID，让客户端知道如何取消
	w.Header().Set("X-Session-ID", sessionID)

	flusher, ok := w.(http.Flusher)
	if !ok {
		SendError(w, http.StatusInternalServerError, 500, "SSE 不支持", "")
		return
	}

	// 保存用户消息
	s.memory.Add(ai.Message{Role: "user", Content: req.Message})

	// 构造消息列表
	messages := s.memory.Get()

	// 工具列表
	tools := s.registry.List()

	// 流式响应
	var fullResponse strings.Builder

	// 创建可取消的 context
	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// 注册 cancel（支持外部主动取消）
	s.streamMu.Lock()
	s.streamCancels[sessionID] = cancel
	s.streamMu.Unlock()

	// 确保完成后清理
	defer func() {
		s.streamMu.Lock()
		delete(s.streamCancels, sessionID)
		s.streamMu.Unlock()
	}()

	sendChunk := func(chunk ai.StreamChunk) {
		data, _ := json.Marshal(chunk)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		if chunk.Type == "content" {
			fullResponse.WriteString(chunk.Content)
		}
	}

	// 发送 session ID 和开始事件
	sendChunk(ai.StreamChunk{
		Type:    "session",
		Content: sessionID,
	})
	sendChunk(ai.StreamChunk{Type: "start", Content: "生成中..."})

	// 执行流式调用
	var streamErr error
	if s.aiClient != nil {
		streamErr = s.aiClient.Stream(ctx, messages, tools, sendChunk)
	} else {
		// Mock 流式
		streamErr = ai.MockStream(ctx, messages, sendChunk)
	}

	if streamErr != nil && ctx.Err() != context.Canceled {
		sendChunk(ai.StreamChunk{Type: "error", Error: streamErr.Error()})
		metrics.DefaultBusiness.IncAIChat(false, fullResponse.Len())
		return
	}

	// 保存 AI 回复
	s.memory.Add(ai.Message{Role: "assistant", Content: fullResponse.String()})

	// 记录指标
	metrics.DefaultBusiness.IncAIChat(true, fullResponse.Len())
	metrics.DefaultBusiness.IncEvent("ai_chat_stream")
}

// handleAIStreamCancel 取消正在进行的 AI 流式会话
// DELETE /api/v1/ai/chat/stream/{session_id}
func (s *Server) handleAIStreamCancel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "DELETE only", http.StatusMethodNotAllowed)
		return
	}

	// 从 URL 中提取 session_id
	sessionID := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, "/"), "ai/chat/stream/")
	if sessionID == "" {
		SendError(w, http.StatusBadRequest, 400, "缺少 session_id", "")
		return
	}

	s.streamMu.Lock()
	cancel, ok := s.streamCancels[sessionID]
	delete(s.streamCancels, sessionID)
	s.streamMu.Unlock()

	if !ok {
		SendError(w, http.StatusNotFound, 404, "会话不存在或已结束", "")
		return
	}

	cancel() // 中断流式生成
	SendSuccess(w, map[string]string{"status": "cancelled", "session_id": sessionID})
}

// handleTaskSubmit 提交任务
func (s *Server) handleTaskSubmit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "POST only", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Type     string                 `json:"type"`
		Payload  map[string]interface{} `json:"payload"`
		Priority int                    `json:"priority"`
		UserID   string                 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		SendError(w, http.StatusBadRequest, 400, "JSON 解析失败", err.Error())
		return
	}

	if req.Type == "" {
		SendError(w, http.StatusBadRequest, 400, "类型为空", "请提供 type 字段")
		return
	}

	if s.taskManager == nil {
		SendError(w, http.StatusServiceUnavailable, 1007, "任务管理器未启动", "")
		return
	}

	// 创建任务
	t := task.NewTask(req.Type, req.Payload)
	if req.Priority > 0 {
		t.Priority = task.Priority(req.Priority)
	}
	t.UserID = req.UserID

	// 提交
	if err := s.taskManager.Submit(t); err != nil {
		SendError(w, http.StatusInternalServerError, 500, "提交任务失败", err.Error())
		return
	}

	SendSuccess(w, t)
}

// handleTaskGet 已迁移到 task_routes.go

// handleTaskList 列出任务
func (s *Server) handleTaskList(w http.ResponseWriter, r *http.Request) {
	status := r.URL.Query().Get("status")
	taskType := r.URL.Query().Get("type")
	userID := r.URL.Query().Get("user_id")
	limitStr := r.URL.Query().Get("limit")

	limit := 20
	if limitStr != "" {
		fmt.Sscanf(limitStr, "%d", &limit)
	}

	tasks, err := s.taskManager.List(task.ListFilter{
		Status: status,
		Type:   taskType,
		UserID: userID,
		Limit:  limit,
	})
	if err != nil {
		SendError(w, http.StatusInternalServerError, 500, "查询失败", err.Error())
		return
	}

	SendSuccess(w, map[string]interface{}{
		"tasks": tasks,
		"count": len(tasks),
	})
}

// handleTaskCancel 已迁移到 task_routes.go

// handleTaskStats 任务统计
func (s *Server) handleTaskStats(w http.ResponseWriter, r *http.Request) {
	stats, err := s.taskManager.Stats()
	if err != nil {
		SendError(w, http.StatusInternalServerError, 500, "统计失败", err.Error())
		return
	}

	SendSuccess(w, map[string]interface{}{
		"stats":   stats,
		"running": s.taskManager.Running(),
	})
}

// keep-alive ticker for SSE
func keepAlive(interval time.Duration) func() {
	ticker := time.NewTicker(interval)
	stop := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
			case <-stop:
				ticker.Stop()
				return
			}
		}
	}()
	return func() { close(stop) }
}
