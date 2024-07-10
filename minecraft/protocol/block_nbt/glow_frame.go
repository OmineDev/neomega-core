package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 荧光物品展示框
type GlowFrame struct {
	general.Frame
}

// ID ...
func (*GlowFrame) ID() string {
	return IDGlowFrame
}
