package protocol

import (
	"bytes"
	"encoding/binary"
)

// Reader 用于解析 MC 包
type Reader struct {
	buf *bytes.Reader
}

// NewReader 创建 Reader
func NewReader(data []byte) *Reader {
	return &Reader{buf: bytes.NewReader(data)}
}

// ReadByte 读取单字节
func (r *Reader) ReadByte() byte {
	b, _ := r.buf.ReadByte()
	return b
}

// ReadBytes 读取定长字节
func (r *Reader) ReadBytes(n int) []byte {
	data := make([]byte, n)
	r.buf.Read(data)
	return data
}

// Read 读取字节（io.Reader 接口）
func (r *Reader) Read(p []byte) (int, error) {
	return r.buf.Read(p)
}

// ReadBool 读取布尔
func (r *Reader) ReadBool() bool {
	return r.ReadByte() != 0
}

// ReadString 读取 varint-prefixed 字符串
func (r *Reader) ReadString() string {
	n := r.ReadVarint()
	data := make([]byte, n)
	r.buf.Read(data)
	return string(data)
}

// ReadInt16 读取 int16 (LE)
func (r *Reader) ReadInt16() int16 {
	var v int16
	binary.Read(r.buf, binary.LittleEndian, &v)
	return v
}

// ReadInt32 读取 int32 (LE)
func (r *Reader) ReadInt32() int32 {
	var v int32
	binary.Read(r.buf, binary.LittleEndian, &v)
	return v
}

// ReadInt64 读取 int64 (LE)
func (r *Reader) ReadInt64() int64 {
	var v int64
	binary.Read(r.buf, binary.LittleEndian, &v)
	return v
}

// ReadUint16 读取 uint16 (LE)
func (r *Reader) ReadUint16() uint16 {
	var v uint16
	binary.Read(r.buf, binary.LittleEndian, &v)
	return v
}

// ReadUint32 读取 uint32 (LE)
func (r *Reader) ReadUint32() uint32 {
	var v uint32
	binary.Read(r.buf, binary.LittleEndian, &v)
	return v
}

// ReadFloat32 读取 float32
func (r *Reader) ReadFloat32() float32 {
	var v float32
	binary.Read(r.buf, binary.LittleEndian, &v)
	return v
}

// ReadFloat64 读取 float64
func (r *Reader) ReadFloat64() float64 {
	var v float64
	binary.Read(r.buf, binary.LittleEndian, &v)
	return v
}

// ReadVarint 读取 varint
func (r *Reader) ReadVarint() uint64 {
	var result uint64
	for i := 0; i < 10; i++ {
		b, err := r.buf.ReadByte()
		if err != nil {
			return 0
		}
		result |= uint64(b&0x7F) << (7 * i)
		if b&0x80 == 0 {
			return result
		}
	}
	return result
}

// Writer 用于构造 MC 包
type Writer struct {
	buf *bytes.Buffer
}

// NewWriter 创建 Writer
func NewWriter() *Writer {
	return &Writer{buf: &bytes.Buffer{}}
}

// Bytes 返回已写入数据
func (w *Writer) Bytes() []byte {
	return w.buf.Bytes()
}

// WriteByte 写入单字节
func (w *Writer) WriteByte(b byte) {
	w.buf.WriteByte(b)
}

// WriteBool 写入布尔
func (w *Writer) WriteBool(v bool) {
	if v {
		w.buf.WriteByte(1)
	} else {
		w.buf.WriteByte(0)
	}
}

// Write 写入字节
func (w *Writer) Write(p []byte) (int, error) {
	return w.buf.Write(p)
}

// WriteBytes 写入字节
func (w *Writer) WriteBytes(data []byte) {
	w.buf.Write(data)
}

// WriteString 写入 varint-prefixed 字符串
func (w *Writer) WriteString(s string) {
	WriteVarintBytes(w.buf, uint64(len(s)))
	w.buf.WriteString(s)
}

// WriteVarint 写入 varint
func (w *Writer) WriteVarint(v uint64) {
	WriteVarintBytes(w.buf, v)
}

// WriteInt16 写入 int16 (LE)
func (w *Writer) WriteInt16(v int16) {
	binary.Write(w.buf, binary.LittleEndian, v)
}

// WriteInt32 写入 int32 (LE)
func (w *Writer) WriteInt32(v int32) {
	binary.Write(w.buf, binary.LittleEndian, v)
}

// WriteInt64 写入 int64 (LE)
func (w *Writer) WriteInt64(v int64) {
	binary.Write(w.buf, binary.LittleEndian, v)
}

// WriteUint16 写入 uint16 (LE)
func (w *Writer) WriteUint16(v uint16) {
	binary.Write(w.buf, binary.LittleEndian, v)
}

// WriteUint32 写入 uint32 (LE)
func (w *Writer) WriteUint32(v uint32) {
	binary.Write(w.buf, binary.LittleEndian, v)
}

// WriteFloat32 写入 float32
func (w *Writer) WriteFloat32(v float32) {
	binary.Write(w.buf, binary.LittleEndian, v)
}

// WriteFloat64 写入 float64
func (w *Writer) WriteFloat64(v float64) {
	binary.Write(w.buf, binary.LittleEndian, v)
}

// EncodeClientCacheStatus 编码 ClientCacheStatus 包
func EncodeClientCacheStatus(enabled bool) []byte {
	buf := NewWriter()
	buf.WriteByte(IDClientCacheStatus)
	buf.WriteBool(enabled)
	return buf.Bytes()
}

// EncodePing 编码 Ping 包（0x00）
func EncodePing(pingID int32) []byte {
	buf := NewWriter()
	buf.WriteByte(0x00)
	buf.WriteInt32(pingID)
	return buf.Bytes()
}

// EncodePong 编码 Pong 包（0x03）
func EncodePong(pingID int32) []byte {
	buf := NewWriter()
	buf.WriteByte(0x03)
	buf.WriteInt32(pingID)
	return buf.Bytes()
}

// EncodeRequestPermissions 编码权限请求包
func EncodeRequestPermissions(perm int32) []byte {
	buf := NewWriter()
	buf.WriteByte(IDRequestPermissions)
	buf.WriteInt32(perm)
	buf.WriteString("IcePointCoffee")
	return buf.Bytes()
}

// EncodePlayerAction 编码玩家操作包
func EncodePlayerAction(runtimeID int64, action int32) []byte {
	buf := NewWriter()
	buf.WriteByte(IDPlayerAction)
	buf.WriteUint32(uint32(runtimeID))
	buf.WriteInt32(action)
	buf.WriteInt32(0) // pos x
	buf.WriteInt32(0) // pos y
	buf.WriteInt32(0) // pos z
	buf.WriteInt32(0) // result pos x
	buf.WriteInt32(0) // result pos y
	buf.WriteInt32(0) // result pos z
	buf.WriteInt32(0) // face
	return buf.Bytes()
}

// EncodeAdventure 编码冒险设置包
func EncodeAdventure() []byte {
	buf := NewWriter()
	buf.WriteByte(IDAdventure)
	buf.WriteInt32(0) // flags
	buf.WriteInt32(0) // world flags
	buf.WriteInt32(0) // permissions
	buf.WriteInt32(-1) // 命令权限
	buf.WriteInt32(0) // 其他标志
	buf.WriteInt32(0) // player flags
	buf.WriteInt64(0) // actor unique id
	return buf.Bytes()
}

// EncodeSetTime 编码 SetTime 包
func EncodeSetTime(time int32) []byte {
	buf := NewWriter()
	buf.WriteByte(IDSetTime)
	buf.WriteInt32(time)
	return buf.Bytes()
}
