package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/fields"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 化合物创建器
type ChemistryTable struct {
	general.BlockActor
	Item protocol.Optional[fields.ChemistryTableItem]
}

// ID ...
func (*ChemistryTable) ID() string {
	return IDChemistryTable
}

func (c *ChemistryTable) Marshal(io protocol.IO) {
	protocol.Single(io, &c.BlockActor)
	protocol.OptionalMarshaler(io, &c.Item)
}

func (c *ChemistryTable) ToNBT() map[string]any {
	var temp map[string]any
	if item, has := c.Item.Value(); has {
		temp = item.ToNBT()
	}
	return slices_wrapper.MergeMaps(
		c.BlockActor.ToNBT(), temp,
	)
}

func (c *ChemistryTable) FromNBT(x map[string]any) {
	c.BlockActor.FromNBT(x)

	new := fields.ChemistryTableItem{}
	if new.CheckExist(x) {
		new.FromNBT(x)
		c.Item = protocol.Optional[fields.ChemistryTableItem]{Set: true, Val: new}
	}
}
