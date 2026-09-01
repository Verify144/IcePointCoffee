package protocol

import (
	"bytes"
	"testing"
)

func TestVarintRoundTrip(t *testing.T) {
	values := []uint64{0, 1, 127, 128, 255, 1000, 16383, 16384, 65535, 100000, 1<<20, 1<<30, 1<<31 - 1}
	for _, v := range values {
		buf := new(bytes.Buffer)
		WriteVarintBytes(buf, v)
		decoded, err := ReadVarintBR(bytes.NewReader(buf.Bytes()))
		if err != nil {
			t.Errorf("read varint %d: %v", v, err)
			continue
		}
		if decoded != v {
			t.Errorf("varint roundtrip %d: got %d", v, decoded)
		}
	}
}

func TestVarintMaxBytes(t *testing.T) {
	// varint encoding should never exceed 10 bytes for int64
	buf := new(bytes.Buffer)
	WriteVarintBytes(buf, 1<<62)
	if buf.Len() > 10 {
		t.Errorf("varint too long: %d bytes", buf.Len())
	}
}

func TestVarintBoundary(t *testing.T) {
	// Test max values for 7-bit, 14-bit, 21-bit etc
	tests := []struct {
		val      uint64
		wantBits int
	}{
		{0, 1},
		{0x7F, 1},
		{0x80, 2},
		{0x3FFF, 2},
		{0x4000, 3},
		{0x1FFFFF, 3},
		{0x200000, 4},
	}
	for _, tc := range tests {
		buf := new(bytes.Buffer)
		WriteVarintBytes(buf, tc.val)
		gotBits := buf.Len() * 7
		if gotBits < tc.wantBits*7-(7) && gotBits > tc.wantBits*7 {
			t.Errorf("varint bits for %d: expected around %d, got %d", tc.val, tc.wantBits, gotBits)
		}
	}
}

func TestWriterString(t *testing.T) {
	w := NewWriter()
	w.WriteString("hello")
	w.WriteByte(0)
	got := w.Bytes()
	want := []byte{5, 'h', 'e', 'l', 'l', 'o', 0}
	if !bytes.Equal(got, want) {
		t.Errorf("WriteString: got %v want %v", got, want)
	}
}

func TestReaderBasic(t *testing.T) {
	data := []byte{0x05, 0x01, 0x02, 0x03, 0x04, 0x05, 0x08, 0x00, 0x00}
	r := NewReader(data)

	// ReadByte
	b := r.ReadByte()
	if b != 0x05 {
		t.Errorf("ReadByte: got 0x%02x want 0x05", b)
	}

	// ReadBytes
	bs := r.ReadBytes(4)
	want := []byte{0x01, 0x02, 0x03, 0x04}
	if !bytes.Equal(bs, want) {
		t.Errorf("ReadBytes: got %v want %v", bs, want)
	}

	// ReadBool
	r.buf.Seek(6, 0)
	b = r.ReadByte()
	if b != 0x08 {
		t.Errorf("ReadByte for bool: got 0x%02x want 0x08", b)
	}

	_ = r.ReadUint16()
}

func TestPyRpcEncodeDecode(t *testing.T) {
	method := "TestMethod"
	args := []any{"arg1", 123, true}

	encoded := EncodePyRpc(method, args, PyRpcOperationTypeSend)
	if len(encoded) == 0 {
		t.Fatal("encoded empty")
	}
	if encoded[0] != IDPyRpc {
		t.Errorf("wrong packet ID: %d", encoded[0])
	}

	decoded, err := DecodePyRpc(encoded)
	if err != nil {
		t.Fatalf("decode PyRpc: %v", err)
	}
	if len(decoded.Value) < 1 {
		t.Fatal("value too short")
	}
	gotMethod, ok := decoded.Value[0].(string)
	if !ok || gotMethod != method {
		t.Errorf("method mismatch: got %v want %s", gotMethod, method)
	}
}

func TestNeteaseJsonEncode(t *testing.T) {
	encoded := EncodeNeteaseJson("LOGIN_UID", "", "12345")
	if len(encoded) == 0 {
		t.Fatal("encoded empty")
	}
	if encoded[0] != IDNeteaseJson {
		t.Errorf("wrong packet ID: %d", encoded[0])
	}
}
