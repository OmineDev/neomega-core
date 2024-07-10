package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 潜影盒
type ShulkerBox struct {
	Facing uint32 `nbt:"facing"` // * TAG_Byte(1) = 0
	Chest
}

// ID ...
func (*ShulkerBox) ID() string {
	return IDShulkerBox
}

func (s *ShulkerBox) Marshal(io protocol.IO) {
	io.Varuint32(&s.Facing)
	s.Chest.Marshal(io)
}

func (s *ShulkerBox) ToNBT() map[string]any {
	return utils.MergeMaps(
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
