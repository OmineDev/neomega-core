package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 信标
type Beacon struct {
	Primary   int32 `nbt:"primary"`   // TAG_Int(4) = 0
	Secondary int32 `nbt:"secondary"` // TAG_Int(4) = 0
	general.Global
}

// ID ...
func (*Beacon) ID() string {
	return IDBeacon
}

func (b *Beacon) Marshal(io protocol.IO) {
	b.Global.Marshal(io)
	io.Varint32(&b.Primary)
	io.Varint32(&b.Secondary)
}

func (b *Beacon) ToNBT() map[string]any {
	return slices.MergeMaps(
		b.Global.ToNBT(),
		map[string]any{
			"primary":   b.Primary,
			"secondary": b.Secondary,
		},
	)
}

func (b *Beacon) FromNBT(x map[string]any) {
	b.Global.FromNBT(x)
	b.Primary = x["primary"].(int32)
	b.Secondary = x["secondary"].(int32)
}
