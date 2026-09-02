package protocol

import (
	"testing"
)

func TestEncodeInteract(t *testing.T) {
	data := EncodeInteract(InteractActionLeftClick, 123, 10.0, 64.0, 20.0)
	if len(data) == 0 {
		t.Fatal("EncodeInteract returned empty data")
	}
	if data[0] != IDInteract {
		t.Errorf("Expected packet ID %x, got %x", IDInteract, data[0])
	}
}

func TestDecodeInteract(t *testing.T) {
	data := EncodeInteract(InteractActionRightClick, 456, 1.5, 70.0, 3.0)
	pk := DecodeInteract(data)
	if pk.Action != InteractActionRightClick {
		t.Errorf("Expected action %d, got %d", InteractActionRightClick, pk.Action)
	}
	if pk.TargetID != 456 {
		t.Errorf("Expected target ID 456, got %d", pk.TargetID)
	}
}

func TestEncodeRespawn(t *testing.T) {
	data := EncodeRespawn(0.0, 65.0, 0.0)
	if len(data) == 0 {
		t.Fatal("EncodeRespawn returned empty data")
	}
	if data[0] != IDRespawn {
		t.Errorf("Expected packet ID %x, got %x", IDRespawn, data[0])
	}
}

func TestEncodeMobEquipment(t *testing.T) {
	item := ItemStack{NetworkID: 271, Count: 1}
	data := EncodeMobEquipment(1, item, 0, 0)
	if len(data) == 0 {
		t.Fatal("EncodeMobEquipment returned empty data")
	}
	if data[0] != IDMobEquipment {
		t.Errorf("Expected packet ID %x, got %x", IDMobEquipment, data[0])
	}
}

func TestEncodePlaySound(t *testing.T) {
	data := EncodePlaySound(SoundExplode, 0, 64, 0, 1.0, 1.0)
	if len(data) == 0 {
		t.Fatal("EncodePlaySound returned empty data")
	}
	if data[0] != IDPlaySound {
		t.Errorf("Expected packet ID %x, got %x", IDPlaySound, data[0])
	}
}

func TestEncodeStopSound(t *testing.T) {
	data := EncodeStopSound(SoundAmbientCave, false)
	if len(data) == 0 {
		t.Fatal("EncodeStopSound returned empty data")
	}
	if data[0] != IDStopSound {
		t.Errorf("Expected packet ID %x, got %x", IDStopSound, data[0])
	}
}

func TestEncodeSetDefaultGamemode(t *testing.T) {
	data := EncodeSetDefaultGamemode(0) // survival
	if len(data) == 0 {
		t.Fatal("EncodeSetDefaultGamemode returned empty data")
	}
	if data[0] != IDSetDefaultGamemode {
		t.Errorf("Expected packet ID %x, got %x", IDSetDefaultGamemode, data[0])
	}
}

func BenchmarkEncodeInteract(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeInteract(InteractActionLeftClick, uint64(i), 10.0, 64.0, 20.0)
	}
}

func BenchmarkEncodeRespawn(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeRespawn(0.0, 65.0, 0.0)
	}
}

func BenchmarkEncodePlaySound(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodePlaySound(SoundLevelUp, 0, 64, 0, 1.0, 1.0)
	}
}
