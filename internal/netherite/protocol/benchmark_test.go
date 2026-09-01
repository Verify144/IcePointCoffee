package protocol

import (
	"bytes"
	"encoding/json"
	"sync"
	"testing"
)

// ========== Writer Benchmarks ==========

func BenchmarkEncodeMovePlayer(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeMovePlayer(1, 100.5, 64.0, 200.3)
	}
}

func BenchmarkEncodeTeleport(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeTeleport(1, 0, 100, 0)
	}
}

func BenchmarkEncodeText(b *testing.B) {
	b.ReportAllocs()
	text := "Hello, IcePoint Coffee! 你好！"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeText(TextTypeChat, text)
	}
}

func BenchmarkEncodeSetTitle(b *testing.B) {
	b.ReportAllocs()
	title := "Welcome to IcePoint Coffee!"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeSetTitle(SetTitleSetTitle, title)
	}
}

func BenchmarkEncodeCommandRequest(b *testing.B) {
	b.ReportAllocs()
	cmd := "say Hello from IcePoint!"
	origin := CommandOrigin{Origin: CommandOriginPlayer}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeCommandRequest(cmd, origin, 0)
	}
}

func BenchmarkEncodeBossBar(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeBossBar(12345, "IceDragon", 0.75)
	}
}

func BenchmarkEncodePlayerListAdd(b *testing.B) {
	b.ReportAllocs()
	entries := []PlayerListEntry{
		{UUID: "player-1-uuid", EntityID: 1, Name: "Player1", XUID: "12345"},
		{UUID: "player-2-uuid", EntityID: 2, Name: "Player2", XUID: "12346"},
		{UUID: "player-3-uuid", EntityID: 3, Name: "Player3", XUID: "12347"},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodePlayerListAdd(entries)
	}
}

func BenchmarkEncodePlayerListRemove(b *testing.B) {
	b.ReportAllocs()
	uuids := []string{"uuid1", "uuid2", "uuid3", "uuid4", "uuid5"}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodePlayerListRemove(uuids)
	}
}

func BenchmarkEncodeInventoryContent(b *testing.B) {
	b.ReportAllocs()
	items := []ItemStack{
		{NetworkID: 1, Count: 64, Metadata: 0},
		{NetworkID: 2, Count: 32, Metadata: 0},
		{NetworkID: 3, Count: 16, Metadata: 0},
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeInventoryContent(ContainerInventory, items)
	}
}

func BenchmarkEncodeContainerOpen(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeContainerOpen(0, 54, 100, 64, 200)
	}
}

func BenchmarkEncodePlayerSkin(b *testing.B) {
	b.ReportAllocs()
	skin := &SkinData{
		SkinID:       "skin_id_123",
		PremiumSkin:  true,
		SkinType:     "standard_slim",
		ArmSize:      "normal",
		TrustedSkin:  true,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodePlayerSkin("uuid-1", "Player1", "skin-id", skin)
	}
}

func BenchmarkEncodeTransfer(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeTransfer("play.lobby.server.com", 19132)
	}
}

func BenchmarkEncodeMapData(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = EncodeMapData(12345, 2)
	}
}

// ========== Reader Benchmarks ==========

func BenchmarkDecodeVarint(b *testing.B) {
	testCases := []uint64{0, 127, 128, 16383, 16384, 2097151, 2097152, 268435455}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, v := range testCases {
			buf := NewWriter()
			buf.WriteVarint(v)
			r := NewReader(buf.Bytes())
			_ = r.ReadVarint()
		}
	}
}

func BenchmarkReadWriteFloat64(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewWriter()
		buf.WriteFloat64(123.456789)
		r := NewReader(buf.Bytes())
		_ = r.ReadFloat64()
	}
}

func BenchmarkReadWriteInt32(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewWriter()
		buf.WriteInt32(123456)
		r := NewReader(buf.Bytes())
		_ = r.ReadInt32()
	}
}

func BenchmarkReadWriteString(b *testing.B) {
	b.ReportAllocs()
	testStr := "Hello, IcePoint Coffee! 你好冰点咖啡！"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := NewWriter()
		buf.WriteString(testStr)
		r := NewReader(buf.Bytes())
		_ = r.ReadString()
	}
}

// ========== Buffer Pool Benchmarks ==========

func BenchmarkBufferAlloc(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bytes.NewBuffer(make([]byte, 0, 1024))
		buf.WriteString("Hello, IcePoint Coffee!")
		_ = buf.Bytes()
	}
}

var bufferPool = sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 0, 1024))
	},
}

func BenchmarkBufferPool(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := bufferPool.Get().(*bytes.Buffer)
		buf.Reset()
		buf.WriteString("Hello, IcePoint Coffee!")
		_ = buf.Bytes()
		bufferPool.Put(buf)
	}
}

// ========== JSON Benchmarks ==========

func BenchmarkJSONEncode(b *testing.B) {
	data := map[string]interface{}{
		"type":    "command",
		"success": true,
		"count":   123,
		"message": "Hello, IcePoint!",
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf := &bytes.Buffer{}
		enc := json.NewEncoder(buf)
		enc.Encode(data)
	}
}

func BenchmarkJSONDecode(b *testing.B) {
	raw := []byte(`{"type":"command","success":true,"count":123,"message":"Hello"}`)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var m map[string]interface{}
		json.Unmarshal(raw, &m)
	}
}

// ========== Command Batch Benchmarks ==========

func BenchmarkCommandBatchFlush(b *testing.B) {
	mock := &mockCmdSender{}
	batch := NewCommandBatch(mock, 100, 0)
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		batch.Add("say Hello")
		batch.Add("say World")
		batch.Add("say IcePoint")
		batch.Flush()
	}
}

type mockCmdSender struct{}

func (m *mockCmdSender) SendCommand(cmd string) error {
	return nil
}

// ========== Backoff Benchmarks ==========

func BenchmarkBackoff(b *testing.B) {
	cfg := DefaultRetryConfig
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for attempt := 0; attempt < 5; attempt++ {
			_ = cfg.Backoff(attempt)
		}
	}
}

// ========== Connection Pool Benchmarks ==========

func BenchmarkConnectionPoolAcquireRelease(b *testing.B) {
	pool := NewConnectionPool(nil)
	pool.Put("test", "connection")
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		conn := pool.Acquire("test")
		pool.Release("test")
		_ = conn
	}
}

func BenchmarkConnectionPoolCleanup(b *testing.B) {
	pool := NewConnectionPool(nil)
	for i := 0; i < 100; i++ {
		pool.Put(string(rune(i)), "conn")
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pool.Cleanup(0)
	}
}

// ========== Event Buffer Benchmarks ==========

func BenchmarkEventBufferPush(b *testing.B) {
	buf := NewEventBuffer(1000, func(evts []Event) {
		_ = evts
	})
	event := Event{Type: "test", Data: "data"}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		buf.Push(event)
	}
}

// ========== Concurrent Benchmarks ==========

func BenchmarkConcurrentWrites(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		buf := NewWriter()
		for pb.Next() {
			buf.WriteString("test")
			buf.WriteVarint(123)
			_ = buf.Bytes()
			buf.buf.Reset()
		}
	})
}

func BenchmarkConcurrentVarintWrites(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		buf := NewWriter()
		for pb.Next() {
			buf.WriteVarint(uint64(b.N % 1000000))
			_ = buf.Bytes()
			buf.buf.Reset()
		}
	})
}
