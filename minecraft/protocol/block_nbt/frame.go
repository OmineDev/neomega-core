package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 物品展示框
type Frame struct {
	general.Frame
}

// ID ...
func (*Frame) ID() string {
	return IDFrame
}
