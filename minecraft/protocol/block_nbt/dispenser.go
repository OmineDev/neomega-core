package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 发射器
type Dispenser struct {
	general.Dispenser
}

// ID ...
func (*Dispenser) ID() string {
	return IDDispenser
}
