package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 潜影盒
type ShulkerBox struct {
	general.ChestBlockActor
	Facing uint32 `nbt:"facing"` // * TAG_Byte(1) = 0
}

// ID ...
func (*ShulkerBox) ID() string {
	return IDShulkerBox
}

func (s *ShulkerBox) Marshal(io protocol.IO) {
	io.Varuint32(&s.Facing)
	protocol.Single(io, &s.ChestBlockActor)
}

func (s *ShulkerBox) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		map[string]any{
			"facing": byte(s.Facing),
		},
		s.ChestBlockActor.ToNBT(),
	)
}

func (s *ShulkerBox) FromNBT(x map[string]any) {
	s.Facing = uint32(x["facing"].(byte))
	s.ChestBlockActor.FromNBT(x)
}
