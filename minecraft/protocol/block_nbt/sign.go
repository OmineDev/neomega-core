package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/fields"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 告示牌
type Sign struct {
	BackText  fields.SignText
	FrontText fields.SignText
	IsWaxed   byte `nbt:"IsWaxed"` // TAG_Byte(1) = 0
	general.Global
}

// ID ...
func (*Sign) ID() string {
	return IDSign
}

func (s *Sign) Marshal(io protocol.IO) {
	protocol.Single(io, &s.Global)
	protocol.Single(io, &s.FrontText)
	protocol.Single(io, &s.BackText)
	io.Uint8(&s.IsWaxed)
}

func (s *Sign) ToNBT() map[string]any {
	return slices.MergeMaps(
		s.Global.ToNBT(),
		map[string]any{
			"BackText":  s.BackText.ToNBT(),
			"FrontText": s.FrontText.ToNBT(),
			"IsWaxed":   s.IsWaxed,
		},
	)
}

func (s *Sign) FromNBT(x map[string]any) {
	s.Global.FromNBT(x)
	s.BackText.FromNBT(x["BackText"].(map[string]any))
	s.FrontText.FromNBT(x["FrontText"].(map[string]any))
	s.IsWaxed = x["IsWaxed"].(byte)
}
