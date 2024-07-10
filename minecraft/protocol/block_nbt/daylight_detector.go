package block_nbt

import "github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"

// 阳光探测器
type DayLightDetector struct {
	general.Global
}

// ID ...
func (*DayLightDetector) ID() string {
	return IDDayLightDetector
}
