package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 校频幽匿感测体
type CalibratedSculkSensor struct {
	general.SculkSensor
}

// ID ...
func (*CalibratedSculkSensor) ID() string {
	return IDCalibratedSculkSensor
}
