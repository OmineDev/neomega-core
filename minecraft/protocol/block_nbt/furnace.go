package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 熔炉
type Furnace struct {
	general.Furnace
}

// ID ...
func (*Furnace) ID() string {
	return IDFurnace
}
