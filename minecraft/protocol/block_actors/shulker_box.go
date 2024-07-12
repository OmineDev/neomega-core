package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
)

// 潜影盒
type ShulkerBox struct {
	general.ChestBlockActor
	Facing byte `mapstructure:"facing"` // TAG_Byte(1) = 0
}

// ID ...
func (*ShulkerBox) ID() string {
	return IDShulkerBox
}

func (s *ShulkerBox) Marshal(io protocol.IO) {
	protocol.NBTInt(&s.Facing, io.Varuint32)
	protocol.Single(io, &s.ChestBlockActor)
}
