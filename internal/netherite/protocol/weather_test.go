package protocol

import "testing"

func TestEncodeWeatherPacket(t *testing.T) {
	data := EncodeWeatherPacket(WeatherRain, 0, 64, 0)
	if len(data) == 0 {
		t.Fatal("EncodeWeatherPacket returned empty")
	}
	// 走 EncodeLevelEvent (包 ID=0x18)
}

func TestEncodeWorldBorderSize(t *testing.T) {
	data := EncodeWorldBorderSize(10000000)
	if len(data) == 0 {
		t.Fatal("EncodeWorldBorderSize returned empty")
	}
	if data[0] != IDWorldBorder {
		t.Errorf("Expected packet ID %x, got %x", IDWorldBorder, data[0])
	}
}

func TestEncodeWorldBorderCenter(t *testing.T) {
	data := EncodeWorldBorderCenter(0, 0)
	if len(data) == 0 {
		t.Fatal("EncodeWorldBorderCenter returned empty")
	}
}

func TestEncodeWorldBorderLerp(t *testing.T) {
	data := EncodeWorldBorderLerp(1000000, 10000000, 100000)
	if len(data) == 0 {
		t.Fatal("EncodeWorldBorderLerp returned empty")
	}
}

func TestEncodeSetDifficulty(t *testing.T) {
	tests := []struct {
		d     Difficulty
		label string
	}{
		{DifficultyPeaceful, "peaceful"},
		{DifficultyEasy, "easy"},
		{DifficultyNormal, "normal"},
		{DifficultyHard, "hard"},
	}
	for _, tt := range tests {
		data := EncodeSetDifficulty(tt.d)
		if len(data) == 0 {
			t.Errorf("EncodeSetDifficulty(%s) returned empty", tt.label)
		}
	}
}

func TestEncodeSpawnPositionPacket(t *testing.T) {
	data := EncodeSpawnPositionPacket(0, 64, 0, 0)
	if len(data) == 0 {
		t.Fatal("EncodeSpawnPositionPacket returned empty")
	}
	if data[0] != IDSpawnPosition {
		t.Errorf("Expected packet ID %x, got %x", IDSpawnPosition, data[0])
	}
}

func TestEncodeUpdateAbilitiesPacket(t *testing.T) {
	data := EncodeUpdateAbilitiesPacket(false)
	if len(data) == 0 {
		t.Fatal("EncodeUpdateAbilitiesPacket returned empty")
	}
	data2 := EncodeUpdateAbilitiesPacket(true)
	if len(data2) == 0 {
		t.Fatal("EncodeUpdateAbilitiesPacket(true) returned empty")
	}
}

func TestEncodeTickSync(t *testing.T) {
	data := EncodeTickSync(12345)
	if len(data) == 0 {
		t.Fatal("EncodeTickSync returned empty")
	}
	if data[0] != IDTickSync {
		t.Errorf("Expected packet ID %x, got %x", IDTickSync, data[0])
	}
}

func TestEncodeScriptMessage(t *testing.T) {
	data := EncodeScriptMessage("hello from IcePoint")
	if len(data) == 0 {
		t.Fatal("EncodeScriptMessage returned empty")
	}
	if data[0] != IDScriptMessage {
		t.Errorf("Expected packet ID %x, got %x", IDScriptMessage, data[0])
	}
}

func BenchmarkEncodeWeatherPacket(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeWeatherPacket(WeatherRain, 0, 64, 0)
	}
}

func BenchmarkEncodeWorldBorderSize(b *testing.B) {
	for i := 0; i < b.N; i++ {
		EncodeWorldBorderSize(10000000)
	}
}
