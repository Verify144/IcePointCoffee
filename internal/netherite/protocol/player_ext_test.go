package protocol

import "testing"

// TestEncodePlayerListAdd 测试添加玩家
func TestEncodePlayerListAdd(t *testing.T) {
	entries := []PlayerListEntry{
		{UUID: "uuid-1", EntityID: 1, Name: "Alice", XUID: "111"},
		{UUID: "uuid-2", EntityID: 2, Name: "Bob", XUID: "222"},
	}
	pk := EncodePlayerListAdd(entries)
	if len(pk) == 0 {
		t.Error("EncodePlayerListAdd returned empty")
	}
	if pk[0] != IDPlayerList {
		t.Errorf("Expected packet ID %x, got %x", IDPlayerList, pk[0])
	}
}

// TestEncodePlayerListRemove 测试移除玩家
func TestEncodePlayerListRemove(t *testing.T) {
	uuids := []string{"uuid-1", "uuid-2", "uuid-3"}
	pk := EncodePlayerListRemove(uuids)
	if len(pk) == 0 {
		t.Error("EncodePlayerListRemove returned empty")
	}
}

// TestEncodePlayerSkin 测试玩家皮肤
func TestEncodePlayerSkin(t *testing.T) {
	skin := &SkinData{
		SkinID:      "test_skin",
		PremiumSkin: true,
		SkinType:    "slim",
		ArmSize:     "normal",
		TrustedSkin: true,
	}
	pk := EncodePlayerSkin("uuid", "Test", "skin-id", skin)
	if len(pk) == 0 {
		t.Error("EncodePlayerSkin returned empty")
	}
	if pk[0] != IDPlayerSkin {
		t.Errorf("Expected packet ID %x, got %x", IDPlayerSkin, pk[0])
	}
}

// TestEncodeBossBar 测试 Boss 血条
func TestEncodeBossBar(t *testing.T) {
	pk := EncodeBossBar(12345, "IceDragon", 0.75)
	if len(pk) == 0 {
		t.Error("EncodeBossBar returned empty")
	}
}

// TestEncodeMapData 测试地图数据
func TestEncodeMapData(t *testing.T) {
	pk := EncodeMapData(12345, 2)
	if len(pk) == 0 {
		t.Error("EncodeMapData returned empty")
	}
}

// TestEncodeTransfer 测试服务器传送
func TestEncodeTransfer(t *testing.T) {
	pk := EncodeTransfer("play.server.com", 19132)
	if len(pk) == 0 {
		t.Error("EncodeTransfer returned empty")
	}
}

// TestEncodeGameRules 测试游戏规则
func TestEncodeGameRules(t *testing.T) {
	rules := map[string]bool{
		"doDaylightCycle": true,
		"keepInventory":   false,
		"mobGriefing":     true,
	}
	pk := EncodeGameRules(rules)
	if len(pk) == 0 {
		t.Error("EncodeGameRules returned empty")
	}
}

// TestEncodePlayerEnchantOptions 测试附魔选项
func TestEncodePlayerEnchantOptions(t *testing.T) {
	options := []EnchantOption{
		{
			Cost: 5,
			Seed: 12345,
			Enchants: []Enchantment{
				{ID: 0, Level: 4}, // 锋利 IV
				{ID: 1, Level: 3}, // 击退 III
			},
		},
	}
	pk := EncodePlayerEnchantOptions(options)
	if len(pk) == 0 {
		t.Error("EncodePlayerEnchantOptions returned empty")
	}
}

// TestEncodeWeather 测试天气
func TestEncodeWeather(t *testing.T) {
	pk := EncodeWeather(1.0, 0.5)
	if len(pk) == 0 {
		t.Error("EncodeWeather returned empty")
	}
}

// TestEncodeCameraPresets 测试相机预设
func TestEncodeCameraPresets(t *testing.T) {
	pk := EncodeCameraPresets()
	if len(pk) == 0 {
		t.Error("EncodeCameraPresets returned empty")
	}
}
