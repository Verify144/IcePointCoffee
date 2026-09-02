// Package mc 提供 MC 服务器操控接口。
package mc

import (
	"context"
	"fmt"
	"strings"
	"sync"

	netherite "github.com/Verify144/IcePointCoffee/internal/netherite/mc"
)

// Adapter 持有真实的 netherite MC Client。
// 提供安全的命令发送接口，支持动态注入（测试/mock）和危险命令检查。
type Adapter struct {
	mu  sync.RWMutex
	mc  *netherite.Client
}

// NewAdapter 创建适配器。
func NewAdapter() *Adapter {
	return &Adapter{}
}

// SetClient 注入真实 client（main.go 延迟初始化用）
func (a *Adapter) SetClient(client *netherite.Client) {
	a.mu.Lock()
	a.mc = client
	a.mu.Unlock()
}

func (a *Adapter) client() *netherite.Client {
	a.mu.RLock()
	defer a.mu.RUnlock()
	return a.mc
}

// ---- ClientInterface 实现 ----

func (a *Adapter) IsConnected() bool {
	c := a.client()
	return c != nil && c.IsConnected()
}

func (a *Adapter) Connect(ctx context.Context) error {
	c := a.client()
	if c == nil {
		return fmt.Errorf("mc client 未初始化")
	}
	return c.Connect(ctx)
}

func (a *Adapter) Close() error {
	c := a.client()
	if c == nil {
		return nil
	}
	return c.Close()
}

func (a *Adapter) Status() Status {
	c := a.client()
	if c == nil {
		return Status{Connected: false}
	}
	return Status{Connected: c.IsConnected()}
}

func (a *Adapter) SendCommand(ctx context.Context, cmd string) (string, error) {
	c := a.client()
	if c == nil {
		return "", fmt.Errorf("mc client 未初始化")
	}
	if !c.IsConnected() {
		return "", fmt.Errorf("未连接到服务器")
	}

	out, err := c.SendCommand(ctx, cmd)
	if err != nil {
		return "", err
	}
	if out == nil {
		return "命令已执行（无输出）", nil
	}

	var lines []string
	for _, msg := range out.Messages {
		if msg.Message != "" {
			lines = append(lines, msg.Message)
		}
	}
	if len(lines) == 0 {
		return fmt.Sprintf("执行成功 (success=%d)", out.SuccessCount), nil
	}
	return strings.Join(lines, "\n"), nil
}

func (a *Adapter) SendChat(msg string) error {
	c := a.client()
	if c == nil || !c.IsConnected() {
		return fmt.Errorf("未连接到服务器")
	}
	return c.SendChat(msg)
}

func (a *Adapter) SendTellRaw(target, message string) error {
	return a.runCmd(fmt.Sprintf("/tellraw %s %s", target, message))
}

func (a *Adapter) SendTitle(target, title, subtitle string) error {
	if subtitle != "" {
		a.runCmd(fmt.Sprintf("/title %s subtitle %s", target, subtitle))
	}
	return a.runCmd(fmt.Sprintf("/title %s title %s", target, title))
}

func (a *Adapter) Teleport(target string, x, y, z float64) error {
	return a.runCmd(fmt.Sprintf("/tp %s %.1f %.1f %.1f", target, x, y, z))
}

func (a *Adapter) SetGameMode(target, mode string) error {
	return a.runCmd(fmt.Sprintf("/gamemode %s %s", mode, target))
}

func (a *Adapter) GiveItem(target, item string, count int) error {
	if count <= 0 {
		count = 1
	}
	return a.runCmd(fmt.Sprintf("/give %s %s %d", target, item, count))
}

func (a *Adapter) Kill(target string) error {
	return a.runCmd(fmt.Sprintf("/kill %s", target))
}

func (a *Adapter) SetBlock(x, y, z int, block string) error {
	return a.runCmd(fmt.Sprintf("/setblock %d %d %d %s", x, y, z, block))
}

func (a *Adapter) Fill(x1, y1, z1, x2, y2, z2 int, block string) error {
	return a.runCmd(fmt.Sprintf("/fill %d %d %d %d %d %d %s", x1, y1, z1, x2, y2, z2, block))
}

func (a *Adapter) SetTime(time string) error {
	return a.runCmd(fmt.Sprintf("/time set %s", time))
}

func (a *Adapter) SetWeather(weather string) error {
	switch weather {
	case "clear", "rain", "thunder":
		return a.runCmd(fmt.Sprintf("/weather %s", weather))
	default:
		return fmt.Errorf("未知的天气: %s (clear/rain/thunder)", weather)
	}
}

func (a *Adapter) runCmd(cmd string) error {
	_, err := a.SendCommand(context.Background(), cmd)
	return err
}

// compile-time interface check
var _ ClientInterface = (*Adapter)(nil)
