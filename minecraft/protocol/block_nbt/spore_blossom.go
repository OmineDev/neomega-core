package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 孢子花
type SporeBlossom struct {
	general.Global
}

// ID ...
func (*SporeBlossom) ID() string {
	return IDSporeBlossom
}
