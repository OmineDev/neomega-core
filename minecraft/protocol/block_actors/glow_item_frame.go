package block_actors

import general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"

// 荧光物品展示框
type GlowItemFrame struct {
	general.ItemFrameBlockActor
}

// ID ...
func (*GlowItemFrame) ID() string {
	return IDGlowItemFrame
}