package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 烟熏炉
type Smoker struct {
	general.Furnace
}

// ID ...
func (*Smoker) ID() string {
	return IDSmoker
}
