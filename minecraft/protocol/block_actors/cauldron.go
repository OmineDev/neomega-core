package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
)

// 炼药锅
type Cauldron struct {
	general.BlockActor
	Items       []protocol.ItemWithSlot `mapstructure:"Items"`       // TAG_List[TAG_Compound] (9[10])
	PotionId    int16                   `mapstructure:"PotionId"`    // TAG_Short(3) = -1
	PotionType  int16                   `mapstructure:"PotionType"`  // TAG_Short(3) = -1
	CustomColor int32                   `mapstructure:"CustomColor"` // TAG_Int(4) = 0
}

// ID ...
func (*Cauldron) ID() string {
	return IDCauldron
}

func (c *Cauldron) Marshal(io protocol.IO) {
	protocol.Single(io, &c.BlockActor)
	io.ItemList(&c.Items)
	io.Varint16(&c.PotionId)
	io.Varint16(&c.PotionType)
	io.Varint32(&c.CustomColor)
}
