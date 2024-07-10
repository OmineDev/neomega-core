package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 信标
type Beacon struct {
	general.BlockActor
	Primary   int32 `nbt:"primary"`   // TAG_Int(4) = 0
	Secondary int32 `nbt:"secondary"` // TAG_Int(4) = 0
}

// ID ...
func (*Beacon) ID() string {
	return IDBeacon
}

func (b *Beacon) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
	io.Varint32(&b.Primary)
	io.Varint32(&b.Secondary)
}

func (b *Beacon) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		b.BlockActor.ToNBT(),
		map[string]any{
			"primary":   b.Primary,
			"secondary": b.Secondary,
		},
	)
}

func (b *Beacon) FromNBT(x map[string]any) {
	b.BlockActor.FromNBT(x)
	b.Primary = x["primary"].(int32)
	b.Secondary = x["secondary"].(int32)
}
