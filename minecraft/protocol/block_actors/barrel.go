package block_actors

import general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"

// 木桶
type Barrel struct {
	general.ChestBlockActor
}

// ID ...
func (*Barrel) ID() string {
	return IDBarrel
}
