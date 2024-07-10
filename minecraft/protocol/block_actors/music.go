package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 音符盒
type Music struct {
	general.BlockActor
	Note uint32 `nbt:"note"` // * TAG_Byte(1) = 0
}

// ID ...
func (*Music) ID() string {
	return IDMusic
}

func (n *Music) Marshal(io protocol.IO) {
	protocol.Single(io, &n.BlockActor)
	io.Varuint32(&n.Note)
}

func (n *Music) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		n.BlockActor.ToNBT(),
		map[string]any{
			"note": byte(n.Note),
		},
	)
}

func (n *Music) FromNBT(x map[string]any) {
	n.BlockActor.FromNBT(x)
	n.Note = uint32(x["note"].(byte))
}
