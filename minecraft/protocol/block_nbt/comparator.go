package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 比较器
type Comparator struct {
	OutputSignal int32 `nbt:"OutputSignal"` // TAG_Int(4) = 0
	general.Global
}

// ID ...
func (*Comparator) ID() string {
	return IDComparator
}

func (c *Comparator) Marshal(io protocol.IO) {
	protocol.Single(io, &c.Global)
	io.Varint32(&c.OutputSignal)
}

func (c *Comparator) ToNBT() map[string]any {
	return slices.MergeMaps(
		c.Global.ToNBT(),
		map[string]any{
			"OutputSignal": c.OutputSignal,
		},
	)
}

func (c *Comparator) FromNBT(x map[string]any) {
	c.Global.FromNBT(x)
	c.OutputSignal = x["OutputSignal"].(int32)
}
