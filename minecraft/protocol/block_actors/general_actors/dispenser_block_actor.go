package general

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

// 描述 发射器 和 投掷器 的通用字段
type DispenserBlockActor struct {
	RandomizableBlockActor
	Items []protocol.ItemWithSlot `mapstructure:"Items"` // TAG_List[TAG_Compound] (9[10])
}

func (d *DispenserBlockActor) Marshal(r protocol.IO) {
	var name string = d.CustomName

	protocol.Single(r, &d.RandomizableBlockActor)
	r.ItemList(&d.Items)
	r.String(&name)

	if len(name) > 0 {
		d.CustomName = name
	}
}
