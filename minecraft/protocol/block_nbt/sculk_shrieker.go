package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 幽匿尖啸体
type SculkShrieker struct {
	general.SculkSensor
}

// ID ...
func (*SculkShrieker) ID() string {
	return IDSculkShrieker
}
