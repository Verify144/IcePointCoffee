package protocol

import (
	"testing"
	"time"
)

// TestEncodeMovePlayer 测试移动包编码
func TestEncodeMovePlayer(t *testing.T) {
	pk := EncodeMovePlayer(1, 100.5, 64.0, 200.3)
	if len(pk) == 0 {
		t.Error("EncodeMovePlayer returned empty packet")
	}
	if pk[0] != IDMovePlayer {
		t.Errorf("Expected packet ID %x, got %x", IDMovePlayer, pk[0])
	}
}

// TestEncodeTeleport 测试传送包编码
func TestEncodeTeleport(t *testing.T) {
	pk := EncodeTeleport(1, 0, 100, 0)
	if len(pk) == 0 {
		t.Error("EncodeTeleport returned empty packet")
	}
}

// TestEncodeSetTitle 测试设置标题编码
func TestEncodeSetTitle(t *testing.T) {
	pk := EncodeSetTitle(SetTitleSetTitle, "Hello IcePoint")
	if len(pk) == 0 {
		t.Error("EncodeSetTitle returned empty packet")
	}
	if pk[0] != IDSetTitle {
		t.Errorf("Expected packet ID %x, got %x", IDSetTitle, pk[0])
	}
}

// TestEncodeAnimate 测试动画编码
func TestEncodeAnimate(t *testing.T) {
	pk := EncodeSwingAnimation()
	if len(pk) == 0 {
		t.Error("EncodeSwingAnimation returned empty packet")
	}
	if pk[0] != IDAnimate {
		t.Errorf("Expected packet ID %x, got %x", IDAnimate, pk[0])
	}
}

// TestEncodePlayerHealth 测试设置生命值
func TestEncodePlayerHealth(t *testing.T) {
	pk := EncodePlayerHealth(20.0)
	if len(pk) == 0 {
		t.Error("EncodePlayerHealth returned empty packet")
	}
}

// TestEncodeLevelEvent 测试世界事件编码
func TestEncodeLevelEvent(t *testing.T) {
	pk := EncodeLevelEvent(1000, 10, 64, 20, 0)
	if len(pk) == 0 {
		t.Error("EncodeLevelEvent returned empty packet")
	}
}

// TestEncodeInventoryContent 测试背包内容编码
func TestEncodeInventoryContent(t *testing.T) {
	items := []ItemStack{
		{NetworkID: 1, Count: 64, Metadata: 0},
		{NetworkID: 0}, // 空气
	}
	pk := EncodeInventoryContent(ContainerInventory, items)
	if len(pk) == 0 {
		t.Error("EncodeInventoryContent returned empty packet")
	}
}

// TestEncodeContainerOpen 测试打开容器编码
func TestEncodeContainerOpen(t *testing.T) {
	pk := EncodeContainerOpen(0, 54, 100, 64, 200)
	if len(pk) == 0 {
		t.Error("EncodeContainerOpen returned empty packet")
	}
}

// TestEncodeContainerClose 测试关闭容器编码
func TestEncodeContainerClose(t *testing.T) {
	pk := EncodeContainerClose(0)
	if len(pk) == 0 {
		t.Error("EncodeContainerClose returned empty packet")
	}
}

// TestEncodeSetSpawnPosition 测试设置出生点
func TestEncodeSetSpawnPosition(t *testing.T) {
	pk := SetDefaultSpawnPosition(0, 64, 0)
	if len(pk) == 0 {
		t.Error("SetDefaultSpawnPosition returned empty packet")
	}
}

// TestVec3 测试坐标结构
func TestVec3(t *testing.T) {
	v := NewVec3(1.5, 2.5, 3.5)
	if v.X != 1.5 || v.Y != 2.5 || v.Z != 3.5 {
		t.Errorf("Vec3 mismatch: got %v", v)
	}
}

// TestReadWriteFloat64 测试 Float64 读写
func TestReadWriteFloat64(t *testing.T) {
	buf := NewWriter()
	buf.WriteFloat64(123.456)
	data := buf.Bytes()

	reader := NewReader(data)
	val := reader.ReadFloat64()
	if val != 123.456 {
		t.Errorf("Float64 mismatch: expected 123.456, got %f", val)
	}
}

// TestReadWriteInt16 测试 Int16 读写
func TestReadWriteInt16(t *testing.T) {
	buf := NewWriter()
	buf.WriteInt16(12345)
	data := buf.Bytes()

	reader := NewReader(data)
	val := reader.ReadInt16()
	if val != 12345 {
		t.Errorf("Int16 mismatch: expected 12345, got %d", val)
	}
}

// TestPerfBackoff 测试退避算法
func TestPerfBackoff(t *testing.T) {
	cfg := DefaultRetryConfig

	delay0 := cfg.Backoff(0)
	if delay0 != cfg.BaseDelay {
		t.Errorf("Backoff(0) should be BaseDelay, got %v", delay0)
	}

	delay1 := cfg.Backoff(1)
	if delay1 < cfg.BaseDelay {
		t.Errorf("Backoff(1) should be >= BaseDelay, got %v", delay1)
	}

	delay2 := cfg.Backoff(2)
	if delay2 <= delay1 {
		t.Errorf("Backoff should increase: %v <= %v", delay2, delay1)
	}
}

// TestConnectionPool 测试连接池
func TestConnectionPool(t *testing.T) {
	pool := NewConnectionPool(nil)
	if pool.Len() != 0 {
		t.Error("New pool should be empty")
	}

	pool.Put("test1", "connection1")
	if pool.Len() != 1 {
		t.Errorf("Pool should have 1 conn, got %d", pool.Len())
	}

	conn := pool.Acquire("test1")
	if conn == nil {
		t.Error("Should be able to acquire connection")
	}

	conn2 := pool.Acquire("test1")
	if conn2 != nil {
		t.Error("Should not be able to acquire already acquired connection")
	}

	pool.Release("test1")

	conn3 := pool.Acquire("test1")
	if conn3 == nil {
		t.Error("Should be able to acquire after release")
	}

	pool.Release("test1")
	removed := pool.Cleanup(0)
	if removed != 1 {
		t.Errorf("Should clean 1 conn, cleaned %d", removed)
	}
}

// TestCommandBatch 测试命令批处理
func TestCommandBatch(t *testing.T) {
	var sent []string
	mockConn := &mockCommandSender{
		sendFunc: func(cmd string) error {
			sent = append(sent, cmd)
			return nil
		},
	}

	batch := NewCommandBatch(mockConn, 10, 100*time.Millisecond)
	batch.Add("cmd1")
	batch.Add("cmd2")
	batch.Add("cmd3")

	if batch.Pending() != 3 {
		t.Errorf("Should have 3 pending, got %d", batch.Pending())
	}

	err := batch.Flush()
	if err != nil {
		t.Errorf("Flush failed: %v", err)
	}

	if len(sent) != 1 {
		t.Errorf("Should send 1 combined command, got %d", len(sent))
	}

	if sent[0] != "cmd1\ncmd2\ncmd3" {
		t.Errorf("Unexpected combined command: %s", sent[0])
	}

	batch.Stop()
}

// TestEventBuffer 测试事件缓冲
func TestEventBuffer(t *testing.T) {
	var received []Event
	batch := NewEventBuffer(3, func(events []Event) {
		received = append(received, events...)
	})

	batch.Push(Event{Type: "e1"})
	batch.Push(Event{Type: "e2"})
	batch.Push(Event{Type: "e3"})

	if len(received) != 3 {
		t.Errorf("Should receive 3 events, got %d", len(received))
	}

	batch.Push(Event{Type: "e4"})
	batch.Flush()

	if len(received) != 4 {
		t.Errorf("Should have 4 total events, got %d", len(received))
	}
}

// mockCommandSender 测试用命令发送器
type mockCommandSender struct {
	sendFunc func(cmd string) error
}

func (m *mockCommandSender) SendCommand(cmd string) error {
	return m.sendFunc(cmd)
}
