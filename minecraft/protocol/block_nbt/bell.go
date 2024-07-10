package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 钟
type Bell struct {
	Direction int32 `nbt:"Direction"` // TAG_Int(4) = 255
	Ringing   byte  `nbt:"Ringing"`   // TAG_Byte(1) = 0
	Ticks     int32 `nbt:"Ticks"`     // TAG_Int(4) = 18
	general.Global
}

// ID ...
func (*Bell) ID() string {
	return IDBell
}

func (b *Bell) Marshal(io protocol.IO) {
	b.Global.Marshal(io)
	io.Uint8(&b.Ringing)
	io.Varint32(&b.Ticks)
	io.Varint32(&b.Direction)
}

func (b *Bell) ToNBT() map[string]any {
	return slices.MergeMaps(
		b.Global.ToNBT(),
		map[string]any{
			"Direction": b.Direction,
			"Ringing":   b.Ringing,
			"Ticks":     b.Ticks,
		},
	)
}

func (b *Bell) FromNBT(x map[string]any) {
	b.Global.FromNBT(x)
	b.Direction = x["Direction"].(int32)
	b.Ringing = x["Ringing"].(byte)
	b.Ticks = x["Ticks"].(int32)
}
