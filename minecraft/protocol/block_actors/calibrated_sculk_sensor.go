package block_actors

import general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"

// 校频幽匿感测体
type CalibratedSculkSensor struct {
	general.BlockActor
}

// ID ...
func (*CalibratedSculkSensor) ID() string {
	return IDCalibratedSculkSensor
}
