package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 磁石
type Lodestone struct {
	TrackingHandle protocol.Optional[int32] `nbt:"trackingHandle"` // TAG_Int(4) = 0
	general.Global
}

// ID ...
func (*Lodestone) ID() string {
	return IDLodestone
}

func (l *Lodestone) Marshal(io protocol.IO) {
	l.Global.Marshal(io)
	protocol.OptionalFunc(io, &l.TrackingHandle, io.Varint32)
}

func (l *Lodestone) ToNBT() map[string]any {
	var temp map[string]any
	if trackingHandle, has := l.TrackingHandle.Value(); has {
		temp = map[string]any{
			"trackingHandle": trackingHandle,
		}
	}
	return slices.MergeMaps(
		l.Global.ToNBT(),
		temp,
	)
}

func (l *Lodestone) FromNBT(x map[string]any) {
	l.Global.FromNBT(x)

	if trackingHandle, has := x["trackingHandle"].(int32); has {
		l.TrackingHandle = protocol.Optional[int32]{Set: true, Val: trackingHandle}
	}
}
