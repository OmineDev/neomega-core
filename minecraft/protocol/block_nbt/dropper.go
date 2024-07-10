package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 投掷器
type Dropper struct {
	general.Dispenser
}

// ID ...
func (*Dropper) ID() string {
	return IDDropper
}
