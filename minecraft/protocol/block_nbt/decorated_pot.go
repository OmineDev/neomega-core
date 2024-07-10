package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 饰纹陶罐
type DecoratedPot struct {
	Animation byte          `nbt:"animation"` // Not used; TAG_Byte(1) = 0
	Item      protocol.Item `nbt:"item"`      // Not used; TAG_Compound(10)
	general.Global
}

// ID ...
func (*DecoratedPot) ID() string {
	return IDDecoratedPot
}

func (d *DecoratedPot) Marshal(io protocol.IO) {
	d.Global.Marshal(io)
}

func (d *DecoratedPot) ToNBT() map[string]any {
	return utils.MergeMaps(
		d.Global.ToNBT(),
		map[string]any{
			"animation": d.Animation,
			"item":      d.Item.ToNBT(),
		},
	)
}

func (d *DecoratedPot) FromNBT(x map[string]any) {
	d.Global.FromNBT(x)
	d.Animation = x["animation"].(byte)
	d.Item.FromNBT(x["item"].(map[string]any))
}
