package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 漏斗
type Hopper struct {
	Items            protocol.ItemList `nbt:"Items"`            // TAG_List[TAG_Compound] (9[10])
	TransferCooldown int32             `nbt:"TransferCooldown"` // TAG_Int(4) = 0
	MoveItemSpeed    int16             `nbt:"MoveItemSpeed"`    // TAG_Short(3) = 0
	general.Global
}

// ID ...
func (*Hopper) ID() string {
	return IDHopper
}

func (h *Hopper) Marshal(io protocol.IO) {
	protocol.Single(io, &h.Global)
	protocol.Single(io, &h.Items)
	io.Varint32(&h.TransferCooldown)
	io.Varint16(&h.MoveItemSpeed)
}

func (h *Hopper) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		h.Global.ToNBT(),
		map[string]any{
			"Items":            h.Items.ToNBT(),
			"TransferCooldown": h.TransferCooldown,
			"MoveItemSpeed":    h.MoveItemSpeed,
		},
	)
}

func (h *Hopper) FromNBT(x map[string]any) {
	h.Global.FromNBT(x)
	h.Items.FromNBT(x["Items"].([]any))
	h.TransferCooldown = x["TransferCooldown"].(int32)
	h.MoveItemSpeed = x["MoveItemSpeed"].(int16)
}
