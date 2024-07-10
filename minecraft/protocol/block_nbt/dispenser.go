package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 发射器
type Dispenser struct {
	Items protocol.ItemList `nbt:"Items"`      // TAG_List[TAG_Compound] (9[10])
	Name  string            `nbt:"CustomName"` // TAG_String(8) = ""
	general.Loot
	general.Global
}

// ID ...
func (*Dispenser) ID() string {
	return IDDispenser
}

func (d *Dispenser) Marshal(io protocol.IO) {
	d.Loot.Marshal(io)
	d.Global.Marshal(io)
	d.Items.Marshal(io)
	io.String(&d.Name)
}

func (d *Dispenser) ToNBT() map[string]any {
	if len(d.Name) > 0 {
		temp := d.CustomName
		defer func() {
			d.CustomName = temp
		}()
		d.CustomName = d.Name
	}
	return slices.MergeMaps(
		d.Global.ToNBT(),
		map[string]any{
			"Items": d.Items.ToNBT(),
		},
		d.Loot.ToNBT(),
	)
}

func (d *Dispenser) FromNBT(x map[string]any) {
	d.Global.FromNBT(x)
	d.Items.FromNBT(x["Items"].([]any))
	d.Loot.FromNBT(x)
}
