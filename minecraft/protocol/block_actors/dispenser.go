package block_actors

import general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"

// 发射器
type Dispenser struct {
	general.DispenserBlockActor
}

// ID ...
func (*Dispenser) ID() string {
	return IDDispenser
}
