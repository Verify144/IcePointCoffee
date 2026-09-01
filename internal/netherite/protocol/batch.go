package protocol

import (
	"bytes"
	"sync"
	"time"
)

// Event 通用事件类型
type Event struct {
	Type string
	Data interface{}
}

// CommandBatch 命令批处理器
type CommandBatch struct {
	mu       sync.Mutex
	commands []string
	conn     CommandSender
	maxSize  int
	interval time.Duration
	done     chan struct{}
}

// CommandSender 发送命令接口
type CommandSender interface {
	SendCommand(cmd string) error
}

// NewCommandBatch 创建命令批处理器
func NewCommandBatch(conn CommandSender, maxSize int, interval time.Duration) *CommandBatch {
	return &CommandBatch{
		commands: make([]string, 0, maxSize),
		conn:     conn,
		maxSize:  maxSize,
		interval: interval,
		done:     make(chan struct{}),
	}
}

// Add 添加命令到批次
func (b *CommandBatch) Add(cmd string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.commands = append(b.commands, cmd)
}

// Flush 立即发送所有命令
func (b *CommandBatch) Flush() error {
	b.mu.Lock()
	if len(b.commands) == 0 {
		b.mu.Unlock()
		return nil
	}
	cmds := b.commands
	b.commands = make([]string, 0, b.maxSize)
	b.mu.Unlock()

	// 合并所有命令用 \n 分隔
	fullCmd := bytes.Buffer{}
	for i, cmd := range cmds {
		if i > 0 {
			fullCmd.WriteByte('\n')
		}
		fullCmd.WriteString(cmd)
	}
	return b.conn.SendCommand(fullCmd.String())
}

// StartAutoFlush 启动自动刷新
func (b *CommandBatch) StartAutoFlush() {
	ticker := time.NewTicker(b.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				b.Flush()
			case <-b.done:
				ticker.Stop()
				b.Flush()
				return
			}
		}
	}()
}

// Stop 停止批处理器
func (b *CommandBatch) Stop() {
	close(b.done)
}

// Pending 返回待发送命令数
func (b *CommandBatch) Pending() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.commands)
}

// EventBuffer 事件缓冲
type EventBuffer struct {
	mu       sync.Mutex
	events   []Event
	maxSize  int
	callback func([]Event)
}

// NewEventBuffer 创建事件缓冲
func NewEventBuffer(maxSize int, callback func([]Event)) *EventBuffer {
	return &EventBuffer{
		events:   make([]Event, 0, maxSize),
		maxSize:  maxSize,
		callback: callback,
	}
}

// Push 添加事件
func (b *EventBuffer) Push(event Event) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.events = append(b.events, event)
	if len(b.events) >= b.maxSize {
		events := b.events
		b.events = make([]Event, 0, b.maxSize)
		b.callback(events) // 同步调用，避免竞态
	}
}

// Flush 强制刷新
func (b *EventBuffer) Flush() {
	b.mu.Lock()
	if len(b.events) == 0 {
		b.mu.Unlock()
		return
	}
	events := b.events
	b.events = make([]Event, 0, b.maxSize)
	b.mu.Unlock()
	b.callback(events)
}


