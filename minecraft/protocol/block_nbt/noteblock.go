package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 音符盒
type NoteBlock struct {
	Note uint32 `nbt:"note"` // * TAG_Byte(1) = 0
	general.Global
}

// ID ...
func (*NoteBlock) ID() string {
	return IDNoteBlock
}

func (n *NoteBlock) Marshal(io protocol.IO) {
	n.Global.Marshal(io)
	io.Varuint32(&n.Note)
}

func (n *NoteBlock) ToNBT() map[string]any {
	return utils.MergeMaps(
		n.Global.ToNBT(),
		map[string]any{
			"note": byte(n.Note),
		},
	)
}

func (n *NoteBlock) FromNBT(x map[string]any) {
	n.Global.FromNBT(x)
	n.Note = uint32(x["note"].(byte))
}
