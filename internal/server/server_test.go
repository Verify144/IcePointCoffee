package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	rl := NewRateLimiter(3, 100*time.Millisecond)
	key := "127.0.0.1:12345"

	// 前 3 个应该通过
	for i := 0; i < 3; i++ {
		if !rl.Allow(key) {
			t.Errorf("Request %d should be allowed", i+1)
		}
	}

	// 第 4 个应该被拒
	if rl.Allow(key) {
		t.Error("4th request should be rate limited")
	}

	// 等待窗口过期
	time.Sleep(150 * time.Millisecond)

	// 应该再次通过
	if !rl.Allow(key) {
		t.Error("After window, request should be allowed")
	}
}

func TestRateLimiterMiddleware(t *testing.T) {
	rl := NewRateLimiter(1, 100*time.Millisecond)
	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	}))

	// 第一个应该通过
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("First request should be 200, got %d", rec.Code)
	}

	// 第二个应该被限流
	req2 := httptest.NewRequest("GET", "/", nil)
	rec2 := httptest.NewRecorder()
	handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Errorf("Second request should be 429, got %d", rec2.Code)
	}
}

func TestAPIDoc(t *testing.T) {
	doc := GetAPIDoc()
	if doc.Version == "" {
		t.Error("Version should not be empty")
	}
	if len(doc.Endpoints) == 0 {
		t.Error("Should have endpoints")
	}
}

func TestNewError(t *testing.T) {
	e := NewError(400, "Bad Request", "Check your input")
	if e.Code != 400 {
		t.Error("Code mismatch")
	}
	if e.Message != "Bad Request" {
		t.Error("Message mismatch")
	}
	if e.Hint == "" {
		t.Error("Hint should not be empty")
	}
}

func TestSendError(t *testing.T) {
	rec := httptest.NewRecorder()
	SendError(rec, http.StatusBadRequest, 400, "test", "hint")
	if rec.Code != http.StatusBadRequest {
		t.Errorf("Status mismatch: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "test") {
		t.Error("Body should contain message")
	}
	if !strings.Contains(rec.Body.String(), "hint") {
		t.Error("Body should contain hint")
	}
}

func TestSendSuccess(t *testing.T) {
	rec := httptest.NewRecorder()
	SendSuccess(rec, map[string]string{"hello": "world"})
	if rec.Code != http.StatusOK {
		t.Errorf("Status should be 200, got %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "hello") {
		t.Error("Body should contain data")
	}
}

func TestServerNew(t *testing.T) {
	s := NewServer(8080)
	if s.port != 8080 {
		t.Errorf("Port mismatch: %d", s.port)
	}
	if s.registry == nil {
		t.Error("Registry should not be nil")
	}
	if s.builder == nil {
		t.Error("Builder should not be nil")
	}
	if s.rateLimit == nil {
		t.Error("RateLimit should not be nil")
	}
}

func TestEstimateBlockCount(t *testing.T) {
	tests := []struct {
		input    string
		expected int
	}{
		{"✅ 生成 1024 个块（5 种）\nstone: 500", 1024},
		{"✅ 生成 5000 个块（10 种）", 5000},
		{"no emoji line", 100},
		{"", 100},
	}
	for _, tt := range tests {
		got := estimateBlockCount(tt.input)
		if got != tt.expected {
			t.Errorf("estimateBlockCount(%q) = %d, want %d", tt.input, got, tt.expected)
		}
	}
}
