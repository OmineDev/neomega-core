package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 物品展示框
type ItemFrame struct {
	general.Frame
}

// ID ...
func (*ItemFrame) ID() string {
	return IDItemFrame
}
