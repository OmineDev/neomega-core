package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 高炉
type BlastFurnace struct {
	general.Furnace
}

// ID ...
func (*BlastFurnace) ID() string {
	return IDBlastFurnace
}
