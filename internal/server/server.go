// Package server 负责租赁服连接。
// 冰点咖啡通过 FB token 连接到网易租赁服。
// 注意：实际 Raknet 连接在生产环境需要租用方提供，这里只做指令队列管理。
package server

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Client 租赁服客户端。
type Client struct {
	address  string
	playerName string
	fbToken  string
	mu       sync.Mutex
	connected bool
	commandQueue []string
}

// NewClient 创建租赁服客户端。
func NewClient(address, playerName, fbToken string) *Client {
	return &Client{
		address:   address,
		playerName: playerName,
		fbToken:   fbToken,
		commandQueue: make([]string, 0),
	}
}

// Connect 连接到租赁服。
func (c *Client) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.fbToken == "" {
		return fmt.Errorf("fb_token 未设置")
	}
	if c.address == "" {
		return fmt.Errorf("服务器地址未设置")
	}

	// 这里使用 WebSocket / Raknet 与服务器通信
	// 实际实现可以基于 gophertunnel 协议栈
	// 为了保持冰点咖啡轻量，这里只标记已连接并允许指令提交
	c.connected = true
	return nil
}

// Disconnect 断开连接。
func (c *Client) Disconnect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.connected = false
	c.commandQueue = c.commandQueue[:0]
	return nil
}

// IsConnected 返回是否已连接。
func (c *Client) IsConnected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.connected
}

// SubmitCommand 提交一条指令到租赁服。
// 实际实现：使用 Raknet/WS 协议将指令发送到游戏。
func (c *Client) SubmitCommand(ctx context.Context, command string) error {
	c.mu.Lock()
	if !c.connected {
		c.mu.Unlock()
		return fmt.Errorf("未连接到服务器")
	}
	c.commandQueue = append(c.commandQueue, command)
	c.mu.Unlock()

	// TODO: 实际发送逻辑
	// 这里给上层一个异步处理钩子
	time.Sleep(50 * time.Millisecond) // 模拟网络延迟
	return nil
}

// SubmitCommands 批量提交指令。
func (c *Client) SubmitCommands(ctx context.Context, commands []string) error {
	for _, cmd := range commands {
		if err := c.SubmitCommand(ctx, cmd); err != nil {
			return err
		}
	}
	return nil
}

// Address 返回服务器地址。
func (c *Client) Address() string {
	return c.address
}

// PlayerName 返回玩家名称。
func (c *Client) PlayerName() string {
	return c.playerName
}
