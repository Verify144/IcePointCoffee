package server

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Verify144/IcePointCoffee/internal/metrics"
)

func TestMetricsHandler(t *testing.T) {
	metrics.InitDefault()
	metrics.DefaultBusiness.IncHTTPRequest("/test", "GET", 200)
	metrics.DefaultBusiness.IncAIChat(true, 100)

	s := &Server{}
	handler := s.MetricsHandler()

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/metrics", nil)
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Should be 200, got %d", rec.Code)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "# HELP") {
		t.Error("Should contain HELP lines")
	}
	if !strings.Contains(body, "icepoint_http_requests_total") {
		t.Error("Should contain HTTP requests metric")
	}
	if !strings.Contains(body, "icepoint_ai_chat_total") {
		t.Error("Should contain AI chat metric")
	}
	if !strings.Contains(body, "icepoint_uptime_seconds") {
		t.Error("Should contain uptime metric")
	}
}

func TestMetricsMiddleware(t *testing.T) {
	metrics.InitDefault()
	s := &Server{}

	called := false
	handler := s.MetricsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	}))

	rec := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/test", nil)
	handler.ServeHTTP(rec, req)

	if !called {
		t.Error("Handler should be called")
	}
}

func TestNormalizePath(t *testing.T) {
	tests := []struct {
		input  string
		expect string
	}{
		{"/api/v1/ai/chat", "/api/v1/ai/chat"},
		{"/metrics", "/metrics"},
		{"/health", "/health"},
	}
	for _, tt := range tests {
		got := normalizePath(tt.input)
		if got != tt.expect {
			t.Errorf("normalizePath(%s) = %s, want %s", tt.input, got, tt.expect)
		}
	}
}
