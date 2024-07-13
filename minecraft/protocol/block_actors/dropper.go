package block_actors

import general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"

// 投掷器
type Dropper struct {
	general.DispenserBlockActor `mapstructure:",squash"`
}

// ID ...
func (*Dropper) ID() string {
	return IDDropper
}
