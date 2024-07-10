package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 潮涌核心
type Conduit struct {
	Active byte  `nbt:"Active"` // TAG_Byte(1) = 0
	Target int64 `nbt:"Target"` // TAG_Long(5) = -1
	general.Global
}

// ID ...
func (*Conduit) ID() string {
	return IDConduit
}

func (c *Conduit) Marshal(io protocol.IO) {
	protocol.Single(io, &c.Global)
	io.Varint64(&c.Target)
	io.Uint8(&c.Active)
}

func (c *Conduit) ToNBT() map[string]any {
	return slices.MergeMaps(
		c.Global.ToNBT(),
		map[string]any{
			"Active": c.Active,
			"Target": c.Target,
		},
	)
}

func (c *Conduit) FromNBT(x map[string]any) {
	c.Global.FromNBT(x)
	c.Active = x["Active"].(byte)
	c.Target = x["Target"].(int64)
}
