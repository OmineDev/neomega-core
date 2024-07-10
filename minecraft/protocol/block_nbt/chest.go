package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 箱子
type Chest struct {
	general.Chest
}

// ID ...
func (c *Chest) ID() string {
	return IDChest
}
