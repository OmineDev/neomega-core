package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 幽匿感测体
type SculkSensor struct {
	general.SculkSensor
}

// ID ...
func (*SculkSensor) ID() string {
	return IDSculkSensor
}
