// Package protocol 实现 Minecraft Bedrock 协议包。
package protocol

import (
	"bytes"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
)

// Bedrock 协议版本 (1.21.x)
const ProtocolVersion = 766

// PacketID 定义
const (
	IDLogin                          = 0x01
	IDPlayStatus                     = 0x02
	IDServerToClientHandshake        = 0x03
	IDClientToServerHandshake        = 0x04
	IDDisconnect                     = 0x05
	IDResourcePacksInfo              = 0x06
	IDResourcePackStack              = 0x07
	IDResourcePackClientResponse     = 0x08
	IDText                           = 0x09
	IDSetTime                        = 0x0A
	IDStartGame                      = 0x0B
	IDAddPlayer                      = 0x0C
	IDAddActor                       = 0x0D
	IDRemoveActor                    = 0x0F
	IDAddItemActor                   = 0x10
	IDServerSettingsRequest          = 0x3B
	IDServerSettingsResponse         = 0x3C
	IDRequestPermissions             = 0x52
	IDPermissions                    = 0x53
	IDPlayerAction                   = 0x6B
	IDNetworkSettings                = 0x7F
	IDNetworkStackLatency            = 0x84
	IDSubClientLogin                 = 0x8B
	IDWSHeartbeat                    = 0x8E
	IDWSConnect                      = 0x8F
	IDWebSocketCommand               = 0x90
	IDPong                           = 0x91
	IDLanguageData                   = 0x92
	IDAvailableActorIdentifiers      = 0x93
	IDActorPickRequest               = 0x94
	IDPlayerSkin                     = 0x9F
	IDSubClientEntry                 = 0xA0
	IDClientCacheStatus               = 0xA4
	IDMappability                    = 0xA5
	IDLevelInfo                      = 0xA6
	IDEmote                          = 0xAA
	IDEmoteList                      = 0xAB
	IDSetActorData                   = 0xAF
	IDNeteaseJson                    = 0xBF
	IDCodeBuilder                    = 0xC1
	IDCommandRequest                 = 0xD2
	IDCommandOutput                  = 0xD3
	IDUpdateAttributes               = 0xE2
	IDNpcRequest                     = 0xEE
	IDPhotoTransfer                  = 0xF1
	IDInventoryTransaction           = 0xF2
	IDInventoryContent               = 0xF3
	IDInventorySlot                  = 0xF4
	IDMovePlayer                     = 0xF5
	IDAdventure                      = 0xF7
	IDContainerOpen                  = 0xF8
	IDContainerClose                 = 0xF9
	IDSetTitle                       = 0x5C
	IDAnimate                        = 0xE1
	IDNeteaseMiscellaneous           = 0xF8
	IDPyRpc                          = 0xFF
)

// ===== Login =====

// Login 包结构 (ID=0x01)
// 客户端发送，包含身份信息和公钥
type Login struct {
	ProtocolVersion int32
	Tokens          LoginTokens
}

// LoginTokens JWT-like tokens from Mojang / Xbox
type LoginTokens struct {
	ExtraData     ExtraData     `json:"extraData"`
	Identity      string        `json:"identity"`
	GameVersion   string        `json:"GameVersion"`
	ChainData     []ChainLink   `json:"chainData"`
}

// ChainLink 令牌链中的一环
type ChainLink struct {
	PublicKey       string `json:"publicKey"`
	Signature       string `json:"signature"`
	Identity        string `json:"identity,omitempty"`
	DisplayName     string `json:"displayName,omitempty"`
	Xuid            string `json:"Xuid,omitempty"`
}

// ExtraData 令牌中的额外信息
type ExtraData struct {
	DisplayName string `json:"displayName"`
	Identity    string `json:"identity"`
	XUID        string `json:"XUID"`
}

// DecodeLogin 解码 Login 包
func DecodeLogin(data []byte) (*Login, error) {
	if len(data) < 1 {
		return nil, fmt.Errorf("login packet too short")
	}

	br := bytes.NewReader(data)

	// 协议版本 varint
	proto, err := ReadVarintBR(br)
	if err != nil {
		return nil, fmt.Errorf("read proto: %w", err)
	}

	// tokens 字符串 (length-prefixed JSON)
	tokenLen, err := ReadVarintBR(br)
	if err != nil {
		return nil, fmt.Errorf("read token len: %w", err)
	}
	tokenData := make([]byte, tokenLen)
	if _, err := io.ReadFull(br, tokenData); err != nil {
		return nil, fmt.Errorf("read token data: %w", err)
	}

	var tokens LoginTokens
	if err := json.Unmarshal(tokenData, &tokens); err != nil {
		return nil, fmt.Errorf("parse tokens: %w", err)
	}

	return &Login{
		ProtocolVersion: int32(proto),
		Tokens:         tokens,
	}, nil
}

// EncodeLogin 编码 Login 包
func EncodeLogin(protocol int32, tokens LoginTokens) ([]byte, error) {
	tokenJSON, err := json.Marshal(tokens)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)

	// Protocol version varint
	WriteVarintBytes(buf, uint64(protocol))

	// Token JSON length + data
	WriteVarintBytes(buf, uint64(len(tokenJSON)))
	buf.Write(tokenJSON)

	return buf.Bytes(), nil
}

// ===== PlayStatus =====

// PlayStatus 状态码
const (
	PlayStatusLoginSuccess           = 0
	PlayStatusLoginFailedClient      = 1
	PlayStatusLoginFailedServer      = 2
	PlayStatusPlayerSpawn            = 3
	PlayStatusLoginFailedInvalidTenant = 4
	PlayStatusLoginFailedVanillaEdu  = 5
	PlayStatusLoginFailedEduVanilla  = 6
	PlayStatusLoginFailedServerFull  = 7
)

// PlayStatus 包 (ID=0x02)
type PlayStatus struct {
	Status int32
}

// ===== Text =====
// TextType, Text 见 text.go

// CommandRequest, CommandOrigin, CommandOriginType, CommandOutput 等定义见 command.go

// ===== NeteaseJson =====

// NeteaseJson 网易私有 JSON 包 (ID=0xBF)
// 用于发送网易特有事件，如 LOGIN_UID
type NeteaseJson struct {
	EventName string `json:"eventName"`
	ResID     string `json:"resid"`
	UID       string `json:"uid"`
}

// EncodeNeteaseJson 编码 NeteaseJson 包
func EncodeNeteaseJson(eventName, resid, uid string) []byte {
	data, _ := json.Marshal(NeteaseJson{
		EventName: eventName,
		ResID:     resid,
		UID:       uid,
	})
	buf := new(bytes.Buffer)
	buf.WriteByte(IDNeteaseJson)
	WriteVarintBytes(buf, uint64(len(data)))
	buf.Write(data)
	return buf.Bytes()
}

// ===== PyRpc =====

// PyRpcOperationType 操作类型
type PyRpcOperationType byte

const (
	PyRpcOperationTypeSend       PyRpcOperationType = 0
	PyRpcOperationTypeResponse    PyRpcOperationType = 1
	PyRpcOperationTypeCall       PyRpcOperationType = 2
)

// PyRpc 网易私有 RPC 包 (ID=0xFF)
// Value: []any{ "MethodName", []any{args...}, callbackID }
// OperationType: 0=send, 1=response, 2=call
type PyRpc struct {
	Value          []any
	CallbackID     int64
	OperationType  PyRpcOperationType
}

// EncodePyRpc 编码 PyRpc 包
func EncodePyRpc(method string, args []any, opType PyRpcOperationType) []byte {
	value := []any{method, args, nil}
	data, _ := json.Marshal(value)

	buf := new(bytes.Buffer)
	buf.WriteByte(IDPyRpc)
	WriteVarintBytes(buf, uint64(opType))
	WriteVarintBytes(buf, 0) // callback id = 0
	WriteVarintBytes(buf, uint64(len(data)))
	buf.Write(data)
	return buf.Bytes()
}

// DecodePyRpc 解码 PyRpc 包
func DecodePyRpc(data []byte) (*PyRpc, error) {
	if len(data) < 1 || data[0] != IDPyRpc {
		return nil, fmt.Errorf("not a PyRpc packet")
	}
	br := bytes.NewReader(data[1:])
	opType, _ := ReadVarintBR(br)
	callbackID, _ := ReadVarintBR(br)
	dataLen, _ := ReadVarintBR(br)
	jsonData := make([]byte, dataLen)
	io.ReadFull(br, jsonData)

	var value []any
	json.Unmarshal(jsonData, &value)
	if len(value) < 2 {
		return nil, fmt.Errorf("invalid PyRpc value")
	}

	return &PyRpc{
		Value:         value,
		CallbackID:    int64(callbackID),
		OperationType: PyRpcOperationType(opType),
	}, nil
}

// ===== WebSocket =====

// WSConnect WebSocket 连接包 (ID=0x8F)
type WSConnect struct {
	Token string
}

// WSCmd WebSocket 命令包 (ID=0x90)
type WSCmd struct {
	Command string
}

// EncodeWSCmd 编码 WebSocket 命令
func EncodeWSCmd(command string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(IDWebSocketCommand)
	WriteVarintBytes(buf, 0) // 0 unknown
	WriteVarintBytes(buf, uint64(len(command)))
	buf.WriteString(command)
	return buf.Bytes()
}

// ===== NetworkSettings =====

// NetworkSettings 包 (ID=0x7F)
type NetworkSettings struct {
	CompressionThreshold uint16
	CompressionAlgorithm byte
	EnableClientThrottle bool
	ClientThrottleBase   byte
	ClientThrottleRange  byte
}

// ===== ClientCacheStatus =====

// ClientCacheStatus 包 (ID=0xA4)
type ClientCacheStatus struct {
	Enabled bool
}

// ===== AvailableActorIdentifiers =====

// AvailableActorIdentifiers 包 (ID=0x93)
type AvailableActorIdentifiers struct {
	Identifiers string // SNBT 格式
}

// ===== ResourcePacksInfo =====

// ResourcePacksInfo 包 (ID=0x06)
type ResourcePacksInfo struct {
	MustAccept       bool
	ScriptingEnabled bool
	ForceServer      bool
	BehaviorPackCount int16
	ResourcePackCount int16
}

// ===== ResourcePackStack =====

// ResourcePackStack 包 (ID=0x07)
type ResourcePackStack struct {
	MustAccept       bool
	BehaviorPacks    []any
	ResourcePacks    []any
	GameVersion      string
	Experiments      []any
	ExperimentsSaved bool
}

// ===== ResourcePackClientResponse =====

// ResourcePackClientResponse 包 (ID=0x08)
type ResourcePackClientResponse struct {
	ResponseStatus int8
	PackID         []any
}

// ===== ServerToClientHandshake =====

// ServerToClientHandshake 包 (ID=0x03)
type ServerToClientHandshake struct {
	ServerToken string
}

// ===== ClientToServerHandshake =====

// ClientToServerHandshake 包 (ID=0x04)
type ClientToServerHandshake struct {
	ServerToken string
}

// ====== 工具函数 ======

// EncodeServerHandshake 编码服务端握手
func EncodeServerHandshake(serverToken string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(IDServerToClientHandshake)
	WriteVarintBytes(buf, uint64(len(serverToken)))
	buf.WriteString(serverToken)
	return buf.Bytes()
}

// EncodeClientHandshake 编码客户端握手
func EncodeClientHandshake(clientToken string) []byte {
	buf := new(bytes.Buffer)
	buf.WriteByte(IDClientToServerHandshake)
	WriteVarintBytes(buf, uint64(len(clientToken)))
	buf.WriteString(clientToken)
	return buf.Bytes()
}

// BuildLoginChain 构建登录链（JWT-like）
// 冰点咖啡使用从 auth server 获取的 identity_data
func BuildLoginChain(identityDataB64, chainDataB64 string) (LoginTokens, error) {
	var chainData []ChainLink
	if chainDataB64 != "" {
		// chainData 是 base64 编码的 JSON array
		chainBytes, err := base64.StdEncoding.DecodeString(chainDataB64)
		if err != nil {
			return LoginTokens{}, fmt.Errorf("decode chain: %w", err)
		}
		if err := json.Unmarshal(chainBytes, &chainData); err != nil {
			return LoginTokens{}, fmt.Errorf("parse chain: %w", err)
		}
	}

	var extra ExtraData
	if identityDataB64 != "" {
		idBytes, err := base64.StdEncoding.DecodeString(identityDataB64)
		if err != nil {
			return LoginTokens{}, fmt.Errorf("decode identity: %w", err)
		}
		if err := json.Unmarshal(idBytes, &extra); err != nil {
			return LoginTokens{}, fmt.Errorf("parse identity: %w", err)
		}
	}

	tokens := LoginTokens{
		ExtraData:   extra,
		GameVersion: fmt.Sprintf("%d.%d.%d", 1, 21, 0),
		ChainData:   chainData,
	}

	return tokens, nil
}

// ParsePublicKey 从 PEM 或 raw bytes 解析公钥
func ParsePublicKey(data []byte) ([]byte, error) {
	// 尝试作为 PKIX ASN.1 DER
	if _, err := x509.ParsePKIXPublicKey(data); err == nil {
		return data, nil
	}
	// 尝试 base64 解码
	decoded := make([]byte, base64.StdEncoding.DecodedLen(len(data)))
	n, _ := base64.StdEncoding.Decode(decoded, data)
	return decoded[:n], nil
}
