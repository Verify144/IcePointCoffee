// Package protocol 实现 Minecraft Bedrock 协议。
package protocol

import (
	"fmt"
	"sync"
)

// EventHandler 处理事件的函数。
type EventHandler func(data []byte) error

// EventBus 事件总线，按 packet ID 路由。
type EventBus struct {
	mu       sync.RWMutex
	handlers map[byte][]EventHandler
	once     map[byte][]EventHandler
}

// NewEventBus 创建事件总线。
func NewEventBus() *EventBus {
	return &EventBus{
		handlers: make(map[byte][]EventHandler),
		once:     make(map[byte][]EventHandler),
	}
}

// Register 注册事件处理器，返回取消函数。
func (eb *EventBus) Register(packetID byte, handler EventHandler) func() {
	eb.mu.Lock()
	eb.handlers[packetID] = append(eb.handlers[packetID], handler)
	eb.mu.Unlock()
	return func() {
		eb.mu.Lock()
		defer eb.mu.Unlock()
		list := eb.handlers[packetID]
		for i, h := range list {
			if fmt.Sprintf("%p", h) == fmt.Sprintf("%p", handler) {
				eb.handlers[packetID] = append(list[:i], list[i+1:]...)
				break
			}
		}
	}
}

// RegisterOnce 注册一次性事件处理器。
func (eb *EventBus) RegisterOnce(packetID byte, handler EventHandler) {
	eb.mu.Lock()
	eb.once[packetID] = append(eb.once[packetID], handler)
	eb.mu.Unlock()
}

// Dispatch 分发事件。
func (eb *EventBus) Dispatch(packetID byte, data []byte) {
	eb.mu.RLock()
	handlers := make([]EventHandler, len(eb.handlers[packetID]))
	copy(handlers, eb.handlers[packetID])
	onceHandlers := make([]EventHandler, len(eb.once[packetID]))
	if len(eb.once[packetID]) > 0 {
		copy(onceHandlers, eb.once[packetID])
		clear(eb.once[packetID])
	}
	eb.mu.RUnlock()
	for _, h := range handlers {
		if err := h(data); err != nil {
			// 忽略错误
		}
	}
	for _, h := range onceHandlers {
		if err := h(data); err != nil {
			// 忽略错误
		}
	}
}

// HasHandler 返回是否有对应 packetID 的处理器。
func (eb *EventBus) HasHandler(packetID byte) bool {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.handlers[packetID]) > 0 || len(eb.once[packetID]) > 0
}

// Listeners 返回监听器数量。
func (eb *EventBus) Listeners(packetID byte) int {
	eb.mu.RLock()
	defer eb.mu.RUnlock()
	return len(eb.handlers[packetID])
}

// Clear 清除所有处理器。
func (eb *EventBus) Clear(packetID byte) {
	eb.mu.Lock()
	defer eb.mu.Unlock()
	eb.handlers[packetID] = nil
	eb.once[packetID] = nil
}
