// Package mc 提供 MC 服务器操控接口。
package mc

import "context"

// Status 服务器状态
type Status struct {
	Connected     bool    `json:"connected"`
	PlayerName    string  `json:"player_name,omitempty"`
	Position      *Vec3   `json:"position,omitempty"`
	Health        float32 `json:"health,omitempty"`
	GameMode      string  `json:"game_mode,omitempty"`
	WorldTime     int64   `json:"world_time,omitempty"`
	ServerAddress string  `json:"server_address,omitempty"`
}

// Vec3 三维坐标
type Vec3 struct {
	X, Y, Z float64 `json:"x,y,z"`
}

// ClientInterface MC 客户端接口（用于工具测试和实际执行）
type ClientInterface interface {
	// 基础
	IsConnected() bool
	Connect(ctx context.Context) error
	Close() error
	Status() Status

	// 命令
	SendCommand(ctx context.Context, cmd string) (string, error)
	SendChat(msg string) error
	SendTellRaw(target, message string) error
	SendTitle(target, title, subtitle string) error

	// 玩家
	Teleport(target string, x, y, z float64) error
	SetGameMode(target, mode string) error
	GiveItem(target, item string, count int) error
	Kill(target string) error
	Respawn(x, y, z float64) error
	SwingArm() error

	// 实体交互
	AttackEntity(targetID uint64) error
	SpawnEntity(entityType, name string, x, y, z float64) error
	RemoveEntity(entityID uint64) error

	// 视觉
	PlaySound(soundID string, x, y, z float64, volume, pitch float32) error
	StopSound(soundID string) error
	EmitParticle(particleID int32, x, y, z float64) error
	SetBossBar(target string, title string, healthPercent float32) error

	// 世界
	SetBlock(x, y, z int, block string) error
	Fill(x1, y1, z1, x2, y2, z2 int, block string) error
	SetTime(time string) error
	SetWeather(weather string) error
}
