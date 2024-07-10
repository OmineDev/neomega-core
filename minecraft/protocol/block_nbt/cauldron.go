package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 炼药锅
type Cauldron struct {
	Items       protocol.ItemList `nbt:"Items"`       // TAG_List[TAG_Compound] (9[10])
	PotionId    int16             `nbt:"PotionId"`    // TAG_Short(3) = -1
	PotionType  int16             `nbt:"PotionType"`  // TAG_Short(3) = -1
	CustomColor int32             `nbt:"CustomColor"` // TAG_Int(4) = 0
	general.Global
}

// ID ...
func (*Cauldron) ID() string {
	return IDCauldron
}

func (c *Cauldron) Marshal(io protocol.IO) {
	c.Global.Marshal(io)
	c.Items.Marshal(io)
	io.Varint16(&c.PotionId)
	io.Varint16(&c.PotionType)
	io.Varint32(&c.CustomColor)
}

func (c *Cauldron) ToNBT() map[string]any {
	return utils.MergeMaps(
		c.Global.ToNBT(),
		map[string]any{
			"Items":       c.Items.ToNBT(),
			"PotionId":    c.PotionId,
			"PotionType":  c.PotionType,
			"CustomColor": c.CustomColor,
		},
	)
}

func (c *Cauldron) FromNBT(x map[string]any) {
	c.Global.FromNBT(x)
	c.Items.FromNBT(x["Items"].([]any))
	c.PotionId = x["PotionId"].(int16)
	c.PotionType = x["PotionType"].(int16)
	c.CustomColor = x["CustomColor"].(int32)
}
