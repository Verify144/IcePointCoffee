// Package task 提供任务调度、持久化和执行管理。
package task

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// Status 任务状态
type Status string

const (
	StatusPending   Status = "pending"   // 等待执行
	StatusRunning   Status = "running"   // 执行中
	StatusSuccess   Status = "success"   // 成功
	StatusFailed    Status = "failed"    // 失败
	StatusRetrying  Status = "retrying"  // 重试中
	StatusCancelled Status = "cancelled" // 已取消
)

// Priority 任务优先级
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 2
	PriorityHigh   Priority = 3
)

// Task 任务定义
type Task struct {
	ID          string                 `json:"id"`
	Type        string                 `json:"type"`         // 任务类型（build/command/ai_chat...）
	Payload     map[string]interface{} `json:"payload"`      // 任务参数
	Status      Status                 `json:"status"`
	Priority    Priority               `json:"priority"`
	Progress    int                    `json:"progress"`     // 0-100
	Result      string                 `json:"result"`
	Error       string                 `json:"error,omitempty"`
	Retries     int                    `json:"retries"`
	MaxRetries  int                    `json:"max_retries"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	UserID      string                 `json:"user_id,omitempty"`
	Tags        []string               `json:"tags,omitempty"`
}

// Handler 任务处理器
type Handler func(ctx context.Context, t *Task) error

// Store 任务存储接口
type Store interface {
	Create(t *Task) error
	Update(t *Task) error
	Get(id string) (*Task, error)
	List(filter ListFilter) ([]*Task, error)
	Delete(id string) error
	Count() (int, error)
}

// ListStore 内存中的简单任务存储（不依赖 SQLite）
type ListStore struct {
	mu     sync.RWMutex
	tasks  map[string]*Task
}

// NewListStore 创建内存任务存储
func NewListStore() *ListStore {
	return &ListStore{tasks: make(map[string]*Task)}
}

// Create 创建任务
func (s *ListStore) Create(t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.tasks[t.ID]; exists {
		return fmt.Errorf("task %s already exists", t.ID)
	}
	s.tasks[t.ID] = t
	return nil
}

// Update 更新任务
func (s *ListStore) Update(t *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tasks[t.ID] = t
	return nil
}

// Get 获取任务
func (s *ListStore) Get(id string) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task %s not found", id)
	}
	return t, nil
}

// ListFilter 列表过滤
type ListFilter struct {
	Status string
	Type   string
	UserID string
	Limit  int
}

// List 列出任务
func (s *ListStore) List(filter ListFilter) ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, t := range s.tasks {
		if filter.Status != "" && string(t.Status) != filter.Status {
			continue
		}
		if filter.Type != "" && t.Type != filter.Type {
			continue
		}
		if filter.UserID != "" && t.UserID != filter.UserID {
			continue
		}
		tasks = append(tasks, t)
	}

	// 按创建时间倒序
	for i := 0; i < len(tasks); i++ {
		for j := i + 1; j < len(tasks); j++ {
			if tasks[j].CreatedAt.After(tasks[i].CreatedAt) {
				tasks[i], tasks[j] = tasks[j], tasks[i]
			}
		}
	}

	if filter.Limit > 0 && len(tasks) > filter.Limit {
		tasks = tasks[:filter.Limit]
	}
	return tasks, nil
}

// Delete 删除任务
func (s *ListStore) Delete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.tasks, id)
	return nil
}

// Count 任务总数
func (s *ListStore) Count() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.tasks), nil
}

// GenerateID 生成任务 ID
func GenerateID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return "task_" + hex.EncodeToString(b)
}

// NewTask 创建新任务
func NewTask(taskType string, payload map[string]interface{}) *Task {
	return &Task{
		ID:         GenerateID(),
		Type:       taskType,
		Payload:    payload,
		Status:     StatusPending,
		Priority:   PriorityNormal,
		Progress:   0,
		MaxRetries: 3,
		CreatedAt:  time.Now(),
	}
}

// ErrCancelled 任务被取消
var ErrCancelled = errors.New("task cancelled")

// Manager 任务管理器
type Manager struct {
	store    Store
	handlers map[string]Handler
	mu       sync.RWMutex

	// worker pool
	workers     int
	workChannel chan *Task
	stopChannel chan struct{}
	running     int32
	stopped     bool

	// 正在执行的任务的 cancel 映射，用于主动取消
	cancels      map[string]context.CancelFunc
	cancelledIDs map[string]struct{} // 主动取消的任务 ID，重试时不执行
	cancelsMu    sync.Mutex
}

// NewManager 创建任务管理器
func NewManager(store Store, workers int) *Manager {
	if workers <= 0 {
		workers = 4
	}
	m := &Manager{
		store:        store,
		handlers:     make(map[string]Handler),
		workers:      workers,
		workChannel:  make(chan *Task, 100),
		stopChannel:  make(chan struct{}),
		cancels:      make(map[string]context.CancelFunc),
		cancelledIDs: make(map[string]struct{}),
	}
	return m
}

// Register 注册任务处理器
func (m *Manager) Register(taskType string, h Handler) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.handlers[taskType] = h
}

// Start 启动 worker pool
func (m *Manager) Start() {
	for i := 0; i < m.workers; i++ {
		go m.worker()
	}
}

// Stop 停止管理器
func (m *Manager) Stop() {
	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return
	}
	m.stopped = true
	close(m.stopChannel)
	m.mu.Unlock()
}

// Submit 提交任务
func (m *Manager) Submit(t *Task) error {
	if t.ID == "" {
		t.ID = GenerateID()
	}
	if t.Status == "" {
		t.Status = StatusPending
	}
	if t.CreatedAt.IsZero() {
		t.CreatedAt = time.Now()
	}
	if err := m.store.Create(t); err != nil {
		return err
	}

	select {
	case m.workChannel <- t:
		return nil
	case <-m.stopChannel:
		return errors.New("manager stopped")
	}
}

// Cancel 取消任务
func (m *Manager) Cancel(id string) error {
	t, err := m.store.Get(id)
	if err != nil {
		return err
	}

	// 标记为已取消（用于阻止后续重试）
	m.cancelsMu.Lock()
	m.cancelledIDs[t.ID] = struct{}{}
	if cancel, ok := m.cancels[t.ID]; ok {
		cancel() // 中断正在执行的 handler
		delete(m.cancels, t.ID)
	}
	m.cancelsMu.Unlock()

	if t.Status == StatusPending {
		t.Status = StatusCancelled
		return m.store.Update(t)
	}
	// 如果是 Running 状态，handler 会收到 cancel 信号并返回
	return nil
}

// worker 工作协程
func (m *Manager) worker() {
	for {
		select {
		case <-m.stopChannel:
			return
		case t := <-m.workChannel:
			atomic.AddInt32(&m.running, 1)
			m.executeTask(t)
			atomic.AddInt32(&m.running, -1)
		}
	}
}

// executeTask 执行单个任务
func (m *Manager) executeTask(t *Task) {
	m.mu.RLock()
	handler, ok := m.handlers[t.Type]
	m.mu.RUnlock()

	if !ok {
		t.Status = StatusFailed
		t.Error = "no handler for type: " + t.Type
		now := time.Now()
		t.CompletedAt = &now
		m.store.Update(t)
		return
	}

	// 标记运行
	t.Status = StatusRunning
	now := time.Now()
	t.StartedAt = &now
	m.store.Update(t)

	// 执行（带重试）
	var err error
	for attempt := 0; attempt <= t.MaxRetries; attempt++ {
		// 重试前检查是否已被主动取消
		m.cancelsMu.Lock()
		if _, cancelled := m.cancelledIDs[t.ID]; cancelled {
			delete(m.cancelledIDs, t.ID)
			m.cancelsMu.Unlock()
			t.Status = StatusCancelled
			t.Error = "task cancelled"
			now := time.Now()
			t.CompletedAt = &now
			m.store.Update(t)
			return
		}
		m.cancelsMu.Unlock()

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		// 注册 cancel 以支持外部主动取消
		m.cancelsMu.Lock()
		m.cancels[t.ID] = cancel
		m.cancelsMu.Unlock()

		err = handler(ctx, t)

		// 清理 cancel
		m.cancelsMu.Lock()
		delete(m.cancels, t.ID)
		m.cancelsMu.Unlock()
		cancel()

		if err == nil {
			break
		}
		if err == ErrCancelled {
			t.Status = StatusCancelled
			t.Error = err.Error()
			now := time.Now()
			t.CompletedAt = &now
			m.store.Update(t)
			return
		}
		if attempt < t.MaxRetries {
			t.Status = StatusRetrying
			t.Retries = attempt + 1
			t.Error = err.Error()
			m.store.Update(t)
			// 指数退避：1s, 2s, 4s, 8s...
			delay := time.Duration(1<<attempt) * time.Second
			select {
			case <-time.After(delay):
			case <-m.stopChannel:
				return
			}
		}
	}

	if err != nil {
		t.Status = StatusFailed
		t.Error = err.Error()
		now := time.Now()
		t.CompletedAt = &now
	} else {
		t.Status = StatusSuccess
		t.Progress = 100
		now := time.Now()
		t.CompletedAt = &now
	}
	m.store.Update(t)
}

// Running 返回正在运行的任务数
func (m *Manager) Running() int32 {
	return atomic.LoadInt32(&m.running)
}

// Get 获取任务
func (m *Manager) Get(id string) (*Task, error) {
	return m.store.Get(id)
}

// List 列出任务
func (m *Manager) List(filter ListFilter) ([]*Task, error) {
	return m.store.List(filter)
}

// UpdateProgress 更新任务进度
func (m *Manager) UpdateProgress(id string, progress int, result string) error {
	t, err := m.store.Get(id)
	if err != nil {
		return err
	}
	t.Progress = progress
	if result != "" {
		t.Result = result
	}
	return m.store.Update(t)
}

// Stats 返回任务统计（委托给 store）
func (m *Manager) Stats() (map[string]int, error) {
	type statsStore interface {
		Stats() (map[string]int, error)
	}
	if s, ok := m.store.(statsStore); ok {
		return s.Stats()
	}
	// 回退到内存统计
	tasks, _ := m.List(ListFilter{Limit: 1000})
	stats := map[string]int{"total": len(tasks)}
	for _, t := range tasks {
		stats[string(t.Status)]++
	}
	stats["running"] = int(m.Running())
	return stats, nil
}
