package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 花盆
type FlowerPot struct {
	PlantBlock map[string]any `nbt:"PlantBlock"` // TAG_Compound(10)
	general.Global
}

// ID ...
func (*FlowerPot) ID() string {
	return IDFlowerPot
}

func (f *FlowerPot) Marshal(io protocol.IO) {
	f.Global.Marshal(io)
	io.NBTWithLength(&f.PlantBlock)
}

func (f *FlowerPot) ToNBT() map[string]any {
	return slices.MergeMaps(
		f.Global.ToNBT(),
		map[string]any{
			"PlantBlock": f.PlantBlock,
		},
	)
}

func (f *FlowerPot) FromNBT(x map[string]any) {
	f.Global.FromNBT(x)
	f.PlantBlock = x["PlantBlock"].(map[string]any)
}
