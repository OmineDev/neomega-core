package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 潜影盒
type ShulkerBox struct {
	Facing uint32 `nbt:"facing"` // * TAG_Byte(1) = 0
	general.Chest
}

// ID ...
func (*ShulkerBox) ID() string {
	return IDShulkerBox
}

func (s *ShulkerBox) Marshal(io protocol.IO) {
	io.Varuint32(&s.Facing)
	protocol.Single(io, &s.Chest)
}

func (s *ShulkerBox) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		map[string]any{
			"facing": byte(s.Facing),
		},
		s.Chest.ToNBT(),
	)
}

func (s *ShulkerBox) FromNBT(x map[string]any) {
	s.Facing = uint32(x["facing"].(byte))
	s.Chest.FromNBT(x)
}
