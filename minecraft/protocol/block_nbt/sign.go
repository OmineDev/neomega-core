package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 告示牌
type Sign struct {
	BackText  general.SignText
	FrontText general.SignText
	IsWaxed   byte `nbt:"IsWaxed"` // TAG_Byte(1) = 0
	general.Global
}

// ID ...
func (*Sign) ID() string {
	return IDSign
}

func (s *Sign) Marshal(io protocol.IO) {
	s.Global.Marshal(io)
	s.FrontText.Marshal(io)
	s.BackText.Marshal(io)
	io.Uint8(&s.IsWaxed)
}

func (s *Sign) ToNBT() map[string]any {
	return utils.MergeMaps(
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
