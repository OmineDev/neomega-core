package block_actors

import general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"

// 箱子
type Chest struct {
	general.ChestBlockActor
}

// ID ...
func (c *Chest) ID() string {
	return IDChest
}
