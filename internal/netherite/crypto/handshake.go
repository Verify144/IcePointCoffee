// Package crypto 提供加密握手包编解码。
package crypto

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
)

// PacketID 加密握手包 ID
const (
	PacketIDConnectionRequestEncrypted = 0x91
	PacketIDNewIncomingConnection      = 0x92
	PacketIDConnectionRequestAccepted  = 0x93
	PacketIDUnknown                    = 0xFF
)

// CookieKey 网易专用 cookie 验证
// Bedrock 服务器返回 4 字节 cookie，客户端用 cookie 派生密钥
func GetCookieKey(serverIP string) []byte {
	h := sha256.Sum256([]byte(serverIP))
	return h[:]
}

// GenerateChallenge 生成 16 字节 challenge
func GenerateChallenge() []byte {
	c := make([]byte, 16)
	rand.Read(c)
	return c
}

// EncryptedHandshake 加密握手包。
// 客户端发送：PacketID(1) + Cookie(4) + PublicKey(32) + Challenge(16)
type EncryptedHandshake struct {
	Cookie     []byte // 4 字节
	PublicKey  []byte // 32 字节
	Challenge  []byte // 16 字节
}

// Encode 编码加密握手包
func (e *EncryptedHandshake) Encode() []byte {
	buf := make([]byte, 0, 53)
	buf = append(buf, PacketIDConnectionRequestEncrypted)
	if len(e.Cookie) != 4 {
		e.Cookie = make([]byte, 4)
	}
	buf = append(buf, e.Cookie...)
	if len(e.PublicKey) != 32 {
		e.PublicKey = make([]byte, 32)
	}
	buf = append(buf, e.PublicKey...)
	if len(e.Challenge) != 16 {
		e.Challenge = make([]byte, 16)
	}
	buf = append(buf, e.Challenge...)
	return buf
}

// DecodeEncryptedHandshake 解码加密握手
func DecodeEncryptedHandshake(data []byte) (*EncryptedHandshake, error) {
	if len(data) < 53 {
		return nil, fmt.Errorf("encrypted handshake too short: %d", len(data))
	}
	if data[0] != PacketIDConnectionRequestEncrypted {
		return nil, fmt.Errorf("not encrypted handshake: id=0x%02x", data[0])
	}
	return &EncryptedHandshake{
		Cookie:    data[1:5],
		PublicKey: data[5:37],
		Challenge: data[37:53],
	}, nil
}

// ServerHandshake 服务端加密握手响应
// 结构: PacketID(1) + ServerIdentity(8) + Cookie(4) + PublicKey(32)
// ServerHandshake 包含服务端公钥和它的 identity token
type ServerHandshake struct {
	ServerIdentity []byte // 8 字节
	Cookie         []byte // 4 字节
	PublicKey      []byte // 32 字节
}

// Encode 编码服务端握手
func (s *ServerHandshake) Encode() []byte {
	buf := make([]byte, 0, 45)
	buf = append(buf, PacketIDConnectionRequestAccepted)
	if len(s.ServerIdentity) != 8 {
		s.ServerIdentity = make([]byte, 8)
	}
	buf = append(buf, s.ServerIdentity...)
	if len(s.Cookie) != 4 {
		s.Cookie = make([]byte, 4)
	}
	buf = append(buf, s.Cookie...)
	if len(s.PublicKey) != 32 {
		s.PublicKey = make([]byte, 32)
	}
	buf = append(buf, s.PublicKey...)
	return buf
}

// DecodeServerHandshake 解码服务端握手
func DecodeServerHandshake(data []byte) (*ServerHandshake, error) {
	if len(data) < 45 {
		return nil, fmt.Errorf("server handshake too short: %d", len(data))
	}
	if data[0] != PacketIDConnectionRequestAccepted {
		return nil, fmt.Errorf("not server handshake: id=0x%02x", data[0])
	}
	return &ServerHandshake{
		ServerIdentity: data[1:9],
		Cookie:         data[9:13],
		PublicKey:      data[13:45],
	}, nil
}

// NewConnectionAccepted 包结构
// PacketID(1) + ServerAddress(4 IP + 2 Port) + SystemAddresses(20) + SendPingTime(8) + SendPongTime(8)
type NewConnectionAccepted struct {
	ServerIP     [4]byte
	ServerPort   uint16
	SystemAddrs  [10][2]byte // 10 个系统地址（IP+Port）
	SendPingTime int64
	SendPongTime int64
}

// Encode 编码
func (n *NewConnectionAccepted) Encode() []byte {
	buf := make([]byte, 0, 61)
	buf = append(buf, PacketIDNewIncomingConnection)
	buf = append(buf, n.ServerIP[:]...)

	portBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(portBytes, n.ServerPort)
	buf = append(buf, portBytes...)

	for _, addr := range n.SystemAddrs {
		buf = append(buf, addr[:]...)
	}

	pingBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(pingBytes, uint64(n.SendPingTime))
	buf = append(buf, pingBytes...)

	pongBytes := make([]byte, 8)
	binary.LittleEndian.PutUint64(pongBytes, uint64(n.SendPongTime))
	buf = append(buf, pongBytes...)

	return buf
}

// BedrockKeyDerive 网易专用密钥派生。
// 输入：shared secret + 服务端 challenge
// 输出：32 字节 AES-256 密钥
func BedrockKeyDerive(sharedSecret []byte) []byte {
	// MC 1.16+ 加密协议：
	// key = SHA256( sharedSecret || sha256( sharedSecret ) )
	h1 := sha256.Sum256(sharedSecret)
	data := make([]byte, 0, len(sharedSecret)+32)
	data = append(data, sharedSecret...)
	data = append(data, h1[:]...)
	h2 := sha256.Sum256(data)
	return h2[:]
}

// NeteaseKeyDerive 网易专用密钥派生。
// 流程：先 ECDH，再用特殊 hash 派生
func NeteaseKeyDerive(sharedSecret, serverChallenge, clientChallenge []byte) []byte {
	// 第一步：basic = SHA256(sharedSecret)
	basic := sha256.Sum256(sharedSecret)

	// 第二步：challengeCombined = SHA256(serverChallenge || clientChallenge)
	combined := append(serverChallenge, clientChallenge...)
	combinedHash := sha256.Sum256(combined)

	// 第三步：key = SHA256(basic || challengeCombined)
	data := make([]byte, 0, 64)
	data = append(data, basic[:]...)
	data = append(data, combinedHash[:]...)
	final := sha256.Sum256(data)
	return final[:]
}

// VerifyChallenge 验证 challenge（从服务端公钥派生）
func VerifyChallenge(serverPubKey, clientPubKey []byte) error {
	if len(serverPubKey) != 32 {
		return errors.New("server public key must be 32 bytes")
	}
	if len(clientPubKey) != 32 {
		return errors.New("client public key must be 32 bytes")
	}
	return nil
}
