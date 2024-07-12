package fields

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

// 描述 物品展示框 和 荧光物品展示框 的共用字段
type FrameItem struct {
	Item           protocol.Item `mapstructure:"Item"`           // TAG_Compound(10)
	ItemDropChance float32       `mapstructure:"ItemDropChance"` // TAG_Float(6) = 1
	ItemRotation   float32       `mapstructure:"ItemRotation"`   // TAG_Float(6) = 0
}

func (f *FrameItem) Marshal(r protocol.IO) {
	r.NBTItem(&f.Item)
	r.Float32(&f.ItemRotation)
	r.Float32(&f.ItemDropChance)
}
