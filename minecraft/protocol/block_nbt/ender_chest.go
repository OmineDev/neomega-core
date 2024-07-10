package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 末影箱
type EnderChest struct {
	general.Chest
}

// ID ...
func (*EnderChest) ID() string {
	return IDEnderChest
}
