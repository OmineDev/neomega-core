package general

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/fields"
)

// 描述 物品展示框 和 荧光物品展示框 的通用字段
type ItemFrameBlockActor struct {
	BlockActor
	*fields.FrameItem `mapstructure:",omitempty"`
}

func (f *ItemFrameBlockActor) Marshal(r protocol.IO) {
	fun := func() *fields.FrameItem {
		if f.FrameItem == nil {
			f.FrameItem = new(fields.FrameItem)
		}
		return f.FrameItem
	}

	protocol.Single(r, &f.BlockActor)
	protocol.NBTOptionalMarshaler(r, f.FrameItem, fun, true)
}
