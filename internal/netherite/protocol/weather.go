package protocol

// ===== 新协议包 ID =====

const (
	IDWorldEvent            = 0x12 // 世界事件（粒子/天气/音效）
	IDWorldBorder          = 0x31 // 世界边界
	IDDifficulty           = 0x39 // 难度
	IDSpawnPosition        = 0x3A // 出生点
	IDUpdateAbilities      = 0x4C // 玩家能力
	IDTickSync            = 0x17 // 同步刻
	IDScriptMessage        = 0x7C // 脚本消息
	IDNetworkChunkPublisher = 0x79 // 区块发布
	IDPlayerFog           = 0xA0 // 雾
)

// ===== Weather (LevelEvent ID=0x12) =====

// WeatherType 天气类型
type WeatherType int32

const (
	WeatherClear    WeatherType = 0 // 晴天
	WeatherRain    WeatherType = 1 // 下雨
	WeatherThunder WeatherType = 2 // 雷雨
)

// EncodeWeatherPacket 编码天气事件（LevelEvent 0x12）
// 用法: weather := EncodeWeatherPacket(WeatherRain, 0, 64, 0)
func EncodeWeatherPacket(weather WeatherType, x, y, z float64) []byte {
	var eventID int16
	switch weather {
	case WeatherClear:
		eventID = 9802 // 停止下雨
	case WeatherRain:
		eventID = 9801 // 开始下雨
	case WeatherThunder:
		eventID = 9803 // 闪电
	}
	return EncodeLevelEvent(eventID, x, y, z, 0)
}

// ===== WorldBorder (0x31) =====

// WorldBorderAction 世界边界动作
type WorldBorderAction int32

const (
	WorldBorderSetSize   WorldBorderAction = 0
	WorldBorderLerpSize WorldBorderAction = 1
	WorldBorderSetCenter WorldBorderAction = 2
	WorldBorderWarning   WorldBorderAction = 4
)

// EncodeWorldBorderSize 设置世界边界半径
func EncodeWorldBorderSize(radius float64) []byte {
	buf := NewWriter()
	buf.WriteByte(IDWorldBorder)
	buf.WriteVarint(uint64(WorldBorderSetSize))
	buf.WriteVarint(uint64(0)) // warning blocks
	buf.WriteFloat64(radius)
	return buf.Bytes()
}

// EncodeWorldBorderLerp 平滑过渡世界边界
func EncodeWorldBorderLerp(oldRadius, newRadius float64, speed int64) []byte {
	buf := NewWriter()
	buf.WriteByte(IDWorldBorder)
	buf.WriteVarint(uint64(WorldBorderLerpSize))
	buf.WriteFloat64(oldRadius)
	buf.WriteFloat64(newRadius)
	buf.WriteVarint(uint64(speed))
	return buf.Bytes()
}

// EncodeWorldBorderCenter 设置世界边界中心
func EncodeWorldBorderCenter(x, z float64) []byte {
	buf := NewWriter()
	buf.WriteByte(IDWorldBorder)
	buf.WriteVarint(uint64(WorldBorderSetCenter))
	buf.WriteFloat64(x)
	buf.WriteFloat64(z)
	return buf.Bytes()
}

// EncodeWorldBorderWarning 设置边界警告
func EncodeWorldBorderWarning(blocks int32, time int32) []byte {
	buf := NewWriter()
	buf.WriteByte(IDWorldBorder)
	buf.WriteVarint(uint64(WorldBorderWarning))
	buf.WriteVarint(uint64(blocks))
	buf.WriteVarint(uint64(time))
	return buf.Bytes()
}

// ===== Difficulty (0x39) =====

// Difficulty 难度值
type Difficulty int32

const (
	DifficultyPeaceful Difficulty = 0
	DifficultyEasy    Difficulty = 1
	DifficultyNormal  Difficulty = 2
	DifficultyHard    Difficulty = 3
)

// EncodeSetDifficulty 设置世界难度
func EncodeSetDifficulty(d Difficulty) []byte {
	buf := NewWriter()
	buf.WriteByte(IDDifficulty)
	buf.WriteVarint(uint64(d))
	return buf.Bytes()
}

// ===== SpawnPosition (0x3A) =====

// SpawnPos 出生点数据
type SpawnPos struct {
	Position Vec3
	Dim     int32
}

// EncodeSpawnPositionPacket 编码出生点协议包
func EncodeSpawnPositionPacket(x, y, z float64, dim int32) []byte {
	buf := NewWriter()
	buf.WriteByte(IDSpawnPosition)
	buf.WriteVarint(0) // spawn type: world spawn
	buf.WriteFloat64(x)
	buf.WriteFloat64(y)
	buf.WriteFloat64(z)
	buf.WriteVarint(uint64(dim))
	buf.WriteBool(false) // triggered
	return buf.Bytes()
}

// ===== UpdateAbilities (0x4C) =====

// EncodeUpdateAbilitiesPacket 编码玩家能力更新（飞/疾跑等）
func EncodeUpdateAbilitiesPacket(fly bool) []byte {
	buf := NewWriter()
	buf.WriteByte(IDUpdateAbilities)
	buf.WriteByte(0) // base layer
	var abilities int32
	if fly {
		abilities |= 1 << 1 // MAY_FLY
		abilities |= 1 << 3 // FLYING
	}
	buf.WriteVarint(uint64(abilities))
	buf.WriteVarint(0) // ability value
	buf.WriteFloat32(0.05) // fly speed
	buf.WriteFloat32(0.1)  // walk speed
	return buf.Bytes()
}

// ===== TickSync (0x17) =====

// EncodeTickSync 编码刻同步（防作弊检测）
func EncodeTickSync(clientTick int64) []byte {
	buf := NewWriter()
	buf.WriteByte(IDTickSync)
	buf.WriteInt64(clientTick)
	buf.WriteInt64(clientTick)
	return buf.Bytes()
}

// ===== ScriptMessage (0x7C) =====

// EncodeScriptMessage 编码脚本消息
func EncodeScriptMessage(message string) []byte {
	buf := NewWriter()
	buf.WriteByte(IDScriptMessage)
	buf.WriteString("IcePoint")
	buf.WriteString(message)
	return buf.Bytes()
}
