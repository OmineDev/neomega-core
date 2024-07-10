package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 床
type Bed struct {
	Color uint32 `nbt:"color"` // * TAG_Byte(1) = 0
	general.Global
}

// ID ...
func (*Bed) ID() string {
	return IDBed
}

func (b *Bed) Marshal(io protocol.IO) {
	b.Global.Marshal(io)
	io.Varuint32(&b.Color)
}

func (b *Bed) ToNBT() map[string]any {
	return utils.MergeMaps(
		b.Global.ToNBT(),
		map[string]any{
			"color": byte(b.Color),
		},
	)
}

func (b *Bed) FromNBT(x map[string]any) {
	b.Global.FromNBT(x)
	b.Color = uint32(x["color"].(byte))
}
