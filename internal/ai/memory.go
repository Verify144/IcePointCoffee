package ai

import (
	"sync"
	"time"
)

// Message 对话消息
type Message struct {
	Role      string    `json:"role"` // system / user / assistant / tool
	Content   string    `json:"content"`
	Name      string    `json:"name,omitempty"`
	ToolCall  string    `json:"tool_call,omitempty"`
	ToolID    string    `json:"tool_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// Memory 对话记忆
type Memory struct {
	mu       sync.RWMutex
	messages []Message
	maxSize  int
}

// NewMemory 创建记忆
func NewMemory(maxSize int) *Memory {
	if maxSize <= 0 {
		maxSize = 50
	}
	return &Memory{
		messages: make([]Message, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add 添加消息
func (m *Memory) Add(msg Message) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if msg.Timestamp.IsZero() {
		msg.Timestamp = time.Now()
	}
	m.messages = append(m.messages, msg)
	// 保留系统消息 + 最近 maxSize 条
	if len(m.messages) > m.maxSize {
		// 找到第一个非系统消息
		cutoff := len(m.messages) - m.maxSize
		newStart := cutoff
		// 如果全是非系统，从头开始截断
		found := false
		for i := 0; i < cutoff; i++ {
			if m.messages[i].Role != "system" {
				newStart = i
				found = true
				break
			}
		}
		if !found {
			// 全部都是系统消息，从头截断
			newStart = cutoff
		}
		// 截断
		newMessages := make([]Message, 0, m.maxSize)
		newMessages = append(newMessages, m.messages[:newStart]...)
		newMessages = append(newMessages, m.messages[cutoff:]...)
		m.messages = newMessages
	}
}

// Get 获取所有消息
func (m *Memory) Get() []Message {
	m.mu.RLock()
	defer m.mu.RUnlock()
	result := make([]Message, len(m.messages))
	copy(result, m.messages)
	return result
}

// Clear 清空
func (m *Memory) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = m.messages[:0]
}

// Size 消息数量
func (m *Memory) Size() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.messages)
}

// LastAssistant 最后一条 assistant 消息
func (m *Memory) LastAssistant() (Message, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	for i := len(m.messages) - 1; i >= 0; i-- {
		if m.messages[i].Role == "assistant" {
			return m.messages[i], true
		}
	}
	return Message{}, false
}
