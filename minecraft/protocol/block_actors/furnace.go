package block_actors

import general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"

// 熔炉
type Furnace struct {
	general.FurnaceBlockActor
}

// ID ...
func (*Furnace) ID() string {
	return IDFurnace
}
