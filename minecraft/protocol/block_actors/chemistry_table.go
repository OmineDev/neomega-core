package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/fields"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
)

// 化合物创建器
type ChemistryTable struct {
	general.BlockActor         `mapstructure:",squash"`
	*fields.ChemistryTableItem `mapstructure:",omitempty"`
}

// ID ...
func (*ChemistryTable) ID() string {
	return IDChemistryTable
}

func (c *ChemistryTable) Marshal(io protocol.IO) {
	f := func() *fields.ChemistryTableItem {
		if c.ChemistryTableItem == nil {
			c.ChemistryTableItem = new(fields.ChemistryTableItem)
		}
		return c.ChemistryTableItem
	}

	protocol.Single(io, &c.BlockActor)
	protocol.NBTOptionalMarshaler(io, c.ChemistryTableItem, f, true)
}
