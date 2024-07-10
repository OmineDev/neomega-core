package general

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/fields"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 描述 物品展示框 和 荧光物品展示框 的通用字段
type Frame struct {
	Frame protocol.Optional[fields.Frame]
	Global
}

func (f *Frame) Marshal(r protocol.IO) {
	protocol.Single(r, &f.Global)
	protocol.OptionalMarshaler(r, &f.Frame)
}

func (f *Frame) ToNBT() map[string]any {
	var temp map[string]any
	if frame, has := f.Frame.Value(); has {
		temp = frame.ToNBT()
	}
	return slices.MergeMaps(
		f.Global.ToNBT(),
		temp,
	)
}

func (f *Frame) FromNBT(x map[string]any) {
	f.Global.FromNBT(x)

	new := fields.Frame{}
	if new.CheckExist(x) {
		new.FromNBT(x)
		f.Frame = protocol.Optional[fields.Frame]{Set: true, Val: new}
	}
}
