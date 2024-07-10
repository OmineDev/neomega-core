package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 悬挂式告示牌
type HangingSign struct {
	general.Sign
}

// ID ...
func (*HangingSign) ID() string {
	return IDHangingSign
}
