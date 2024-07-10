package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/fields"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 化合物创建器
type ChemistryTable struct {
	Item protocol.Optional[fields.ChemistryTableItem]
	general.Global
}

// ID ...
func (*ChemistryTable) ID() string {
	return IDChemistryTable
}

func (c *ChemistryTable) Marshal(io protocol.IO) {
	c.Global.Marshal(io)
	protocol.OptionalMarshaler(io, &c.Item)
}

func (c *ChemistryTable) ToNBT() map[string]any {
	var temp map[string]any
	if item, has := c.Item.Value(); has {
		temp = item.ToNBT()
	}
	return slices.MergeMaps(
		c.Global.ToNBT(), temp,
	)
}

func (c *ChemistryTable) FromNBT(x map[string]any) {
	c.Global.FromNBT(x)

	new := fields.ChemistryTableItem{}
	if new.CheckExist(x) {
		new.FromNBT(x)
		c.Item = protocol.Optional[fields.ChemistryTableItem]{Set: true, Val: new}
	}
}
