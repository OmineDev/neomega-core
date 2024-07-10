package general

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/fields"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 描述各类告示牌的通用字段
type Sign struct {
	BackText  fields.SignText
	FrontText fields.SignText
	IsWaxed   byte `nbt:"IsWaxed"` // TAG_Byte(1) = 0
	Global
}

func (s *Sign) Marshal(r protocol.IO) {
	protocol.Single(r, &s.Global)
	protocol.Single(r, &s.FrontText)
	protocol.Single(r, &s.BackText)
	r.Uint8(&s.IsWaxed)
}

func (s *Sign) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
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
