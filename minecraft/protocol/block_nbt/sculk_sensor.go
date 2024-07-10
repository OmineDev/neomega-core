package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 幽匿感测体
type SculkSensor struct {
	VibrationListener map[string]any `nbt:"VibrationListener"` // Not used; TAG_Compound(10)
	general.Global
}

// ID ...
func (*SculkSensor) ID() string {
	return IDSculkSensor
}

func (s *SculkSensor) Marshal(io protocol.IO) {
	s.Global.Marshal(io)
}

func (s *SculkSensor) ToNBT() map[string]any {
	return slices.MergeMaps(
		s.Global.ToNBT(),
		map[string]any{
			"VibrationListener": s.VibrationListener,
		},
	)
}

func (s *SculkSensor) FromNBT(x map[string]any) {
	s.Global.FromNBT(x)
	s.VibrationListener = x["VibrationListener"].(map[string]any)
}
