package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 木桶
type Barrel struct {
	general.Chest
}

// ID ...
func (*Barrel) ID() string {
	return IDBarrel
}
