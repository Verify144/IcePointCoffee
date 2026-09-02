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

// ---- 新协议包方法 (Mock) ----

func (m *MockClient) Respawn(x, y, z float64) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/spawnpoint @s %.1f %.1f %.1f", x, y, z))
	return nil
}
func (m *MockClient) SwingArm() error {
	return nil
}
func (m *MockClient) AttackEntity(targetID uint64) error {
	return nil
}
func (m *MockClient) SpawnEntity(entityType, name string, x, y, z float64) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/summon %s %.1f %.1f %.1f", entityType, x, y, z))
	return nil
}
func (m *MockClient) RemoveEntity(entityID uint64) error {
	_, _ = m.SendCommand(context.Background(), "/kill @e[type=!player]")
	return nil
}
func (m *MockClient) PlaySound(soundID string, x, y, z float64, volume, pitch float32) error {
	return nil
}
func (m *MockClient) StopSound(soundID string) error {
	return nil
}
func (m *MockClient) EmitParticle(particleID int32, x, y, z float64) error {
	return nil
}
func (m *MockClient) SetBossBar(target string, title string, healthPercent float32) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/bossbar add %s \"%s\"", target, title))
	return nil
}
func (m *MockClient) SetDifficulty(difficulty string) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/difficulty %s", difficulty))
	return nil
}
func (m *MockClient) SetWorldBorder(centerX, centerZ, radius float64) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/worldborder center %.1f %.1f", centerX, centerZ))
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/worldborder set %.1f", radius))
	return nil
}
func (m *MockClient) SetSpawnpoint(x, y, z float64) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/setworldspawn %.1f %.1f %.1f", x, y, z))
	return nil
}
func (m *MockClient) SetNametag(target, nametag string) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/entitydata @e[type=%s] {CustomName:'\"%s\"'}", target, nametag))
	return nil
}
func (m *MockClient) GiveXP(target string, amount int) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/xp %d %s", amount, target))
	return nil
}
func (m *MockClient) ApplyEffect(target, effect string, seconds int) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/effect %s %s %d", target, effect, seconds))
	return nil
}
func (m *MockClient) ClearInventory(target string) error {
	_, _ = m.SendCommand(context.Background(), fmt.Sprintf("/clear %s", target))
	return nil
}
func (m *MockClient) Perceive(radius int, detail string) (*PerceiveResult, error) {
	if !m.Connected {
		return nil, fmt.Errorf("not connected")
	}
	return &PerceiveResult{
		Success: true,
		Player: Player{
			Name:     "MockPlayer",
			X:        0, Y: 64, Z: 0,
			Health:   20,
			MaxHealth: 20,
			Food:     20,
			GameMode: "survival",
			World:    "overworld",
		},
		NearbyPlayers: []NearbyPlayer{
			{Name: "Steve", X: 5, Y: 64, Z: 3, Dist: 5.8},
			{Name: "Alex", X: -4, Y: 65, Z: 7, Dist: 8.1},
		},
		NearbyEntities: []NearbyEntity{
			{Type: "pig", X: 8, Y: 64, Z: -2, Dist: 8.2},
			{Type: "cow", X: -6, Y: 64, Z: 1, Dist: 6.1},
		},
		NearbyBlocks: []NearbyBlock{
			{ID: "grass_block", X: 0, Y: 63, Z: 0, Dist: 1},
			{ID: "dirt", X: 1, Y: 63, Z: 0, Dist: 1},
		},
		Description: "你在主城广场中心，坐标 (0.0, 64.0, 0.0)，生命值 20/20，饱食度 20/20，游戏模式：生存。\n附近有 2 名玩家：Steve(距离6m)、Alex(距离8m)。\n附近有 2 种生物：pig(x1)、cow(x1)。\n脚下地面有方块：grass_block dirt。",
	}, nil
}

var _ ClientInterface = (*MockClient)(nil)
