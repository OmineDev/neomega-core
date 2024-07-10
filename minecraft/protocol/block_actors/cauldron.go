package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 炼药锅
type Cauldron struct {
	general.BlockActor
	Items       protocol.ItemList `nbt:"Items"`       // TAG_List[TAG_Compound] (9[10])
	PotionId    int16             `nbt:"PotionId"`    // TAG_Short(3) = -1
	PotionType  int16             `nbt:"PotionType"`  // TAG_Short(3) = -1
	CustomColor int32             `nbt:"CustomColor"` // TAG_Int(4) = 0
}

// ID ...
func (*Cauldron) ID() string {
	return IDCauldron
}

func (c *Cauldron) Marshal(io protocol.IO) {
	protocol.Single(io, &c.BlockActor)
	protocol.Single(io, &c.Items)
	io.Varint16(&c.PotionId)
	io.Varint16(&c.PotionType)
	io.Varint32(&c.CustomColor)
}

func (c *Cauldron) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		c.BlockActor.ToNBT(),
		map[string]any{
			"Items":       c.Items.ToNBT(),
			"PotionId":    c.PotionId,
			"PotionType":  c.PotionType,
			"CustomColor": c.CustomColor,
		},
	)
}

func (c *Cauldron) FromNBT(x map[string]any) {
	c.BlockActor.FromNBT(x)
	c.Items.FromNBT(x["Items"].([]any))
	c.PotionId = x["PotionId"].(int16)
	c.PotionType = x["PotionType"].(int16)
	c.CustomColor = x["CustomColor"].(int32)
}
