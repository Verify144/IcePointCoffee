// Package protocol 实现 Minecraft Bedrock 协议。
package protocol

import "fmt"

// TextType 文本类型
type TextType byte

const (
	TextTypeRaw TextType = iota
	TextTypeChat
	TextTypeSystem
	TextTypeWhisper
	TextTypeAnnouncement
	TextTypeUwhisper
	TextTypeRawText
	TextTypeTranslate
)

// Text 聊天/提示包
type Text struct {
	Type            TextType
	NeedsTranslation bool
	SourceName      string
	XUID            string
	Message         string
	Parameters      []string
	PlatformChatID  string
}

// DecodeText 解码 Text 包
func DecodeText(data []byte) (*Text, error) {
	if len(data) < 2 || data[0] != IDText {
		return nil, fmt.Errorf("not a Text packet")
	}
	br := NewReader(data[1:])

	text := &Text{}
	text.Type = TextType(br.ReadByte())
	text.NeedsTranslation = br.ReadBool()
	text.SourceName = br.ReadString()
	text.XUID = br.ReadString()
	text.Message = br.ReadString()

	paramCount := br.ReadVarint()
	text.Parameters = make([]string, paramCount)
	for i := 0; i < int(paramCount); i++ {
		text.Parameters[i] = br.ReadString()
	}
	text.PlatformChatID = br.ReadString()
	return text, nil
}

// EncodeText 编码 Text 包
func EncodeText(textType TextType, message string, params ...string) []byte {
	buf := NewWriter()
	buf.WriteByte(IDText)
	buf.WriteByte(byte(textType))
	buf.WriteBool(false)
	buf.WriteString("")
	buf.WriteString("")
	buf.WriteString(message)
	buf.WriteVarint(uint64(len(params)))
	for _, p := range params {
		buf.WriteString(p)
	}
	buf.WriteString("")
	return buf.Bytes()
}

// ChatMessage 聊天消息事件
type ChatMessage struct {
	Sender  string
	Message string
	Type    TextType
	Raw     string
}
