package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 花盆
type FlowerPot struct {
	general.BlockActor
	PlantBlock map[string]any `nbt:"PlantBlock"` // TAG_Compound(10)
}

// ID ...
func (*FlowerPot) ID() string {
	return IDFlowerPot
}

func (f *FlowerPot) Marshal(io protocol.IO) {
	protocol.Single(io, &f.BlockActor)
	io.NBTWithLength(&f.PlantBlock)
}

func (f *FlowerPot) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		f.BlockActor.ToNBT(),
		map[string]any{
			"PlantBlock": f.PlantBlock,
		},
	)
}

func (f *FlowerPot) FromNBT(x map[string]any) {
	f.BlockActor.FromNBT(x)
	f.PlantBlock = x["PlantBlock"].(map[string]any)
}
