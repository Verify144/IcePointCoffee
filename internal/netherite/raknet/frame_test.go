package raknet

import (
	"bytes"
	"testing"

	"github.com/Verify144/IcePointCoffee/internal/netherite/protocol"
)

func TestFrameEncodeDecode(t *testing.T) {
	original := &Frame{
		Flags:       FlagReliable,
		ReliableIdx: 12345,
		Data:        []byte{0x01, 0x02, 0x03, 0x04, 0x05},
	}

	encoded, err := original.Encode()
	if err != nil {
		t.Fatalf("encode failed: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("encoded empty")
	}

	decoded, n, err := DecodeFrame(encoded)
	if err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if n != len(encoded) {
		t.Errorf("decoded len mismatch: got %d want %d", n, len(encoded))
	}
	if decoded.Flags != original.Flags {
		t.Errorf("flags mismatch: got %d want %d", decoded.Flags, original.Flags)
	}
	if decoded.ReliableIdx != original.ReliableIdx {
		t.Errorf("reliable idx mismatch: got %d want %d", decoded.ReliableIdx, original.ReliableIdx)
	}
	if !bytes.Equal(decoded.Data, original.Data) {
		t.Errorf("data mismatch: got %v want %v", decoded.Data, original.Data)
	}
}

func TestPingEncode(t *testing.T) {
	ping := &UnconnectedPing{
		PingTime:   1234567890,
		ClientGUID: 12345,
		ClientName: "TestClient",
	}
	encoded := ping.Encode()
	if len(encoded) < 30 {
		t.Errorf("ping too short: %d", len(encoded))
	}
	if encoded[0] != IDUnconnectedPing {
		t.Errorf("wrong packet ID: %d", encoded[0])
	}

	decoded, err := DecodePing(encoded)
	if err != nil {
		t.Fatalf("decode ping: %v", err)
	}
	if decoded.PingTime != ping.PingTime {
		t.Errorf("ping time mismatch: %d vs %d", decoded.PingTime, ping.PingTime)
	}
	if decoded.ClientName != ping.ClientName {
		t.Errorf("client name mismatch: %s vs %s", decoded.ClientName, ping.ClientName)
	}
}

func TestOpenConnectionRequest1(t *testing.T) {
	req := &OpenConnectionRequest1{
		MTU:            1200,
		ProtocolVersion: ProtocolVersion,
	}
	encoded := req.Encode()
	if len(encoded) != int(req.MTU) {
		t.Errorf("encoded len should match MTU: got %d want %d", len(encoded), req.MTU)
	}
	if encoded[0] != IDOpenConnectionRequest1 {
		t.Errorf("wrong packet ID: %d", encoded[0])
	}
}

func TestOpenConnectionRequest2(t *testing.T) {
	req := &OpenConnectionRequest2{
		MTU:       1200,
		ClientGUID: 0x12345678ABCDEF01,
	}
	encoded := req.Encode()
	if encoded[0] != IDOpenConnectionRequest2 {
		t.Errorf("wrong packet ID: %d", encoded[0])
	}
}

func TestACKEncoding(t *testing.T) {
	c := &Conn{}
	acks := []uint32{1, 5, 10, 100}
	encoded := c.encodeACK(acks)
	if len(encoded) == 0 {
		t.Fatal("encoded empty")
	}
	if encoded[0] != 0xC0 {
		t.Errorf("wrong ACK ID: %d", encoded[0])
	}
}

func TestVarintRoundTrip(t *testing.T) {
	values := []uint64{0, 1, 127, 128, 255, 16383, 16384, 65535, 4294967295}

	for _, v := range values {
		buf := new(bytes.Buffer)
		protocol.WriteVarintBytes(buf, v)

		br := bytes.NewReader(buf.Bytes())
		decoded, err := protocol.ReadVarintBR(br)
		if err != nil {
			t.Errorf("decode %d: %v", v, err)
			continue
		}
		if decoded != v {
			t.Errorf("varint roundtrip: got %d want %d", decoded, v)
		}
	}
}
