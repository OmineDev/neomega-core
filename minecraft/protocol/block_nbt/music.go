package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 音符盒
type Music struct {
	Note uint32 `nbt:"note"` // * TAG_Byte(1) = 0
	general.Global
}

// ID ...
func (*Music) ID() string {
	return IDMusic
}

func (n *Music) Marshal(io protocol.IO) {
	protocol.Single(io, &n.Global)
	io.Varuint32(&n.Note)
}

func (n *Music) ToNBT() map[string]any {
	return slices.MergeMaps(
		n.Global.ToNBT(),
		map[string]any{
			"note": byte(n.Note),
		},
	)
}

func (n *Music) FromNBT(x map[string]any) {
	n.Global.FromNBT(x)
	n.Note = uint32(x["note"].(byte))
}
