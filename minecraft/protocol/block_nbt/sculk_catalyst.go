package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 幽匿催发体
type SculkCatalyst struct {
	general.Global
}

// ID ...
func (*SculkCatalyst) ID() string {
	return IDSculkCatalyst
}
