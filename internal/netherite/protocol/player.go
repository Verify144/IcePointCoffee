// Package protocol 实现 Minecraft Bedrock 协议。
package protocol

// MovePlayerMode 移动模式
type MovePlayerMode byte

const (
	MovePlayerModeNormal MovePlayerMode = iota
	MovePlayerModeReset
	MovePlayerModeTeleport
	MovePlayerModeRotation
)

// MovePlayer 包 (ID=0xF5)
type MovePlayer struct {
	RuntimeID uint64
	Position Vec3
	Pitch    float32
	Yaw      float32
	HeadYaw  float32
	Mode     MovePlayerMode
	OnGround bool
	Tick     uint64
}

// Vec3 三维坐标
type Vec3 struct {
	X, Y, Z float64
}

// NewVec3 创建坐标
func NewVec3(x, y, z float64) Vec3 {
	return Vec3{X: x, Y: y, Z: z}
}

// EncodeMovePlayer 编码 MovePlayer 包
func EncodeMovePlayer(runtimeID uint64, x, y, z float64) []byte {
	buf := NewWriter()
	buf.WriteByte(IDMovePlayer)
	buf.WriteVarint(runtimeID)
	buf.WriteFloat64(x)
	buf.WriteFloat64(y)
	buf.WriteFloat64(z)
	buf.WriteFloat32(0) // pitch
	buf.WriteFloat32(0) // yaw
	buf.WriteFloat32(0) // head yaw
	buf.WriteByte(byte(MovePlayerModeNormal))
	buf.WriteBool(false) // on ground
	buf.WriteVarint(0)  // tick
	return buf.Bytes()
}

// EncodeTeleport 编码传送
func EncodeTeleport(runtimeID uint64, x, y, z float64) []byte {
	buf := NewWriter()
	buf.WriteByte(IDMovePlayer)
	buf.WriteVarint(runtimeID)
	buf.WriteFloat64(x)
	buf.WriteFloat64(y)
	buf.WriteFloat64(z)
	buf.WriteFloat32(0)
	buf.WriteFloat32(0)
	buf.WriteFloat32(0)
	buf.WriteByte(byte(MovePlayerModeTeleport))
	buf.WriteBool(false)
	buf.WriteVarint(0)
	return buf.Bytes()
}

// SetTitleAction 标题操作
type SetTitleAction int32

const (
	SetTitleSetTitle SetTitleAction = iota
	SetTitleSetSubtitle
	SetTitleSetActionBar
	SetTitleClearTitle
	SetTitleResetTitle
)

// SetTitle 包 (ID=0x5C)
type SetTitle struct {
	TitleType SetTitleAction
	Title    string
	FadeIn   int32
	Stay     int32
	FadeOut  int32
}

// EncodeSetTitle 编码 SetTitle
func EncodeSetTitle(titleType SetTitleAction, title string) []byte {
	buf := NewWriter()
	buf.WriteByte(IDSetTitle)
	buf.WriteVarint(uint64(titleType))
	buf.WriteString(title)
	buf.WriteVarint(100) // fade in (ticks)
	buf.WriteVarint(100) // stay (ticks)
	buf.WriteVarint(100) // fade out (ticks)
	return buf.Bytes()
}

// AnimateAction 动画动作
type AnimateAction int32

const (
	AnimateActionSwing ArmAnimation = iota
	AnimateActionWakeUp
	AnimateActionCriticalHit
	AnimateActionMagicCriticalHit
)

// ArmAnimation 手臂动画
type ArmAnimation int32

// Animate 包 (ID=0xE1)
type Animate struct {
	Action  AnimateAction
	Tick    int64
}

// EncodeAnimate 编码动画
func EncodeAnimate(action int32) []byte {
	buf := NewWriter()
	buf.WriteByte(IDAnimate)
	buf.WriteVarint(uint64(action))
	buf.WriteVarint(0) // tick
	return buf.Bytes()
}

// EncodeSwingAnimation 编码摆臂动画
func EncodeSwingAnimation() []byte {
	return EncodeAnimate(int32(AnimateActionSwing))
}

// LevelEvent 包 (ID=0x18)
// 用于触发世界事件，如雷击、粒子等
type LevelEvent struct {
	EventID  int16
	Position Vec3
	Data     int32
}

// EncodeLevelEvent 编码世界事件
func EncodeLevelEvent(eventID int16, x, y, z float64, data int32) []byte {
	buf := NewWriter()
	buf.WriteByte(0x18)
	buf.WriteInt16(eventID)
	buf.WriteFloat32(float32(x))
	buf.WriteFloat32(float32(y))
	buf.WriteFloat32(float32(z))
	buf.WriteInt32(data)
	return buf.Bytes()
}

// ParticleEvent 粒子事件
type ParticleEvent struct {
	ParticleID int32
	Position   Vec3
}

// EncodeParticle 编码粒子效果
func EncodeParticle(particleID int32, x, y, z float64) []byte {
	buf := NewWriter()
	buf.WriteByte(0x17)
	buf.WriteInt32(particleID)
	buf.WriteFloat32(float32(x))
	buf.WriteFloat32(float32(y))
	buf.WriteFloat32(float32(z))
	buf.WriteBool(false)
	return buf.Bytes()
}

// UpdateAttributes 包 (ID=0xE2)
type UpdateAttributes struct {
	RuntimeID uint64
	Attributes []Attribute
	Tick     int64
}

// Attribute 属性
type Attribute struct {
	Name   string
	Min    float64
	Max    float64
	Value  float64
}

// EncodeUpdateAttributes 编码属性更新
func EncodeUpdateAttributes(runtimeID uint64, attrs []Attribute) []byte {
	buf := NewWriter()
	buf.WriteByte(IDUpdateAttributes)
	buf.WriteVarint(runtimeID)
	buf.WriteVarint(uint64(len(attrs)))
	for _, a := range attrs {
		buf.WriteString(a.Name)
		buf.WriteFloat64(a.Min)
		buf.WriteFloat64(a.Max)
		buf.WriteFloat64(a.Value)
	}
	buf.WriteVarint(0) // tick
	return buf.Bytes()
}

// EncodePlayerHealth 设置玩家生命值
func EncodePlayerHealth(health float64) []byte {
	return EncodeUpdateAttributes(0, []Attribute{
		{Name: "minecraft:health", Min: 0, Max: 20, Value: health},
	})
}
