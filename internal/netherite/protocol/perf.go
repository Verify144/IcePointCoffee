package protocol

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

// DefaultHeartbeatInterval 默认心跳间隔 (30秒)
const DefaultHeartbeatInterval = 30 * time.Second

// DefaultReadTimeout 默认读取超时 (60秒)
const DefaultReadTimeout = 60 * time.Second

// DefaultWriteTimeout 默认写入超时 (30秒)
const DefaultWriteTimeout = 30 * time.Second

// DefaultPingInterval 默认 Ping 间隔 (10秒)
const DefaultPingInterval = 10 * time.Second

// PerfConfig 性能配置
type PerfConfig struct {
	HeartbeatInterval time.Duration
	ReadTimeout       time.Duration
	WriteTimeout      time.Duration
	PingInterval      time.Duration
	MaxQueueSize      int
	EnableCompression bool
	CompressionThreshold int
}

// DefaultPerfConfig 默认性能配置
var DefaultPerfConfig = PerfConfig{
	HeartbeatInterval: DefaultHeartbeatInterval,
	ReadTimeout:       DefaultReadTimeout,
	WriteTimeout:      DefaultWriteTimeout,
	PingInterval:      DefaultPingInterval,
	MaxQueueSize:      1000,
	EnableCompression: false,
	CompressionThreshold: 1024,
}

// ConnectionPool 连接池
type ConnectionPool struct {
	mu      sync.Mutex
	conns   map[string]*PooledConn
	config  *PerfConfig
}

// PooledConn 池化连接
type PooledConn struct {
	ID       string
	conn     interface{}
	lastUsed time.Time
	refCount int32
	mu       sync.Mutex
}

// NewConnectionPool 创建连接池
func NewConnectionPool(config *PerfConfig) *ConnectionPool {
	if config == nil {
		config = &DefaultPerfConfig
	}
	return &ConnectionPool{
		conns:  make(map[string]*PooledConn),
		config: config,
	}
}

// Acquire 获取连接
func (p *ConnectionPool) Acquire(id string) *PooledConn {
	p.mu.Lock()
	defer p.mu.Unlock()

	pc, exists := p.conns[id]
	if !exists {
		return nil
	}

	if atomic.LoadInt32(&pc.refCount) > 0 {
		return nil // 已被占用
	}

	atomic.AddInt32(&pc.refCount, 1)
	pc.lastUsed = time.Now()
	return pc
}

// Release 释放连接
func (p *ConnectionPool) Release(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	pc, exists := p.conns[id]
	if exists {
		atomic.AddInt32(&pc.refCount, -1)
	}
}

// Put 保存连接到池
func (p *ConnectionPool) Put(id string, conn interface{}) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.conns[id] = &PooledConn{
		ID:       id,
		conn:     conn,
		lastUsed: time.Now(),
		refCount: 0,
	}
}

// Remove 移除连接
func (p *ConnectionPool) Remove(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	delete(p.conns, id)
}

// Cleanup 清理过期连接
func (p *ConnectionPool) Cleanup(maxAge time.Duration) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now()
	removed := 0
	for id, pc := range p.conns {
		if now.Sub(pc.lastUsed) > maxAge && atomic.LoadInt32(&pc.refCount) == 0 {
			delete(p.conns, id)
			removed++
		}
	}
	return removed
}

// Len 返回连接池大小
func (p *ConnectionPool) Len() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.conns)
}

// ContextWithTimeout 创建带超时的 context
func ContextWithTimeout(parent context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		timeout = DefaultReadTimeout
	}
	return context.WithTimeout(parent, timeout)
}

// ContextWithDeadline 创建带截止时间的 context
func ContextWithDeadline(parent context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	return context.WithDeadline(parent, deadline)
}

// RetryConfig 重试配置
type RetryConfig struct {
	MaxAttempts int
	BaseDelay   time.Duration
	MaxDelay    time.Duration
	Multiplier  float64
}

// DefaultRetryConfig 默认重试配置
var DefaultRetryConfig = RetryConfig{
	MaxAttempts: 3,
	BaseDelay:   100 * time.Millisecond,
	MaxDelay:    5 * time.Second,
	Multiplier:  2.0,
}

// Backoff 计算退避时间
func (c *RetryConfig) Backoff(attempt int) time.Duration {
	delay := float64(c.BaseDelay)
	for i := 0; i < attempt && delay < float64(c.MaxDelay); i++ {
		delay *= c.Multiplier
	}
	if delay > float64(c.MaxDelay) {
		delay = float64(c.MaxDelay)
	}
	return time.Duration(delay)
}

// Retry 执行带重试的操作
func Retry[T any](ctx context.Context, cfg *RetryConfig, fn func() (T, error)) (T, error) {
	if cfg == nil {
		cfg = &DefaultRetryConfig
	}

	var lastErr error
	for attempt := 0; attempt < cfg.MaxAttempts; attempt++ {
		if attempt > 0 {
			delay := cfg.Backoff(attempt - 1)
			select {
			case <-ctx.Done():
				var zero T
				return zero, ctx.Err()
			case <-time.After(delay):
			}
		}

		result, err := fn()
		if err == nil {
			return result, nil
		}
		lastErr = err
	}

	var zero T
	return zero, lastErr
}
