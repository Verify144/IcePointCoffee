package protocol

// PlayerSkinFlags 皮肤标志
type PlayerSkinFlags int32

const (
	SkinPersonaPieceFlashing PlayerSkinFlags = 1 << iota
	SkinPersonaAnimationBlinking
	SkinMetaDataForceDisabled
	SkinMetaDataPremiumTutorial
	SkinMetaDataUsingSkin
	SkinMetaDataSpawnedForWorld
)

// SkinImage 皮肤图像数据
type SkinImage struct {
	Width    int32
	Height   int32
	DataLen  int32
	Data     []byte
}

// SkinData 皮肤数据
type SkinData struct {
	SkinID            string
	PlayFabID         string
	SkinResourcePatch string
	SkinImage         SkinImage
	CapeImage         SkinImage
	GeometryName      string
	GeometryData      string
	AnimationData     string
	PremiumSkin       bool
	SkinType          string
	ArmSize           string
	TrustedSkin       bool
}

// PlayerSkin 包 (ID=0x9F)
type PlayerSkin struct {
	UUID    string
	Name    string
	SkinID  string
	Skin    SkinData
}

// EncodePlayerSkin 编码玩家皮肤
func EncodePlayerSkin(uuid, name, skinID string, skin *SkinData) []byte {
	buf := NewWriter()
	buf.WriteByte(IDPlayerSkin)
	buf.WriteString(uuid)
	buf.WriteString(name)
	buf.WriteString(skinID)
	if skin != nil {
		buf.WriteBool(true)
		buf.WriteString(skin.SkinID)
		buf.WriteString(skin.SkinResourcePatch)
		buf.WriteInt32(skin.SkinImage.Width)
		buf.WriteInt32(skin.SkinImage.Height)
		buf.WriteInt32(skin.SkinImage.DataLen)
		if len(skin.SkinImage.Data) > 0 {
			buf.Write(skin.SkinImage.Data)
		}
		buf.WriteInt32(0) // cape width
		buf.WriteInt32(0) // cape height
		buf.WriteInt32(0) // cape data len
		buf.WriteString(skin.GeometryName)
		buf.WriteString(skin.GeometryData)
		buf.WriteBool(skin.PremiumSkin)
		buf.WriteString(skin.SkinType)
		buf.WriteString(skin.ArmSize)
		buf.WriteBool(skin.TrustedSkin)
	} else {
		buf.WriteBool(false)
	}
	return buf.Bytes()
}

// PlayerListEntry 玩家列表条目
type PlayerListEntry struct {
	UUID       string
	EntityID   int64
	Name       string
	XUID       string
	PlatformChatID string
	BuildVersion string
	Version    int32
	LibrariesDownloaded bool
	Geometry  SkinData
	FrameCount int32
	AnimatedImageData []SkinImage
}

// PlayerList 包 (ID=0x3F)
type PlayerList struct {
	Action  int32 // 0=add, 4=remove
	Entries []PlayerListEntry
}

// EncodePlayerListAdd 编码添加玩家
func EncodePlayerListAdd(entries []PlayerListEntry) []byte {
	buf := NewWriter()
	buf.WriteByte(IDPlayerList)
	buf.WriteVarint(0) // action = add
	buf.WriteVarint(uint64(len(entries)))
	for _, e := range entries {
		buf.WriteString(e.UUID)
		buf.WriteVarint(uint64(e.EntityID))
		buf.WriteString(e.Name)
		buf.WriteString(e.XUID)
		buf.WriteString(e.PlatformChatID)
		buf.WriteString(e.BuildVersion)
		buf.WriteInt32(e.Version)
		buf.WriteBool(e.LibrariesDownloaded)
		if len(e.Geometry.GeometryData) > 0 {
			buf.WriteBool(true)
			buf.WriteString(e.Geometry.SkinID)
			buf.WriteString(e.Geometry.SkinResourcePatch)
			buf.WriteInt32(0) // image
			buf.WriteInt32(0)
			buf.WriteInt32(0)
			buf.WriteString(e.Geometry.GeometryName)
			buf.WriteString(e.Geometry.GeometryData)
			buf.WriteBool(e.Geometry.PremiumSkin)
			buf.WriteString(e.Geometry.SkinType)
			buf.WriteString(e.Geometry.ArmSize)
			buf.WriteBool(e.Geometry.TrustedSkin)
		} else {
			buf.WriteBool(false)
		}
		buf.WriteVarint(0) // frame count
	}
	return buf.Bytes()
}

// EncodePlayerListRemove 编码移除玩家
func EncodePlayerListRemove(uuids []string) []byte {
	buf := NewWriter()
	buf.WriteByte(IDPlayerList)
	buf.WriteVarint(4) // action = remove
	buf.WriteVarint(uint64(len(uuids)))
	for _, uuid := range uuids {
		buf.WriteString(uuid)
	}
	return buf.Bytes()
}

// CameraPresets 相机预设
const (
	CameraPresetFirstPerson CameraPreset = iota
	CameraPresetThirdPerson
	CameraPresetThirdPersonFront
	CameraPresetThirdPersonRear
)

// CameraPreset 相机预设 ID
type CameraPreset int32

// Camera 包 (ID=0x9B)
type Camera struct {
	CameraPreset  CameraPreset
	Survivor      string
	TargetSurvivor string
}

// EncodeCameraPresets 编码相机预设
func EncodeCameraPresets() []byte {
	buf := NewWriter()
	buf.WriteByte(0x9B)
	buf.WriteVarint(uint64(CameraPresetFirstPerson))
	buf.WriteString("")
	buf.WriteString("")
	return buf.Bytes()
}

// BossEvent 包 (ID=0x4A)
type BossEvent struct {
	EntityUniqueID int64
	Event          int32
	HealthPercent  float32
	Title          string
	ProgressBarString string
	ScreenBarHeight int32
	ColorOverlay   int32
	OverlayPercentage float32
}

// BossEventType Boss 事件类型
const (
	BossEventShow BossEventType = iota
	BossEventRegister
	BossEventUpdateProperties
	BossEventUpdateVisibility
	BossEventEntityEnterLevel
	BossEventUnregister
)

// BossEventType Boss 事件类型
type BossEventType int32

// EncodeBossBar 编码 Boss 血条
func EncodeBossBar(entityID int64, title string, healthPercent float32) []byte {
	buf := NewWriter()
	buf.WriteByte(0x4A)
	buf.WriteVarint(uint64(entityID))
	buf.WriteVarint(uint64(BossEventShow))
	buf.WriteFloat32(healthPercent)
	buf.WriteString(title)
	buf.WriteString("")
	buf.WriteVarint(0)
	buf.WriteVarint(0)
	buf.WriteFloat32(0)
	return buf.Bytes()
}

// MapTrackedObject 地图追踪对象
type MapTrackedObject struct {
	ObjectType int32
	ObjectUUID string
}

// MapDecorations 地图装饰
type MapDecorations struct {
	DecorationType int32
	X              int32
	Z              int32
	Rotation       int32
	Icon           int32
	DecorationUUID string
	Animate        bool
}

// MapData 包 (ID=0x86)
type MapData struct {
	MapID          int64
	Scale         int32
	Locked        bool
	TrackingPos   bool
	DecorationsCount int32
	Decorations   []MapDecorations
}

// EncodeMapData 编码地图数据
func EncodeMapData(mapID int64, scale int32) []byte {
	buf := NewWriter()
	buf.WriteByte(0x86)
	buf.WriteVarint(uint64(mapID))
	buf.WriteByte(byte(scale))
	buf.WriteBool(false) // locked
	buf.WriteBool(false) // tracking pos
	buf.WriteVarint(0) // decorations
	return buf.Bytes()
}

// TransferPacket 包 (ID=0x9A)
type Transfer struct {
	ServerAddress string
	Port          int16
}

// EncodeTransfer 编码服务器传送
func EncodeTransfer(address string, port int16) []byte {
	buf := NewWriter()
	buf.WriteByte(0x9A)
	buf.WriteString(address)
	buf.WriteInt16(port)
	return buf.Bytes()
}

// GameRules 包 (ID=0xA1)
type GameRules struct {
	Rules map[string]GameRuleValue
}

// GameRuleValue 游戏规则值
type GameRuleValue struct {
	Type  int32 // 1=bool, 2=int, 3=float
	Bool  bool
	Int   int32
	Float float32
}

// EncodeGameRules 编码游戏规则
func EncodeGameRules(rules map[string]bool) []byte {
	buf := NewWriter()
	buf.WriteByte(0xA1)
	buf.WriteVarint(uint64(len(rules)))
	for name, val := range rules {
		buf.WriteString(name)
		buf.WriteVarint(1) // bool type
		buf.WriteBool(val)
	}
	return buf.Bytes()
}

// EncodeWeather 编码天气
func EncodeWeather(rainLevel, lightningLevel float32) []byte {
	return EncodeGameRules(map[string]bool{
		"doWeatherCycle": rainLevel > 0 || lightningLevel > 0,
	})
}

// PlayerEnchantOptions 包 (ID=0x9E)
type PlayerEnchantOptions struct {
	Options []EnchantOption
}

// EnchantOption 附魔选项
type EnchantOption struct {
	Cost    int32
	Seed    int64
	Enchants []Enchantment
}

// Enchantment 附魔
type Enchantment struct {
	ID    int32
	Level int32
}

// EncodePlayerEnchantOptions 编码玩家附魔选项
func EncodePlayerEnchantOptions(options []EnchantOption) []byte {
	buf := NewWriter()
	buf.WriteByte(0x9E)
	buf.WriteVarint(uint64(len(options)))
	for _, opt := range options {
		buf.WriteVarint(uint64(opt.Cost))
		buf.WriteVarint(uint64(opt.Seed))
		buf.WriteVarint(uint64(len(opt.Enchants)))
		for _, e := range opt.Enchants {
			buf.WriteVarint(uint64(e.ID))
			buf.WriteVarint(uint64(e.Level))
		}
	}
	return buf.Bytes()
}
