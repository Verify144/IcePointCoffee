package mc

import (
	"context"
	"fmt"
	"strings"
)

// MockClient 用于测试的 Mock 实现
type MockClient struct {
	Connected    bool
	Commands     []string
	LastCommand  string
	SentMessages []string
	Status_      Status
}

func NewMock(connected bool) *MockClient {
	return &MockClient{
		Connected: connected,
		Commands:  []string{},
		Status_:   Status{Connected: connected, PlayerName: "MockPlayer"},
	}
}

func (m *MockClient) IsConnected() bool          { return m.Connected }
func (m *MockClient) Connect(ctx context.Context) error {
	m.Connected = true
	return nil
}
func (m *MockClient) Close() error {
	m.Connected = false
	return nil
}
func (m *MockClient) Status() Status { return m.Status_ }

func (m *MockClient) SendCommand(ctx context.Context, cmd string) (string, error) {
	m.LastCommand = cmd
	m.Commands = append(m.Commands, cmd)

	// 模拟常见命令输出
	cmd = strings.TrimSpace(cmd)
	if strings.HasPrefix(cmd, "/") {
		cmd = cmd[1:]
	}
	parts := strings.Fields(cmd)
	if len(parts) == 0 {
		return "", nil
	}

	switch parts[0] {
	case "list":
		return "There are 0 players online:", nil
	case "time":
		return "Time set to " + strings.Join(parts[1:], " "), nil
	case "weather":
		return "Weather set to " + strings.Join(parts[1:], " "), nil
	case "gamemode":
		return "Game mode set to " + strings.Join(parts[1:], " "), nil
	case "tp":
		return fmt.Sprintf("Teleported %s to %s,%s,%s", parts[1], parts[2], parts[3], parts[4]), nil
	case "give":
		return fmt.Sprintf("Given %s to %s", parts[2], parts[1]), nil
	case "fill", "setblock":
		return fmt.Sprintf("%d blocks changed", 100), nil
	default:
		return fmt.Sprintf("Command executed: /%s", parts[0]), nil
	}
}

func (m *MockClient) SendChat(msg string) error {
	m.SentMessages = append(m.SentMessages, msg)
	return nil
}
func (m *MockClient) SendTellRaw(target, message string) error {
	m.SentMessages = append(m.SentMessages, fmt.Sprintf("/tellraw %s %s", target, message))
	return nil
}
func (m *MockClient) SendTitle(target, title, subtitle string) error {
	return nil
}
func (m *MockClient) Teleport(target string, x, y, z float64) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/tp %s %.1f %.1f %.1f", target, x, y, z))
	return nil
}
func (m *MockClient) SetGameMode(target, mode string) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/gamemode %s %s", mode, target))
	return nil
}
func (m *MockClient) GiveItem(target, item string, count int) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/give %s %s %d", target, item, count))
	return nil
}
func (m *MockClient) Kill(target string) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/kill %s", target))
	return nil
}
func (m *MockClient) SetBlock(x, y, z int, block string) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/setblock %d %d %d %s", x, y, z, block))
	return nil
}
func (m *MockClient) Fill(x1, y1, z1, x2, y2, z2 int, block string) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/fill %d %d %d %d %d %d %s", x1, y1, z1, x2, y2, z2, block))
	return nil
}
func (m *MockClient) SetTime(time string) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/time set %s", time))
	return nil
}
func (m *MockClient) SetWeather(weather string) error {
	switch weather {
	case "clear", "rain", "thunder":
		_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/weather %s", weather))
		return nil
	default:
		return fmt.Errorf("invalid weather: %s", weather)
	}
}

var _ ClientInterface = (*MockClient)(nil)
