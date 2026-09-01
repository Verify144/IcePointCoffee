// Package raknet 实现 Raknet 协议（Bedrock Edition）。
// 参考文献：https://github.com/ruby-net/raknet/blob/master/README.md
package raknet

// 协议版本
const ProtocolVersion = 10

// MTU 大小
const (
	MTUMin        = 576
	MTUMaxClassic = 1400
	MTUDefault    = 1200
)

// Raknet 包 ID 范围
const (
	IDUnconnectedPing           = 0x01
	IDUnconnectedPong          = 0x1C
	IDOpenConnectionRequest1   = 0x05
	IDOpenConnectionReply1     = 0x06
	IDOpenConnectionRequest2   = 0x07
	IDOpenConnectionReply2     = 0x08
	IDConnectionRequest        = 0x09
	IDConnectionResponse       = 0x10
	IDNewConnection            = 0x13
	IDDisconnect               = 0x15
	IDIncompatibleProtocol     = 0x1A
)

// 帧标志位
const (
	FlagReliable            = 1 << 5
	FlagReliableOrdered     = 1<<5 | 1<<4
	FlagReliableSequenced   = 1 << 4
	FlagAcknowledged         = 1 << 7
)

// 帧优先级
const (
	FramePriorityNormal        = 0
	FramePriorityImmediate     = 1
	FramePriorityRealtime      = 2
)

// 帧类型
const (
	FrameTypeNormal            = 0
	FrameTypeSplit             = 1
	FrameTypeACK               = 2
	FrameTypeNAK               = 3
)

// 分片包 ID
const (
	IDSplit1 = 0x84
	IDSplit2 = 0x85
	IDSplit3 = 0x86
	IDSplit4 = 0x87
)

// 心跳间隔
const HeartbeatInterval = 5000 // ms

// 丢包阈值
const PacketLossThreshold = 10 // percent

// Raknet Magic 标识
var RaknetMagic = [16]byte{
	0x00, 0xFF, 0xFF, 0x00, 0xFE, 0xFE, 0xFE, 0xFE,
	0xFD, 0xFD, 0xFD, 0xFD, 0x12, 0x34, 0x56, 0x78,
}

// 连接状态
type State uint8

const (
	StateUnconnected       State = iota
	StateConnecting
	StateConnected
	StateDisconnected
)
