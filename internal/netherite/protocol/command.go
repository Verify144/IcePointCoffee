// Package protocol 实现 Minecraft Bedrock 协议。
package protocol

import (
	"fmt"
)

// CommandRequest 包 (ID=0xD2)
type CommandRequest struct {
	Command      string
	Origin       CommandOrigin
	Internal     bool
	UnLimited    bool
	Version      int32
}

// CommandOrigin 命令来源
type CommandOrigin struct {
	Origin      CommandOriginType
	UUID        [16]byte
	RequestID   string
	PlayerID    int64
}

// CommandOriginType 来源类型
type CommandOriginType uint32

const (
	CommandOriginPlayer    CommandOriginType = 0
	CommandOriginDevConsole CommandOriginType = 1
	CommandOriginTest      CommandOriginType = 2
	CommandOriginAutomation CommandOriginType = 3
	CommandOriginPlayerAuxilary CommandOriginType = 4
)

// EncodeCommandRequest 编码命令请求
func EncodeCommandRequest(cmd string, origin CommandOrigin, version int32) []byte {
	buf := NewWriter()
	buf.WriteByte(IDCommandRequest)

	// command string
	buf.WriteString(cmd)

	// origin
	buf.WriteVarint(uint64(origin.Origin))

	// uuid
	buf.Write(origin.UUID[:])

	// request id length + string
	reqID := origin.RequestID
	buf.WriteString(reqID)

	// player id
	buf.WriteInt64(origin.PlayerID)

	// internal + unlimited
	buf.WriteBool(false) // Internal
	buf.WriteBool(false) // UnLimited
	buf.WriteVarint(uint64(version))

	return buf.Bytes()
}

// DecodeCommandRequest 解码 CommandRequest
func DecodeCommandRequest(data []byte) (*CommandRequest, error) {
	if len(data) < 2 || data[0] != IDCommandRequest {
		return nil, fmt.Errorf("not a CommandRequest packet")
	}
	br := NewReader(data[1:])

	cmd := br.ReadString()

	origin := CommandOrigin{}
	origin.Origin = CommandOriginType(br.ReadVarint())
	_, _ = br.Read(origin.UUID[:])
	origin.RequestID = br.ReadString()
	origin.PlayerID = br.ReadInt64()

	version := br.ReadVarint()

	internal := br.ReadBool()
	unlimited := br.ReadBool()

	return &CommandRequest{
		Command:   cmd,
		Origin:    origin,
		Internal:  internal,
		UnLimited: unlimited,
		Version:   int32(version),
	}, nil
}

// CommandOutput 包 (ID=0xD3)
type CommandOutput struct {
	Origin       CommandOrigin
	OutputType   byte
	SuccessCount int32
	Messages     []CommandOutputMessage
}

// CommandOutputMessage 单条命令输出
type CommandOutputMessage struct {
	Success    bool
	MessageID  string
	Parameters []string
	Internal   bool
	Message    string
}

// EncodeCommandOutput 编码命令输出
func EncodeCommandOutput(origin CommandOrigin, messages []CommandOutputMessage) []byte {
	buf := NewWriter()
	buf.WriteByte(IDCommandOutput)

	// Origin
	buf.WriteVarint(uint64(origin.Origin))
	buf.Write(origin.UUID[:])
	buf.WriteString(origin.RequestID)
	buf.WriteVarint(uint64(origin.PlayerID))

	// OutputType
	buf.WriteByte(0)

	// Success count
	buf.WriteVarint(uint64(len(messages)))

	// Messages
	for _, msg := range messages {
		buf.WriteBool(msg.Success)
		buf.WriteString(msg.MessageID)
		buf.WriteVarint(uint64(len(msg.Parameters)))
		for _, p := range msg.Parameters {
			buf.WriteString(p)
		}
		buf.WriteBool(msg.Internal)
		buf.WriteString(msg.Message)
	}
	return buf.Bytes()
}

// DecodeCommandOutput 解码 CommandOutput
func DecodeCommandOutput(data []byte) (*CommandOutput, error) {
	if len(data) < 2 || data[0] != IDCommandOutput {
		return nil, fmt.Errorf("not a CommandOutput packet")
	}
	br := NewReader(data[1:])

	origin := CommandOrigin{}
	origin.Origin = CommandOriginType(br.ReadVarint())
	_, _ = br.Read(origin.UUID[:])
	origin.RequestID = br.ReadString()
	origin.PlayerID = br.ReadInt64()

	outputType := br.ReadByte()
	var successCount int32
	var messages []CommandOutputMessage

	msgCount := br.ReadVarint()
	messages = make([]CommandOutputMessage, msgCount)
	for i := 0; i < int(msgCount); i++ {
		msg := CommandOutputMessage{}
		msg.Success = br.ReadBool()
		msg.MessageID = br.ReadString()
		paramCount := br.ReadVarint()
		msg.Parameters = make([]string, paramCount)
		for j := 0; j < int(paramCount); j++ {
			msg.Parameters[j] = br.ReadString()
		}
		msg.Internal = br.ReadBool()
		msg.Message = br.ReadString()
		messages[i] = msg
	}

	return &CommandOutput{
		Origin:       origin,
		OutputType:   outputType,
		SuccessCount: successCount,
		Messages:     messages,
	}, nil
}

// SimpleCommandOutput 构造简单的成功消息
func SimpleCommandOutput(messages ...string) CommandOutput {
	var msgs []CommandOutputMessage
	for _, msg := range messages {
		msgs = append(msgs, CommandOutputMessage{
			Success:    true,
			MessageID:  "",
			Parameters: []string{},
			Internal:   false,
			Message:    msg,
		})
	}
	return CommandOutput{
		Origin:       CommandOrigin{Origin: CommandOriginDevConsole},
		OutputType:   0,
		SuccessCount: int32(len(messages)),
		Messages:     msgs,
	}
}