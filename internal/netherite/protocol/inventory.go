// Package protocol 实现 Minecraft Bedrock 协议。
package protocol

// ContainerID 容器 ID 常量
const (
	ContainerInventory      = 0
	ContainerHotbar         = 119
	ContainerOffhand        = 120
	ContainerArmor          = 124
	ContainerEnderChest     = 125
)

// InventoryTransactionType 物品交互类型
type InventoryTransactionType byte

const (
	InventoryTransactionTypeNormal InventoryTransactionType = iota
	InventoryTransactionTypeUseItem
	InventoryTransactionTypeUseItemOnEntity
	InventoryTransactionTypeReleaseItem
)

// InventoryTransaction 包 (ID=0xF2)
type InventoryTransaction struct {
	TransactionData TransactionData
}

// TransactionData 交互数据
type TransactionData struct {
	ActionType int32
	ActionData []TransactionAction
}

// TransactionAction 单个操作
type TransactionAction struct {
	SourceType int32
	Slot       int32
	OldItem    ItemStack
	NewItem    ItemStack
}

// ItemStack 物品栈
type ItemStack struct {
	NetworkID    int32
	Count        int16
	Metadata     int32
	BlockRuntimeID uint32
	Extra        []byte
}

// EncodeInventoryTransaction 编码物品交互
func EncodeInventoryTransaction() []byte {
	buf := NewWriter()
	buf.WriteByte(IDInventoryTransaction)
	buf.WriteVarint(0) // transaction type
	buf.WriteVarint(0) // action count
	return buf.Bytes()
}

// EncodeCreativeItem 发送创世之柜物品
func EncodeCreativeItem(item ItemStack) []byte {
	buf := NewWriter()
	buf.WriteByte(IDInventoryContent)
	buf.WriteVarint(0) // container ID
	buf.WriteVarint(1) // item count
	buf.WriteItem(item)
	return buf.Bytes()
}

// WriteItem 写入物品数据
func (w *Writer) WriteItem(item ItemStack) {
	w.WriteVarint(uint64(item.NetworkID))
	if item.NetworkID == 0 {
		return // 空气
	}
	w.WriteInt16(item.Count)
	w.WriteVarint(uint64(item.Metadata))
	w.WriteVarint(uint64(item.BlockRuntimeID))
	w.WriteBool(len(item.Extra) > 0)
	if len(item.Extra) > 0 {
		w.WriteVarint(uint64(len(item.Extra)))
		w.Write(item.Extra)
	}
}

// ReadItem 读取物品数据
func (r *Reader) ReadItem() ItemStack {
	item := ItemStack{}
	item.NetworkID = int32(r.ReadVarint())
	if item.NetworkID == 0 {
		return item
	}
	item.Count = r.ReadInt16()
	item.Metadata = int32(r.ReadVarint())
	item.BlockRuntimeID = uint32(r.ReadVarint())
	if r.ReadBool() {
		extraLen := r.ReadVarint()
		item.Extra = make([]byte, extraLen)
		r.Read(item.Extra)
	}
	return item
}

// InventoryContent 包 (ID=0xF4)
type InventoryContent struct {
	ContainerID int32
	Items       []ItemStack
}

// EncodeInventoryContent 编码背包内容
func EncodeInventoryContent(containerID int32, items []ItemStack) []byte {
	buf := NewWriter()
	buf.WriteByte(IDInventoryContent)
	buf.WriteVarint(uint64(containerID))
	buf.WriteVarint(uint64(len(items)))
	for _, item := range items {
		buf.WriteItem(item)
	}
	return buf.Bytes()
}

// EncodeInventorySlot 编码单个格子
func EncodeInventorySlot(containerID, slot int32, item ItemStack) []byte {
	buf := NewWriter()
	buf.WriteByte(IDInventorySlot)
	buf.WriteVarint(uint64(containerID))
	buf.WriteVarint(uint64(slot))
	buf.WriteItem(item)
	return buf.Bytes()
}

// ContainerOpen 包 (ID=0xF8)
type ContainerOpen struct {
	ContainerID      byte
	ContainerType   byte
	ContainerPos    Vec3
	MCryptographyID int64
}

// EncodeContainerOpen 编码打开容器
func EncodeContainerOpen(containerID byte, containerType byte, x, y, z float64) []byte {
	buf := NewWriter()
	buf.WriteByte(IDContainerOpen)
	buf.WriteByte(containerID)
	buf.WriteByte(containerType)
	buf.WriteFloat32(float32(x))
	buf.WriteFloat32(float32(y))
	buf.WriteFloat32(float32(z))
	buf.WriteVarint(0) // MCryptographyID
	return buf.Bytes()
}

// ContainerClose 包 (ID=0xF9)
type ContainerClose struct {
	ContainerID byte
}

// EncodeContainerClose 编码关闭容器
func EncodeContainerClose(containerID byte) []byte {
	buf := NewWriter()
	buf.WriteByte(IDContainerClose)
	buf.WriteByte(containerID)
	return buf.Bytes()
}

// SetSpawnPosition 设置出生点
type SetSpawnPosition struct {
	SpawnType  int32
	Position   Vec3
	Respawn    bool
}

// EncodeSetSpawnPosition 编码设置出生点
func EncodeSetSpawnPosition(x, y, z int32) []byte {
	buf := NewWriter()
	buf.WriteByte(0x46)
	buf.WriteVarint(0) // spawn type (0 = player spawn)
	buf.WriteInt32(x)
	buf.WriteInt32(y)
	buf.WriteInt32(z)
	buf.WriteBool(false) // spawn triggered
	return buf.Bytes()
}

// SetDefaultSpawnPosition 设置默认出生点
func SetDefaultSpawnPosition(x, y, z int32) []byte {
	buf := NewWriter()
	buf.WriteByte(0x46)
	buf.WriteVarint(0) // spawn type
	buf.WriteInt32(x)
	buf.WriteInt32(y)
	buf.WriteInt32(z)
	buf.WriteBool(true) // is default spawn
	return buf.Bytes()
}
