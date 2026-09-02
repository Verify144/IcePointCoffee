package server

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/ai"
	"github.com/Verify144/IcePointCoffee/internal/task"
)

func TestHandleTaskSubmitAndGet(t *testing.T) {
	s := NewServer(8080)
	s.SetupRoutes()

	// 提交任务
	payload := map[string]interface{}{
		"type":    "echo",
		"payload": map[string]interface{}{"message": "hello"},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Submit should be 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp SuccessResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	if !resp.Success {
		t.Error("Submit should be success")
	}

	// 提取任务 ID
	taskData, _ := json.Marshal(resp.Data)
	var tk task.Task
	json.Unmarshal(taskData, &tk)
	if tk.ID == "" {
		t.Fatal("Task ID should be set")
	}

	// 等待执行
	time.Sleep(500 * time.Millisecond)

	// 获取任务
	req2 := httptest.NewRequest("GET", "/api/v1/tasks/"+tk.ID, nil)
	rec2 := httptest.NewRecorder()
	s.mux.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Get should be 200, got %d", rec2.Code)
	}
}

func TestHandleTaskList(t *testing.T) {
	s := NewServer(8080)
	s.SetupRoutes()

	// 提交 3 个任务
	for i := 0; i < 3; i++ {
		payload := map[string]interface{}{
			"type":    "echo",
			"payload": map[string]interface{}{"message": fmt.Sprintf("msg-%d", i)},
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		s.mux.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("Submit %d failed: %d", i, rec.Code)
		}
	}

	// 列表
	req := httptest.NewRequest("GET", "/api/v1/tasks", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("List should be 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "tasks") {
		t.Error("Response should contain tasks")
	}
}

func TestHandleTaskCancel(t *testing.T) {
	s := NewServer(8080)
	s.SetupRoutes()

	// 提交一个长任务
	payload := map[string]interface{}{
		"type":    "delay",
		"payload": map[string]interface{}{"seconds": 5.0},
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	var resp SuccessResponse
	json.Unmarshal(rec.Body.Bytes(), &resp)
	taskData, _ := json.Marshal(resp.Data)
	var tk task.Task
	json.Unmarshal(taskData, &tk)

	// 等待开始执行
	time.Sleep(200 * time.Millisecond)

	// 取消
	req2 := httptest.NewRequest("POST", "/api/v1/tasks/"+tk.ID+"/cancel", nil)
	rec2 := httptest.NewRecorder()
	s.mux.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Cancel should be 200, got %d", rec2.Code)
	}

	// 验证状态
	time.Sleep(300 * time.Millisecond)
	taskObj, _ := s.taskManager.Get(tk.ID)
	if taskObj.Status != "cancelled" {
		t.Errorf("Task should be cancelled, got %s", taskObj.Status)
	}
}

func TestHandleTaskStats(t *testing.T) {
	s := NewServer(8080)
	s.SetupRoutes()

	// 提交几个任务
	for i := 0; i < 5; i++ {
		payload := map[string]interface{}{
			"type":    "echo",
			"payload": map[string]interface{}{},
		}
		body, _ := json.Marshal(payload)
		req := httptest.NewRequest("POST", "/api/v1/tasks", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		s.mux.ServeHTTP(rec, req)
	}

	// 等待执行
	time.Sleep(500 * time.Millisecond)

	req := httptest.NewRequest("GET", "/api/v1/tasks/stats", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Stats should be 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "stats") {
		t.Error("Response should contain stats")
	}
}

func TestHandleAIChatStream(t *testing.T) {
	s := NewServer(8080)
	s.SetupRoutes()

	payload := map[string]interface{}{
		"message":   "hello",
		"use_tools": true,
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/ai/chat/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Stream should be 200, got %d: %s", rec.Code, rec.Body.String())
	}

	// 验证 SSE 格式
	scanner := bufio.NewScanner(bytes.NewReader(rec.Body.Bytes()))
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	chunks := 0
	hasContent := false
	hasDone := false
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "data: ") {
			continue
		}
		chunks++
		var chunk ai.StreamChunk
		if err := json.Unmarshal([]byte(strings.TrimPrefix(line, "data: ")), &chunk); err != nil {
			continue
		}
		if chunk.Type == "content" {
			hasContent = true
		}
		if chunk.Done {
			hasDone = true
		}
	}

	if chunks == 0 {
		t.Error("Should receive at least one chunk")
	}
	if !hasContent {
		t.Error("Should have content chunks")
	}
	if !hasDone {
		t.Error("Should have done chunk")
	}
}

func TestHandleAIChatStreamEmptyMessage(t *testing.T) {
	s := NewServer(8080)
	s.SetupRoutes()

	payload := map[string]interface{}{"message": ""}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/ai/chat/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("Empty message should be 400, got %d", rec.Code)
	}
}

func TestHandleAIStreamCancel(t *testing.T) {
	s := NewServer(8080)
	s.SetupRoutes()

	// 先提交一个长流式请求
	payload := map[string]interface{}{
		"message":   "hello",
		"session_id": "cancel-test-123",
	}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("POST", "/api/v1/ai/chat/stream", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	// 启动流式请求（不同步）
	go func() {
		s.mux.ServeHTTP(rec, req)
	}()

	// 等待 session 注册
	time.Sleep(100 * time.Millisecond)

	// 取消会话
	req2 := httptest.NewRequest("DELETE", "/api/v1/ai/chat/stream/cancel-test-123", nil)
	rec2 := httptest.NewRecorder()
	s.mux.ServeHTTP(rec2, req2)

	if rec2.Code != http.StatusOK {
		t.Errorf("Cancel should be 200, got %d: %s", rec2.Code, rec2.Body.String())
	}

	if !strings.Contains(rec2.Body.String(), "cancelled") {
		t.Errorf("Response should contain cancelled: %s", rec2.Body.String())
	}
}

func TestHandleAIStreamCancelNotFound(t *testing.T) {
	s := NewServer(8080)
	s.SetupRoutes()

	req := httptest.NewRequest("DELETE", "/api/v1/ai/chat/stream/nonexistent", nil)
	rec := httptest.NewRecorder()
	s.mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("Cancel nonexistent should be 404, got %d", rec.Code)
	}
}
