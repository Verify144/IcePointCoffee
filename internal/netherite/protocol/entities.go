package protocol

// ===== 新增独有协议包 ID =====

const (
	IDInteract         = 0x21 // 交互（攻击/右键）
	IDRespawn         = 0x2D // 复活
	IDMobEquipment    = 0x1F // 生物装备（穿戴）
	IDBossEvent       = 0x4A // Boss 事件（见 player_ext.go）
	IDStopSound       = 0x5A // 停止声音
	IDPlaySound       = 0x56 // 播放声音
	IDSetDefaultGamemode = 0x6E // 设置默认游戏模式
)

// ===== InteractPacket (0x21) =====

// InteractAction 交互动作
type InteractAction uint64

const (
	InteractActionLeftClick  InteractAction = 100 // 左键攻击
	InteractActionRightClick InteractAction = 101 // 右键交互
	InteractActionInteract    InteractAction = 110 // 通用交互
	InteractActionMount      InteractAction = 111 // 骑乘
	InteractActionDismount   InteractAction = 112 // 下马
)

// Interact 交互/攻击包
type Interact struct {
	Action   InteractAction
	TargetID uint64 // 被攻击的实体 runtime ID
	PosX     float64
	PosY     float64
	PosZ     float64
}

// EncodeInteract 编码攻击/交互包
func EncodeInteract(action InteractAction, targetID uint64, x, y, z float64) []byte {
	buf := NewWriter()
	buf.WriteByte(IDInteract)
	buf.WriteVarint(uint64(action))
	buf.WriteVarint(targetID)
	buf.WriteFloat64(x)
	buf.WriteFloat64(y)
	buf.WriteFloat64(z)
	return buf.Bytes()
}

// DecodeInteract 解码交互包
func DecodeInteract(data []byte) *Interact {
	r := NewReader(data)
	r.ReadByte() // ID
	pk := &Interact{}
	pk.Action = InteractAction(r.ReadVarint())
	pk.TargetID = r.ReadVarint()
	pk.PosX = r.ReadFloat64()
	pk.PosY = r.ReadFloat64()
	pk.PosZ = r.ReadFloat64()
	return pk
}

// ===== RespawnPacket (0x2D) =====

// RespawnState 重生状态
type RespawnState uint64

const (
	RespawnStateSearching     RespawnState = 0
	RespawnStateReadyToSpawn RespawnState = 1
	RespawnStateSpawning     RespawnState = 2
)

// Respawn 重生请求
type Respawn struct {
	State   RespawnState
	PosX    float64
	PosY    float64
	PosZ    float64
}

// EncodeRespawn 编码重生请求
func EncodeRespawn(x, y, z float64) []byte {
	buf := NewWriter()
	buf.WriteByte(IDRespawn)
	buf.WriteVarint(uint64(RespawnStateReadyToSpawn))
	buf.WriteFloat64(x)
	buf.WriteFloat64(y)
	buf.WriteFloat64(z)
	buf.WriteFloat32(0)
	buf.WriteFloat32(0)
	buf.WriteFloat64(0)
	buf.WriteFloat64(0)
	buf.WriteFloat64(0)
	return buf.Bytes()
}

// ===== MobEquipmentPacket (0x1F) =====

// EncodeMobEquipment 编码装备包（手持/穿戴）
func EncodeMobEquipment(runtimeID uint64, item ItemStack, slot, hotbar byte) []byte {
	buf := NewWriter()
	buf.WriteByte(IDMobEquipment)
	buf.WriteVarint(runtimeID)
	buf.WriteItem(item)
	buf.WriteByte(slot)
	buf.WriteByte(hotbar)
	return buf.Bytes()
}

// ===== PlaySound (0x56) =====

// EncodePlaySound 编码播放声音
func EncodePlaySound(soundID string, x, y, z float64, volume, pitch float32) []byte {
	buf := NewWriter()
	buf.WriteByte(IDPlaySound)
	buf.WriteString(soundID)
	buf.WriteVarint(uint64(x))
	buf.WriteVarint(uint64(y))
	buf.WriteVarint(uint64(z))
	buf.WriteFloat32(volume)
	buf.WriteFloat32(pitch)
	return buf.Bytes()
}

// ===== StopSound (0x5A) =====

// EncodeStopSound 编码停止声音
func EncodeStopSound(soundID string, stopAll bool) []byte {
	buf := NewWriter()
	buf.WriteByte(IDStopSound)
	buf.WriteString(soundID)
	buf.WriteBool(stopAll)
	return buf.Bytes()
}

// ===== SetDefaultGamemode (0x6E) =====

// EncodeSetDefaultGamemode 设置默认游戏模式
func EncodeSetDefaultGamemode(gamemode int32) []byte {
	buf := NewWriter()
	buf.WriteByte(IDSetDefaultGamemode)
	buf.WriteVarint(uint64(gamemode))
	return buf.Bytes()
}

// ===== 常用声音 ID =====

const (
	SoundAmbientCave     = "ambient.cave"
	SoundAmbientNether   = "ambient.nether"
	SoundRain            = "ambient.weather.rain"
	SoundThunder         = "ambient.weather.thunder"
	SoundClick           = "ui.button.click"
	SoundPop             = "random.pop"
	SoundBowUse          = "bow.use"
	SoundBowHit          = "bow.hit"
	SoundFizz            = "random.fizz"
	SoundLevelUp         = "random.levelup"
	SoundExplode         = "explode"
	SoundBass            = "note.bass"
	SoundPling           = "note.pling"
	SoundBell            = "note.bell"
	SoundBlockBreak      = "dig.stone"
	SoundBlockPlace      = "step.stone"
	SoundChicken         = "mob.chicken.say"
	SoundCow             = "mob.cow.say"
	SoundPig             = "mob.pig.say"
	SoundSheep           = "mob.sheep.say"
	SoundWolf            = "mob.wolf.growl"
	SoundZombie          = "mob.zombie.say"
	SoundSkeleton        = "mob.skeleton.say"
	SoundCreeper         = "mob.creeper.say"
	SoundEnderman        = "mob.endermen.scream"
	SoundGhast           = "mob.ghast.moan"
	SoundBlaze           = "mob.blaze.breathe"
	SoundLava            = "liquid.lava"
	SoundWater           = "liquid.water"
	SoundPortal          = "portal.portal"
	SoundChestOpen       = "random.chestopen"
	SoundChestClose      = "random.chestclosed"
	SoundDoor            = "random.door"
	SoundFuse            = "game.tnt.primed"
	SoundGuardian        = "mob.guardian.flop"
	SoundHurt            = "game.player.hurt"
	SoundDeath           = "game.player.die"
	SoundDamage          = "damage.hit"
	SoundSplash          = "liquid.splash"
	SoundCollect         = "item.bottle.fill"
	SoundBrewing         = "potion.brewing"
	SoundAnvilUse        = "anvil.use"
	SoundAnvilLand       = "anvil.land"
	SoundSmithing        = "ui.stonecutter.take_result"
)

// ===== 常用粒子 ID =====

const (
	ParticleSmoke           = 1
	ParticleExplode        = 2
	ParticleBubble         = 3
	ParticleSplash        = 4
	ParticleWhiteSmoke     = 5
	ParticleSparkler       = 6
	ParticleLava           = 7
	ParticleFlame          = 8
	ParticleLavaDrip       = 9
	ParticleCritical       = 10
	ParticleTrail         = 11
	ParticleHeart         = 12
	ParticleInk          = 14
	ParticleSlime        = 15
	ParticleRainSplash   = 16
	ParticleSnowball     = 17
	ParticleMist         = 20
	ParticleEmerald      = 21
	ParticleVillager     = 22
	ParticleHappyVillager = 24
	ParticleFirework    = 25
	ParticleNoteBlock    = 26
	ParticleEndRod      = 29
	ParticleDragonBreath = 33
	ParticleSoul         = 36
	ParticleFlash       = 37
	ParticleBlockDust   = 38
	ParticleSculkSoul   = 40
)
