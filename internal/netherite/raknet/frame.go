package raknet

import (
	"errors"
	"fmt"
	"io"
)

// Frame 表示一个 Raknet 帧。
// 帧结构 (普通类型):
//   flags (1 byte)
//   length_bits_16 (1 byte): length*8 in bits of remaining data
//   reliable_message_number (3 bytes, only if reliable bit set)
//   sequencing_bits_24 (3 bytes, only if sequenced/ordered)
//   ordering_channel (1 byte, only if sequenced/ordered)
//   ordering_index (3 bytes)
//   data (variable)
type Frame struct {
	Flags       byte
	ReliableIdx uint32
	SeqIdx      uint32
	OrderChan   byte
	Data        []byte
	// 用于分片
	SplitID     uint32
	SplitIndex  uint32
	SplitCount  uint32
	IsSplit     bool
}

// Encode 编码帧到字节流。
// 帧格式: flags(1) + length(1) + [reliable_idx(3)] + [seq_idx(3) + chan(1)] + [split(12)] + data
func (f *Frame) Encode() ([]byte, error) {
	// 构建 body 部分
	body := make([]byte, 0, 20+len(f.Data))

	if f.Flags&FlagReliable != 0 {
		body = append(body,
			byte(f.ReliableIdx),
			byte(f.ReliableIdx>>8),
			byte(f.ReliableIdx>>16),
		)
	}
	if f.Flags&FlagReliableOrdered != 0 || f.Flags&FlagReliableSequenced != 0 {
		body = append(body,
			byte(f.SeqIdx),
			byte(f.SeqIdx>>8),
			byte(f.SeqIdx>>16),
			f.OrderChan,
		)
	}
	if f.IsSplit {
		body = append(body,
			byte(f.SplitID), byte(f.SplitID>>8), byte(f.SplitID>>16), byte(f.SplitID>>24),
			byte(f.SplitIndex), byte(f.SplitIndex>>8), byte(f.SplitIndex>>16), byte(f.SplitIndex>>24),
			byte(f.SplitCount), byte(f.SplitCount>>8), byte(f.SplitCount>>16), byte(f.SplitCount>>24),
		)
	}
	body = append(body, f.Data...)

	// length = body_len (value is body bytes - 1, max 0xFF)
	bodyLen := len(body)
	if bodyLen == 0 || bodyLen > 256 {
		return nil, fmt.Errorf("body length %d invalid", bodyLen)
	}

	// total: flags(1) + length(1) + body
	result := make([]byte, 0, bodyLen+2)
	result = append(result, f.Flags)
	result = append(result, byte(bodyLen-1)) // length field
	result = append(result, body...)
	return result, nil
}

// DecodeFrame 解码帧。
// 返回 (frame, consumed_bytes, error)
func DecodeFrame(data []byte) (*Frame, int, error) {
	if len(data) < 2 {
		return nil, 0, io.ErrUnexpectedEOF
	}

	flags := data[0]
	lengthField := int(data[1])
	bodyLen := lengthField + 1 // actual body bytes

	needed := 2 + bodyLen
	if len(data) < needed {
		return nil, 0, fmt.Errorf("frame: need %d bytes got %d (bodyLen=%d)", needed, len(data), bodyLen)
	}

	offset := 2
	frame := &Frame{Flags: flags}

	if flags&FlagReliable != 0 {
		if offset+3 > needed {
			return nil, 0, errors.New("reliable index truncated")
		}
		frame.ReliableIdx = uint32(data[offset]) |
			uint32(data[offset+1])<<8 |
			uint32(data[offset+2])<<16
		offset += 3
	}
	if flags&FlagReliableOrdered != 0 || flags&FlagReliableSequenced != 0 {
		if offset+4 > needed {
			return nil, 0, errors.New("ordering fields truncated")
		}
		frame.SeqIdx = uint32(data[offset]) |
			uint32(data[offset+1])<<8 |
			uint32(data[offset+2])<<16
		frame.OrderChan = data[offset+3]
		offset += 4
	}

	// 数据
	dataStart := offset
	dataEnd := needed
	frame.Data = make([]byte, dataEnd-dataStart)
	copy(frame.Data, data[dataStart:dataEnd])
	return frame, needed, nil
}

// UnconnectedPing Ping 包（无连接）
// 结构：PacketID(1) + PingTime(8) + ClientGUID(4) + Magic(16) + ClientName
type UnconnectedPing struct {
	PingTime  int64
	ClientGUID int32
	ClientName string
}

// Encode 编码 Ping
func (p *UnconnectedPing) Encode() []byte {
	nameBytes := []byte(p.ClientName)
	buf := make([]byte, 0, 30+len(nameBytes))
	buf = append(buf, IDUnconnectedPing)
	// Ping time
	for i := 0; i < 8; i++ {
		buf = append(buf, byte(p.PingTime>>(i*8)))
	}
	// Client GUID
	buf = append(buf,
		byte(p.ClientGUID),
		byte(p.ClientGUID>>8),
		byte(p.ClientGUID>>16),
		byte(p.ClientGUID>>24),
	)
	// Magic
	buf = append(buf, RaknetMagic[:]...)
	// Name
	buf = append(buf, nameBytes...)
	return buf
}

// DecodePing 解析 Ping
func DecodePing(data []byte) (*UnconnectedPing, error) {
	if len(data) < 30 {
		return nil, errors.New("ping too short")
	}
	if data[0] != IDUnconnectedPing {
		return nil, errors.New("not ping")
	}
	offset := 1
	var pingTime int64
	for i := 0; i < 8; i++ {
		pingTime |= int64(data[offset+i]) << (i * 8)
	}
	offset += 8
	clientGUID := int32(data[offset]) | int32(data[offset+1])<<8 |
		int32(data[offset+2])<<16 | int32(data[offset+3])<<24
	offset += 4
	// skip magic check (offset+=16)
	offset += 16
	name := string(data[offset:])
	return &UnconnectedPing{
		PingTime:   pingTime,
		ClientGUID: clientGUID,
		ClientName: name,
	}, nil
}

// UnconnectedPong Pong 响应包
// 结构: PacketID(1) + PingTime(8) + ServerGUID(8) + Magic(16) + ServerName
type UnconnectedPong struct {
	PingTime  int64
	ServerGUID int64
	ServerName string
}

// Encode 编码 Pong
func (p *UnconnectedPong) Encode() []byte {
	nameBytes := []byte(p.ServerName)
	buf := make([]byte, 0, 34+len(nameBytes))
	buf = append(buf, IDUnconnectedPong)
	for i := 0; i < 8; i++ {
		buf = append(buf, byte(p.PingTime>>(i*8)))
	}
	for i := 0; i < 8; i++ {
		buf = append(buf, byte(p.ServerGUID>>(i*8)))
	}
	buf = append(buf, RaknetMagic[:]...)
	buf = append(buf, nameBytes...)
	return buf
}

// OpenConnectionRequest1 第一阶段连接请求
// 结构: PacketID(1) + Magic(16) + ProtocolVersion(1) + MTU(0..1492, 填充)
type OpenConnectionRequest1 struct {
	MTU            uint16
	ProtocolVersion byte
}

// Encode 编码 OpenConnectionRequest1
func (r *OpenConnectionRequest1) Encode() []byte {
	buf := make([]byte, 0, int(r.MTU))
	buf = append(buf, IDOpenConnectionRequest1)
	buf = append(buf, RaknetMagic[:]...)
	buf = append(buf, r.ProtocolVersion)
	// 填充到 MTU
	for len(buf) < int(r.MTU) {
		buf = append(buf, 0)
	}
	return buf
}

// OpenConnectionReply1 第一阶段响应
// 结构: PacketID(1) + Magic(16) + ServerGUID(8) + MTU(2) + HasSecurity(1) + CookiePort(4, hasSecurity=1)
type OpenConnectionReply1 struct {
	ServerGUID  int64
	MTU         uint16
	HasSecurity bool
	CookiePort  uint16
}

// Encode 编码 OpenConnectionReply1
func (r *OpenConnectionReply1) Encode() []byte {
	buf := make([]byte, 0, 30)
	buf = append(buf, IDOpenConnectionReply1)
	buf = append(buf, RaknetMagic[:]...)
	for i := 0; i < 8; i++ {
		buf = append(buf, byte(r.ServerGUID>>(i*8)))
	}
	buf = append(buf, byte(r.MTU), byte(r.MTU>>8))
	if r.HasSecurity {
		buf = append(buf, 1)
		buf = append(buf, byte(r.CookiePort), byte(r.CookiePort>>8), 0, 0)
	} else {
		buf = append(buf, 0)
	}
	return buf
}

// OpenConnectionRequest2 第二阶段请求
// 结构: PacketID(1) + Magic(16) + Cookie(4) + HasSecurity(1) + ClientAddress + MTU(2) + ClientGUID(8)
type OpenConnectionRequest2 struct {
	Cookie   uint32
	MTU      uint16
	ClientGUID int64
}

// Encode 编码 OpenConnectionRequest2
func (r *OpenConnectionRequest2) Encode() []byte {
	buf := make([]byte, 0, 30)
	buf = append(buf, IDOpenConnectionRequest2)
	buf = append(buf, RaknetMagic[:]...)
	buf = append(buf,
		byte(r.Cookie),
		byte(r.Cookie>>8),
		byte(r.Cookie>>16),
		byte(r.Cookie>>24),
	)
	// HasSecurity = false
	buf = append(buf, 0)
	// Client address: 4 (IP v4) + 2 (port)
	// 这里发送全 0
	buf = append(buf, 0, 0, 0, 0, 0, 0)
	// MTU
	buf = append(buf, byte(r.MTU), byte(r.MTU>>8))
	// Client GUID
	for i := 0; i < 8; i++ {
		buf = append(buf, byte(r.ClientGUID>>(i*8)))
	}
	return buf
}

// OpenConnectionReply2 第二阶段响应
// 结构: PacketID(1) + Magic(16) + ServerGUID(8) + ClientAddress + MTU(2) + HasSecurity(1) + CookiePort
type OpenConnectionReply2 struct {
	ServerGUID  int64
	MTU         uint16
	HasSecurity bool
	CookiePort  uint16
}

// Encode 编码 OpenConnectionReply2
func (r *OpenConnectionReply2) Encode() []byte {
	buf := make([]byte, 0, 30)
	buf = append(buf, IDOpenConnectionReply2)
	buf = append(buf, RaknetMagic[:]...)
	for i := 0; i < 8; i++ {
		buf = append(buf, byte(r.ServerGUID>>(i*8)))
	}
	// Client address (6 bytes 0)
	buf = append(buf, 0, 0, 0, 0, 0, 0)
	// MTU
	buf = append(buf, byte(r.MTU), byte(r.MTU>>8))
	if r.HasSecurity {
		buf = append(buf, 1)
		buf = append(buf, byte(r.CookiePort), byte(r.CookiePort>>8), 0, 0)
	} else {
		buf = append(buf, 0)
	}
	return buf
}

// DecodePong 解析 Pong 响应
func DecodePong(data []byte) (*UnconnectedPong, error) {
	if len(data) < 34 {
		return nil, errors.New("pong too short")
	}
	if data[0] != IDUnconnectedPong {
		return nil, errors.New("not pong")
	}
	offset := 1
	var pingTime int64
	for i := 0; i < 8; i++ {
		pingTime |= int64(data[offset+i]) << (i * 8)
	}
	offset += 8
	var serverGUID int64
	for i := 0; i < 8; i++ {
		serverGUID |= int64(data[offset+i]) << (i * 8)
	}
	offset += 8
	offset += 16 // magic
	name := string(data[offset:])
	return &UnconnectedPong{
		PingTime:   pingTime,
		ServerGUID: serverGUID,
		ServerName: name,
	}, nil
}
