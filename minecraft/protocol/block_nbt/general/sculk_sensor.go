package general

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 描述 幽匿感测体、校频幽匿感测体 和 幽匿尖啸体 的通用字段
type SculkSensor struct {
	VibrationListener map[string]any `nbt:"VibrationListener"` // Not used; TAG_Compound(10)
	Global
}

func (s *SculkSensor) Marshal(r protocol.IO) {
	protocol.Single(r, &s.Global)
}

func (s *SculkSensor) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
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
