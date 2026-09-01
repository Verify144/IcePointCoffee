package server

import (
	"net/http"
	"strings"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/metrics"
)

// MetricsHandler handles the /metrics endpoint.
func (s *Server) MetricsHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		metrics.DefaultRegistry.WriteTo(w)
	})
}

// MetricsMiddleware records HTTP request metrics.
func (s *Server) MetricsMiddleware(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		handler.ServeHTTP(rec, r)
		path := normalizePath(r.URL.Path)
		metrics.DefaultBusiness.IncHTTPRequest(path, r.Method, rec.status)
		metrics.DefaultBusiness.ObserveHTTPLatency(path, time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func normalizePath(path string) string {
	if path == "/metrics" || path == "/health" {
		return path
	}
	if len(path) > 100 {
		return "/other"
	}
	return strings.ToLower(path)
}
