package server

import (
	"net/http"
	"sync"
	"time"
)

// RateLimiter 限流器
type RateLimiter struct {
	mu        sync.Mutex
	requests  map[string][]time.Time
	max       int
	window    time.Duration
	cleanup   time.Duration
	lastClean time.Time
}

// NewRateLimiter 创建限流器
// max: 窗口内最大请求数
// window: 时间窗口
func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests:  make(map[string][]time.Time),
		max:       max,
		window:    window,
		cleanup:   window,
		lastClean: time.Now(),
	}
}

// Allow 检查是否允许请求
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// 清理过期
	if now.Sub(r.lastClean) > r.cleanup {
		for k, times := range r.requests {
			filtered := times[:0]
			for _, t := range times {
				if t.After(cutoff) {
					filtered = append(filtered, t)
				}
			}
			if len(filtered) == 0 {
				delete(r.requests, k)
			} else {
				r.requests[k] = filtered
			}
		}
		r.lastClean = now
	}

	// 过滤窗口内请求
	times := r.requests[key]
	validTimes := times[:0]
	for _, t := range times {
		if t.After(cutoff) {
			validTimes = append(validTimes, t)
		}
	}

	if len(validTimes) >= r.max {
		r.requests[key] = validTimes
		return false
	}

	validTimes = append(validTimes, now)
	r.requests[key] = validTimes
	return true
}

// Middleware 限流中间件
func (r *RateLimiter) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		key := req.RemoteAddr
		if !r.Allow(key) {
			SendError(w, http.StatusTooManyRequests, 429, "请求过快", "请稍后再试")
			return
		}
		next.ServeHTTP(w, req)
	})
}

// Size 返回当前追踪的 key 数
func (r *RateLimiter) Size() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.requests)
}
