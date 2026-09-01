package protocol

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// ReadVarint 从 bytes.Reader 读取 varint
func ReadVarint(r *bytes.Reader) (uint64, error) {
	var result uint64
	for i := 0; i < 10; i++ {
		b, err := r.ReadByte()
		if err != nil {
			return 0, err
		}
		result |= uint64(b&0x7F) << (7 * i)
		if b&0x80 == 0 {
			return result, nil
		}
	}
	return 0, fmt.Errorf("varint too long")
}

// ReadVarintBR is alias for ReadVarint
var ReadVarintBR = ReadVarint

// WriteVarintBytes 写入 varint 到 bytes.Buffer
func WriteVarintBytes(buf *bytes.Buffer, v uint64) {
	for {
		b := byte(v & 0x7F)
		v >>= 7
		if v != 0 {
			b |= 0x80
		}
		buf.WriteByte(b)
		if v == 0 {
			break
		}
	}
}

// WriteVarint is alias for WriteVarintBytes
func WriteVarint(buf *bytes.Buffer, v uint64) {
	WriteVarintBytes(buf, v)
}

// ReadInt32BE 从 bytes.Reader 读取 int32 (BigEndian)
func ReadInt32BE(r *bytes.Reader) (int32, error) {
	var v int32
	if err := binary.Read(r, binary.BigEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}

// ReadUint64LE 从 bytes.Reader 读取 uint64 (LittleEndian)
func ReadUint64LE(r *bytes.Reader) (uint64, error) {
	var v uint64
	if err := binary.Read(r, binary.LittleEndian, &v); err != nil {
		return 0, err
	}
	return v, nil
}
