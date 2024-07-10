package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/fields"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 物品展示框
type Frame struct {
	Frame protocol.Optional[fields.Frame]
	general.Global
}

// ID ...
func (*Frame) ID() string {
	return IDFrame
}

func (f *Frame) Marshal(io protocol.IO) {
	f.Global.Marshal(io)
	protocol.OptionalMarshaler(io, &f.Frame)
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
