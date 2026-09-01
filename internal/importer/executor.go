// Package importer 负责将建筑指令实际发送到租赁服。
package importer

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Verify144/IcePointCoffee/internal/netherite/mc"
)

// Executor 指令执行器。
type Executor struct {
	client *mc.Client
}

// NewExecutor 创建指令执行器。
func NewExecutor(client *mc.Client) *Executor {
	return &Executor{client: client}
}

// Execute 执行单条指令。
func (e *Executor) Execute(ctx context.Context, command string) error {
	if e.client == nil {
		return fmt.Errorf("mc client 未初始化")
	}
	if !e.client.IsConnected() {
		return fmt.Errorf("未连接到服务器")
	}
	if _, err := e.client.SendCommand(ctx, command); err != nil {
		return err
	}
	time.Sleep(50 * time.Millisecond)
	return nil
}

// ExecuteBatch 批量执行指令。
func (e *Executor) ExecuteBatch(ctx context.Context, commands []string) error {
	for i, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" || strings.HasPrefix(cmd, "#") {
			continue
		}
		if err := e.Execute(ctx, cmd); err != nil {
			return fmt.Errorf("执行指令 [%d] %s 失败: %w", i+1, cmd, err)
		}
	}
	return nil
}
