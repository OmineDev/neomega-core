package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 荧光物品展示框
type GlowItemFrame struct {
	general.Frame
}

// ID ...
func (*GlowItemFrame) ID() string {
	return IDGlowItemFrame
}
