// Package mc 提供 Minecraft Bedrock 连接管理。
// 整合 Raknet 传输层、FB 认证、登录握手和游戏协议。
package mc

import (
	"log"
	"sync"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/netherite/protocol"
)

// EventProcessor 事件处理器，负责接收和分发事件。
type EventProcessor struct {
	client     *Client
	eventBus   *protocol.EventBus
	handlers   map[byte]protocol.EventHandler
	mu         sync.RWMutex
	stopChan   chan struct{}
	running    bool
}

// NewEventProcessor 创建事件处理器。
func NewEventProcessor(client *Client) *EventProcessor {
	ep := &EventProcessor{
		client:   client,
		handlers: make(map[byte]protocol.EventHandler),
		stopChan: make(chan struct{}),
	}
	ep.eventBus = protocol.NewEventBus()
	return ep
}

// RegisterHandler 注册事件处理器。
func (ep *EventProcessor) RegisterHandler(packetID byte, handler protocol.EventHandler) func() {
	ep.mu.Lock()
	ep.handlers[packetID] = handler
	ep.mu.Unlock()

	ep.eventBus.Register(packetID, func(data []byte) error {
		ep.mu.RLock()
		h, ok := ep.handlers[packetID]
		ep.mu.RUnlock()
		if !ok {
			return nil
		}
		return h(data)
	})

	return func() {
		ep.mu.Lock()
		defer ep.mu.Unlock()
		delete(ep.handlers, packetID)
	}
}

// Start 启动事件处理器。
func (ep *EventProcessor) Start() {
	ep.mu.Lock()
	if ep.running {
		ep.mu.Unlock()
		return
	}
	ep.running = true
	ep.mu.Unlock()

	go ep.eventLoop()
	log.Printf("[MC] EventProcessor started")
}

// Stop 停止事件处理器。
func (ep *EventProcessor) Stop() {
	ep.mu.Lock()
	if !ep.running {
		ep.mu.Unlock()
		return
	}
	ep.running = false
	ep.mu.Unlock()

	close(ep.stopChan)
	log.Printf("[MC] EventProcessor stopped")
}

// eventLoop 事件循环，从连接读取包并分发。
func (ep *EventProcessor) eventLoop() {
	for {
		select {
		case <-ep.stopChan:
			return
		default:
		}

		frame, err := ep.client.conn.Recv()
		if err != nil {
			select {
			case <-ep.stopChan:
				return
			case <-time.After(100 * time.Millisecond):
				continue
			}
		}
		if frame == nil || len(frame.Data) == 0 {
			continue
		}

		packetID := frame.Data[0]
		ep.eventBus.Dispatch(packetID, frame.Data)
	}
}

// Dispatch 直接分发事件（用于测试）。
func (ep *EventProcessor) Dispatch(packetID byte, data []byte) {
	ep.eventBus.Dispatch(packetID, data)
}

// HasHandler 检查是否有对应 packetID 的处理器。
func (ep *EventProcessor) HasHandler(packetID byte) bool {
	return ep.eventBus.HasHandler(packetID)
}

// EventBus 获取底层事件总线。
func (ep *EventProcessor) EventBus() *protocol.EventBus {
	return ep.eventBus
}

// ConnectHandler 连接事件处理器。
type ConnectHandler struct {
	*EventProcessor
	OnConnect    func(*Client)
	OnDisconnect func(*Client, error)
	OnError      func(*Client, error)
}

// NewConnectHandler 创建连接事件处理器。
func NewConnectHandler(client *Client) *ConnectHandler {
	return &ConnectHandler{
		EventProcessor: NewEventProcessor(client),
	}
}

// HandleConnect 处理连接事件。
func (ch *ConnectHandler) HandleConnect(c *Client) {
	if ch.OnConnect != nil {
		ch.OnConnect(c)
	}
}

// HandleDisconnect 处理断开事件。
func (ch *ConnectHandler) HandleDisconnect(c *Client, err error) {
	if ch.OnDisconnect != nil {
		ch.OnDisconnect(c, err)
	}
}

// HandleError 处理错误事件。
func (ch *ConnectHandler) HandleError(c *Client, err error) {
	if ch.OnError != nil {
		ch.OnError(c, err)
	}
}
