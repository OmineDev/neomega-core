package general

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 描述 发射器 和 投掷器 的通用字段
type Dispenser struct {
	Items protocol.ItemList `nbt:"Items"`      // TAG_List[TAG_Compound] (9[10])
	Name  string            `nbt:"CustomName"` // TAG_String(8) = ""
	Loot
	Global
}

func (d *Dispenser) Marshal(r protocol.IO) {
	protocol.Single(r, &d.Loot)
	protocol.Single(r, &d.Global)
	protocol.Single(r, &d.Items)
	r.String(&d.Name)
}

func (d *Dispenser) ToNBT() map[string]any {
	if len(d.Name) > 0 {
		temp := d.CustomName
		defer func() {
			d.CustomName = temp
		}()
		d.CustomName = d.Name
	}
	return slices_wrapper.MergeMaps(
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
