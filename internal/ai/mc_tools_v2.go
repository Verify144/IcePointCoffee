package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Verify144/IcePointCoffee/internal/mc"
)

// =====================================================================
// Phase 12D: 更多 MC 协议包工具（库存/天气/世界管理/实体数据）
// =====================================================================

// MCWeatherTool 天气控制
type MCWeatherTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCWeatherTool() *MCWeatherTool { return &MCWeatherTool{} }
func (t *MCWeatherTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCWeatherTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCWeatherTool) Name() string        { return "mc_weather" }
func (t *MCWeatherTool) Description() string { return "设置天气（clear=晴天/rain=下雨/thunder=雷雨）" }
func (t *MCWeatherTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"weather": map[string]interface{}{
				"type": "string",
				"enum":  []string{"clear", "rain", "thunder"},
				"description": "天气类型：clear=晴天, rain=下雨, thunder=雷雨",
			},
		},
		"required": []string{"weather"},
	}
}
func (t *MCWeatherTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Weather string `json:"weather"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().SetWeather(params.Weather); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "weather": params.Weather}, nil
}

// MCWorldBorderTool 世界边界
type MCWorldBorderTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCWorldBorderTool() *MCWorldBorderTool { return &MCWorldBorderTool{} }
func (t *MCWorldBorderTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCWorldBorderTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCWorldBorderTool) Name() string        { return "mc_worldborder" }
func (t *MCWorldBorderTool) Description() string { return "设置世界边界中心点和半径（单位：方块）" }
func (t *MCWorldBorderTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"center_x": map[string]interface{}{"type": "number", "description": "边界中心 X 坐标"},
			"center_z": map[string]interface{}{"type": "number", "description": "边界中心 Z 坐标"},
			"radius":   map[string]interface{}{"type": "number", "description": "边界半径（方块）"},
		},
		"required": []string{"radius"},
	}
}
func (t *MCWorldBorderTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		CenterX float64 `json:"center_x"`
		CenterZ float64 `json:"center_z"`
		Radius  float64 `json:"radius"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if params.Radius <= 0 {
		return map[string]interface{}{"success": false, "error": "半径必须 > 0"}, nil
	}
	if params.CenterX == 0 && params.CenterZ == 0 {
		params.CenterX, params.CenterZ = 0, 0
	}
	if err := t.getClient().SetWorldBorder(params.CenterX, params.CenterZ, params.Radius); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success":  true,
		"center":    map[string]float64{"x": params.CenterX, "z": params.CenterZ},
		"radius":    params.Radius,
	}, nil
}

// MCSpawnpointTool 设置重生点
type MCSpawnpointTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCSpawnpointTool() *MCSpawnpointTool { return &MCSpawnpointTool{} }
func (t *MCSpawnpointTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCSpawnpointTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCSpawnpointTool) Name() string        { return "mc_spawnpoint" }
func (t *MCSpawnpointTool) Description() string { return "设置世界出生点" }
func (t *MCSpawnpointTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"x": map[string]interface{}{"type": "number"},
			"y": map[string]interface{}{"type": "number"},
			"z": map[string]interface{}{"type": "number"},
		},
		"required": []string{"x", "y", "z"},
	}
}
func (t *MCSpawnpointTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().SetSpawnpoint(params.X, params.Y, params.Z); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success":  true,
		"spawn":    map[string]float64{"x": params.X, "y": params.Y, "z": params.Z},
	}, nil
}

// MCDifficultyTool 难度设置
type MCDifficultyTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCDifficultyTool() *MCDifficultyTool { return &MCDifficultyTool{} }
func (t *MCDifficultyTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCDifficultyTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCDifficultyTool) Name() string        { return "mc_difficulty" }
func (t *MCDifficultyTool) Description() string { return "设置游戏难度（peaceful/easy/normal/hard）" }
func (t *MCDifficultyTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"difficulty": map[string]interface{}{
				"type": "string",
				"enum":  []string{"peaceful", "easy", "normal", "hard"},
			},
		},
		"required": []string{"difficulty"},
	}
}
func (t *MCDifficultyTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Difficulty string `json:"difficulty"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().SetDifficulty(params.Difficulty); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "difficulty": params.Difficulty}, nil
}

// MCNametagTool 命名牌
type MCNametagTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCNametagTool() *MCNametagTool { return &MCNametagTool{} }
func (t *MCNametagTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCNametagTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCNametagTool) Name() string        { return "mc_nametag" }
func (t *MCNametagTool) Description() string { return "给生物设置命名牌（显示自定义名称）" }
func (t *MCNametagTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target":   map[string]interface{}{"type": "string", "description": "生物类型，如 zombie/pig/cow/creeper"},
			"nametag":  map[string]interface{}{"type": "string", "description": "显示的名称"},
		},
		"required": []string{"target", "nametag"},
	}
}
func (t *MCNametagTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Target  string `json:"target"`
		Nametag string `json:"nametag"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().SetNametag(params.Target, params.Nametag); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success":  true,
		"target":   params.Target,
		"nametag":  params.Nametag,
	}, nil
}

// MCXPTool 经验值
type MCXPTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCXPTool() *MCXPTool { return &MCXPTool{} }
func (t *MCXPTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCXPTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCXPTool) Name() string        { return "mc_xp" }
func (t *MCXPTool) Description() string { return "给玩家添加经验值（正数）或设置等级（负数，如 -30L）" }
func (t *MCXPTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{"type": "string", "description": "目标玩家（@s=自己）", "default": "@s"},
			"amount": map[string]interface{}{"type": "integer", "description": "经验值点数（正数）或等级（负数，如 -30L）"},
		},
		"required": []string{"amount"},
	}
}
func (t *MCXPTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Target string `json:"target"`
		Amount int    `json:"amount"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if params.Target == "" {
		params.Target = "@s"
	}
	if err := t.getClient().GiveXP(params.Target, params.Amount); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success": true,
		"target":  params.Target,
		"amount":  params.Amount,
	}, nil
}

// MCEffectTool 药水效果
type MCEffectTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCEffectTool() *MCEffectTool { return &MCEffectTool{} }
func (t *MCEffectTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCEffectTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCEffectTool) Name() string { return "mc_effect" }
func (t *MCEffectTool) Description() string {
	return "给玩家施加药水效果。常用: speed/slowness/strength/jump_boost/regeneration/fire_resistance/invisibility/night_vision/haste"
}
func (t *MCEffectTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target":   map[string]interface{}{"type": "string", "description": "目标玩家", "default": "@s"},
			"effect":   map[string]interface{}{"type": "string", "description": "药水效果名称"},
			"seconds":  map[string]interface{}{"type": "integer", "description": "持续秒数", "default": 30},
		},
		"required": []string{"effect"},
	}
}
func (t *MCEffectTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Target  string `json:"target"`
		Effect  string `json:"effect"`
		Seconds int    `json:"seconds"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if params.Target == "" {
		params.Target = "@s"
	}
	if params.Seconds == 0 {
		params.Seconds = 30
	}
	if err := t.getClient().ApplyEffect(params.Target, params.Effect, params.Seconds); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success":  true,
		"target":   params.Target,
		"effect":   params.Effect,
		"seconds":  params.Seconds,
	}, nil
}

// MCClearInventoryTool 清空物品栏
type MCClearInventoryTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCClearInventoryTool() *MCClearInventoryTool { return &MCClearInventoryTool{} }
func (t *MCClearInventoryTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCClearInventoryTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCClearInventoryTool) Name() string        { return "mc_clear_inventory" }
func (t *MCClearInventoryTool) Description() string { return "清空指定玩家的物品栏" }
func (t *MCClearInventoryTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target": map[string]interface{}{"type": "string", "description": "目标玩家", "default": "@s"},
		},
	}
}
func (t *MCClearInventoryTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Target string `json:"target"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if params.Target == "" {
		params.Target = "@s"
	}
	if err := t.getClient().ClearInventory(params.Target); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "target": params.Target}, nil
}

// 编译期检查
var (
	_ Tool = (*MCWeatherTool)(nil)
	_ Tool = (*MCWorldBorderTool)(nil)
	_ Tool = (*MCSpawnpointTool)(nil)
	_ Tool = (*MCDifficultyTool)(nil)
	_ Tool = (*MCNametagTool)(nil)
	_ Tool = (*MCXPTool)(nil)
	_ Tool = (*MCEffectTool)(nil)
	_ Tool = (*MCClearInventoryTool)(nil)
)

// 占位防编译警告
var _ = fmt.Sprintf
