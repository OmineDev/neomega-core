package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 饰纹陶罐
type DecoratedPot struct {
	general.BlockActor
	Animation byte          `nbt:"animation"` // Not used; TAG_Byte(1) = 0
	Item      protocol.Item `nbt:"item"`      // Not used; TAG_Compound(10)
}

// ID ...
func (*DecoratedPot) ID() string {
	return IDDecoratedPot
}

func (d *DecoratedPot) Marshal(io protocol.IO) {
	protocol.Single(io, &d.BlockActor)
}

func (d *DecoratedPot) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		d.BlockActor.ToNBT(),
		map[string]any{
			"animation": d.Animation,
			"item":      d.Item.ToNBT(),
		},
	)
}

func (d *DecoratedPot) FromNBT(x map[string]any) {
	d.BlockActor.FromNBT(x)
	d.Animation = x["animation"].(byte)
	d.Item.FromNBT(x["item"].(map[string]any))
}
