# PHASE 12 — 更多 MC 协议包

## 目标
补齐核心交互能力，让 AI 能更细腻地操控租赁服（不重连）：
- 实体攻击（Hit/伤害）
- 玩家装备（穿戴）
- 粒子效果（特效展示）
- 声音播放
- 方块更新（动画）
- 实体生成/移除
- 死亡/重生
- 世界边界

## 协议包列表

### 玩家交互
- [ ] `mc_attack` - 攻击实体（InteractPacket 0x21）
- [ ] `mc_respawn` - 死亡重生（RespawnPacket 0x2D）
- [ ] `mc_equipment` - 装备物品（MobEquipmentPacket 0x1F）

### 实体操作
- [ ] `mc_spawn_entity` - 生成生物（AddActorPacket 0x0D）
- [ ] `mc_remove_entity` - 移除实体（RemoveActorPacket 0x0E）
- [ ] `mc_emit_particle` - 粒子效果（LevelEvent 0x12）
- [ ] `mc_play_sound` - 播放声音（PlaySound 0x56）

### 视觉效果
- [ ] `mc_block_animation` - 方块动画（BlockEvent 0x1A）
- [ ] `mc_set_time` - 时间已存在（增强版本）
- [ ] `mc_boss_event` - Boss 条

### 玩家状态
- [ ] `mc_health` - 玩家生命/饥饿
- [ ] `mc_experience` - 玩家经验
- [ ] `mc_skin` - 玩家皮肤（已存在 player_ext.go）
- [ ] `mc_animation` - 玩家动画（挥剑/挥击等）

## 协议 ID 参考（Bedrock 1.21.x）
- InteractPacket = 0x21
- RespawnPacket = 0x2D
- MobEquipmentPacket = 0x1F
- AddActorPacket = 0x0D
- RemoveActorPacket = 0x0E
- LevelEvent = 0x12 (粒子)
- PlaySound = 0x56
- BlockEvent = 0x1A
- AnimatePacket = 0x1B
- UpdateAttributes = 0x0F
- BossEvent = 0x4B

## AI 工具集成
每个新协议对应一个 AI 工具（mc_*），让 AI 在对话中能使用：
- mc_attack, mc_respawn, mc_equipment
- mc_spawn_entity, mc_remove_entity
- mc_particle, mc_sound
- mc_animate
- mc_health

## 测试
- 每个协议包的 Encode/Decode 单测
- 协议包性能基准
- AI 工具黑名单/参数验证
