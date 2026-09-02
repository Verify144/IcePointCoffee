package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/Verify144/IcePointCoffee/internal/mc"
)

// =====================================================================
// Phase 12: 更多 MC 协议包工具
// =====================================================================

// MCAttackTool 攻击实体
type MCAttackTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCAttackTool() *MCAttackTool { return &MCAttackTool{} }
func (t *MCAttackTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCAttackTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCAttackTool) Name() string        { return "mc_attack" }
func (t *MCAttackTool) Description() string { return "攻击指定实体（用实体 runtime ID）" }
func (t *MCAttackTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target_id": map[string]interface{}{"type": "integer", "description": "目标实体的 runtime ID"},
		},
		"required": []string{"target_id"},
	}
}
func (t *MCAttackTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		TargetID uint64 `json:"target_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().AttackEntity(params.TargetID); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "attacked": params.TargetID}, nil
}

// MCSpawnEntityTool 生成生物
type MCSpawnEntityTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCSpawnEntityTool() *MCSpawnEntityTool { return &MCSpawnEntityTool{} }
func (t *MCSpawnEntityTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCSpawnEntityTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCSpawnEntityTool) Name() string { return "mc_spawn_entity" }
func (t *MCSpawnEntityTool) Description() string {
	return "在指定坐标召唤生物，如 zombie/pig/cow/creeper"
}
func (t *MCSpawnEntityTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"entity_type": map[string]interface{}{"type": "string", "description": "生物类型，如 zombie/pig/cow"},
			"name":        map[string]interface{}{"type": "string", "description": "自定义名称（可选）"},
			"x":           map[string]interface{}{"type": "number"},
			"y":           map[string]interface{}{"type": "number"},
			"z":           map[string]interface{}{"type": "number"},
		},
		"required": []string{"entity_type", "x", "y", "z"},
	}
}
func (t *MCSpawnEntityTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		EntityType string  `json:"entity_type"`
		Name       string  `json:"name"`
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Z          float64 `json:"z"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().SpawnEntity(params.EntityType, params.Name, params.X, params.Y, params.Z); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success":     true,
		"entity_type": params.EntityType,
		"name":        params.Name,
		"position":    map[string]float64{"x": params.X, "y": params.Y, "z": params.Z},
	}, nil
}

// MCRemoveEntityTool 移除实体
type MCRemoveEntityTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCRemoveEntityTool() *MCRemoveEntityTool { return &MCRemoveEntityTool{} }
func (t *MCRemoveEntityTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCRemoveEntityTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCRemoveEntityTool) Name() string        { return "mc_remove_entity" }
func (t *MCRemoveEntityTool) Description() string { return "删除全部非玩家实体" }
func (t *MCRemoveEntityTool) Parameters() map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
}
func (t *MCRemoveEntityTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	if err := t.getClient().RemoveEntity(0); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "message": "已移除全部非玩家实体"}, nil
}

// MCPlaySoundTool 播放声音
type MCPlaySoundTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCPlaySoundTool() *MCPlaySoundTool { return &MCPlaySoundTool{} }
func (t *MCPlaySoundTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCPlaySoundTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCPlaySoundTool) Name() string { return "mc_play_sound" }
func (t *MCPlaySoundTool) Description() string {
	return "在指定位置播放声音。常用 sound_id: random.levelup/explode/mob.zombie.say/ambient.weather.thunder"
}
func (t *MCPlaySoundTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sound_id": map[string]interface{}{"type": "string", "description": "声音 ID"},
			"x":        map[string]interface{}{"type": "number"},
			"y":        map[string]interface{}{"type": "number"},
			"z":        map[string]interface{}{"type": "number"},
			"volume":   map[string]interface{}{"type": "number", "description": "音量 0.0-1.0", "default": 1.0},
			"pitch":    map[string]interface{}{"type": "number", "description": "音高 0.0-2.0", "default": 1.0},
		},
		"required": []string{"sound_id", "x", "y", "z"},
	}
}
func (t *MCPlaySoundTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		SoundID string  `json:"sound_id"`
		X       float64 `json:"x"`
		Y       float64 `json:"y"`
		Z       float64 `json:"z"`
		Volume  float32 `json:"volume"`
		Pitch   float32 `json:"pitch"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if params.Volume == 0 {
		params.Volume = 1.0
	}
	if params.Pitch == 0 {
		params.Pitch = 1.0
	}
	if err := t.getClient().PlaySound(params.SoundID, params.X, params.Y, params.Z, params.Volume, params.Pitch); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success":   true,
		"sound_id":  params.SoundID,
		"position":  map[string]float64{"x": params.X, "y": params.Y, "z": params.Z},
		"volume":    params.Volume,
		"pitch":     params.Pitch,
	}, nil
}

// MCStopSoundTool 停止声音
type MCStopSoundTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCStopSoundTool() *MCStopSoundTool { return &MCStopSoundTool{} }
func (t *MCStopSoundTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCStopSoundTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCStopSoundTool) Name() string        { return "mc_stop_sound" }
func (t *MCStopSoundTool) Description() string { return "停止指定声音" }
func (t *MCStopSoundTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"sound_id": map[string]interface{}{"type": "string"},
		},
		"required": []string{"sound_id"},
	}
}
func (t *MCStopSoundTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		SoundID string `json:"sound_id"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().StopSound(params.SoundID); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "stopped": params.SoundID}, nil
}

// MCMakeParticlesTool 发射粒子
type MCMakeParticlesTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCMakeParticlesTool() *MCMakeParticlesTool { return &MCMakeParticlesTool{} }
func (t *MCMakeParticlesTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCMakeParticlesTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCMakeParticlesTool) Name() string { return "mc_particle" }
func (t *MCMakeParticlesTool) Description() string {
	return "在指定位置发射粒子效果。常用 particle_id: 1=smoke, 2=explode, 6=sparkler, 10=critical, 12=heart, 24=happy_villager, 25=firework, 29=end_rod, 33=dragon_breath"
}
func (t *MCMakeParticlesTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"particle_id": map[string]interface{}{"type": "integer", "description": "粒子 ID"},
			"x":           map[string]interface{}{"type": "number"},
			"y":           map[string]interface{}{"type": "number"},
			"z":           map[string]interface{}{"type": "number"},
		},
		"required": []string{"particle_id", "x", "y", "z"},
	}
}
func (t *MCMakeParticlesTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		ParticleID int32   `json:"particle_id"`
		X          float64 `json:"x"`
		Y          float64 `json:"y"`
		Z          float64 `json:"z"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().EmitParticle(params.ParticleID, params.X, params.Y, params.Z); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success":     true,
		"particle_id": params.ParticleID,
		"position":    map[string]float64{"x": params.X, "y": params.Y, "z": params.Z},
	}, nil
}

// MCRespawnTool 重生
type MCRespawnTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCRespawnTool() *MCRespawnTool { return &MCRespawnTool{} }
func (t *MCRespawnTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCRespawnTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCRespawnTool) Name() string        { return "mc_respawn" }
func (t *MCRespawnTool) Description() string { return "重生到指定坐标" }
func (t *MCRespawnTool) Parameters() map[string]interface{} {
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
func (t *MCRespawnTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if err := t.getClient().Respawn(params.X, params.Y, params.Z); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "position": map[string]float64{"x": params.X, "y": params.Y, "z": params.Z}}, nil
}

// MCSwingArmTool 挥臂
type MCSwingArmTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCSwingArmTool() *MCSwingArmTool { return &MCSwingArmTool{} }
func (t *MCSwingArmTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCSwingArmTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCSwingArmTool) Name() string        { return "mc_swing" }
func (t *MCSwingArmTool) Description() string { return "让玩家挥臂（挥剑/使用物品动画）" }
func (t *MCSwingArmTool) Parameters() map[string]interface{} {
	return map[string]interface{}{"type": "object", "properties": map[string]interface{}{}}
}
func (t *MCSwingArmTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	if err := t.getClient().SwingArm(); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{"success": true, "action": "swing_arm"}, nil
}

// MCBossBarTool Boss 血条
type MCBossBarTool struct {
	mu     sync.RWMutex
	client mc.ClientInterface
}

func NewMCBossBarTool() *MCBossBarTool { return &MCBossBarTool{} }
func (t *MCBossBarTool) SetClient(c mc.ClientInterface) {
	t.mu.Lock(); t.client = c; t.mu.Unlock()
}
func (t *MCBossBarTool) getClient() mc.ClientInterface {
	t.mu.RLock(); defer t.mu.RUnlock(); return t.client
}
func (t *MCBossBarTool) Name() string        { return "mc_bossbar" }
func (t *MCBossBarTool) Description() string { return "给玩家添加 Boss 血条（如难度提示、限时任务）" }
func (t *MCBossBarTool) Parameters() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"target":   map[string]interface{}{"type": "string", "description": "目标玩家，@s=自己"},
			"title":    map[string]interface{}{"type": "string", "description": "血条标题"},
			"health":   map[string]interface{}{"type": "number", "description": "血量比例 0.0-1.0", "default": 1.0},
		},
		"required": []string{"target", "title"},
	}
}
func (t *MCBossBarTool) Execute(ctx context.Context, args json.RawMessage) (interface{}, error) {
	var params struct {
		Target  string  `json:"target"`
		Title   string  `json:"title"`
		Health  float32 `json:"health"`
	}
	if err := json.Unmarshal(args, &params); err != nil {
		return nil, err
	}
	if params.Health == 0 {
		params.Health = 1.0
	}
	if err := t.getClient().SetBossBar(params.Target, params.Title, params.Health); err != nil {
		return map[string]interface{}{"success": false, "error": err.Error()}, nil
	}
	return map[string]interface{}{
		"success": true,
		"target":  params.Target,
		"title":   params.Title,
		"health":  params.Health,
	}, nil
}

// 编译期检查（10 个新工具全部满足 Tool interface）
var (
	_ Tool = (*MCAttackTool)(nil)
	_ Tool = (*MCSpawnEntityTool)(nil)
	_ Tool = (*MCRemoveEntityTool)(nil)
	_ Tool = (*MCPlaySoundTool)(nil)
	_ Tool = (*MCStopSoundTool)(nil)
	_ Tool = (*MCMakeParticlesTool)(nil)
	_ Tool = (*MCRespawnTool)(nil)
	_ Tool = (*MCSwingArmTool)(nil)
	_ Tool = (*MCBossBarTool)(nil)
)

// Sanity check - 占位防 fmt 引用消失
var _ = fmt.Sprintf
