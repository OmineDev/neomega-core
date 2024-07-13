package block_actors

import general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"

// 悬挂式告示牌
type HangingSign struct {
	general.SignBlockActor `mapstructure:",squash"`
}

// ID ...
func (*HangingSign) ID() string {
	return IDHangingSign
}
