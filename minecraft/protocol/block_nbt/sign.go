package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 告示牌
type Sign struct {
	general.Sign
}

// ID ...
func (*Sign) ID() string {
	return IDSign
}
