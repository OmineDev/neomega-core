package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 唱片机
type Jukebox struct {
	RecordItem protocol.Optional[protocol.Item] `nbt:"RecordItem"` // TAG_Compound(10)
	general.Global
}

// ID ...
func (j *Jukebox) ID() string {
	return IDJukebox
}

func (j *Jukebox) Marshal(io protocol.IO) {
	j.Global.Marshal(io)
	protocol.OptionalMarshaler(io, &j.RecordItem)
}

func (j *Jukebox) ToNBT() map[string]any {
	var temp map[string]any
	if recordItem, has := j.RecordItem.Value(); has {
		temp = map[string]any{
			"RecordItem": recordItem.ToNBT(),
		}
	}
	return utils.MergeMaps(
		j.Global.ToNBT(),
		temp,
	)
}

func (j *Jukebox) FromNBT(x map[string]any) {
	j.Global.FromNBT(x)

	if recordItem, has := x["RecordItem"].(map[string]any); has {
		new := protocol.Item{}
		new.FromNBT(recordItem)
		j.RecordItem = protocol.Optional[protocol.Item]{Set: true, Val: new}
	}
}
