package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 箱子
type Chest struct {
	Findable byte              `nbt:"Findable"` // TAG_Byte(1) = 0
	Items    protocol.ItemList `nbt:"Items"`    // TAG_List[TAG_Compound] (9[10])

	HasPair     byte  // Not a TAG, but a mark used to decide how to decode these four fields.
	Pairlead    byte  `nbt:"pairlead"`    // TAG_Byte(1) = 0
	Pairx       int32 `nbt:"pairx"`       // TAG_Int(4) = 0
	Pairz       int32 `nbt:"Pairz"`       // TAG_Int(4) = 0
	ForceUnpair byte  `nbt:"forceunpair"` // TAG_Byte(1) = 1

	CustomSize protocol.Optional[int16] `nbt:"CustomSize"` // TAG_Short(3) = 0

	general.Loot
	general.Global
}

// ID ...
func (c *Chest) ID() string {
	return IDChest
}

func (c *Chest) Marshal(io protocol.IO) {
	c.Loot.Marshal(io)
	c.Global.Marshal(io)

	io.Uint8(&c.Pairlead)
	io.Uint8(&c.HasPair)
	if c.HasPair == 1 {
		io.Varint32(&c.Pairx)
		io.Varint32(&c.Pairz)
	} else {
		io.Uint8(&c.ForceUnpair)
	}

	protocol.OptionalFunc(io, &c.CustomSize, io.Varint16)
	c.Items.Marshal(io)
	io.Uint8(&c.Findable)
}

func (c *Chest) ToNBT() map[string]any {
	var pair map[string]any
	var customSize map[string]any

	if c.HasPair == 1 {
		pair = map[string]any{
			"pairlead": c.Pairlead,
			"pairx":    c.Pairx,
			"pairz":    c.Pairz,
		}
	} else if c.ForceUnpair == 1 {
		pair = map[string]any{
			"forceunpair": c.ForceUnpair,
		}
	}

	if data, has := c.CustomSize.Value(); has {
		customSize = map[string]any{
			"CustomSize": data,
		}
	}

	return utils.MergeMaps(
		c.Global.ToNBT(),
		map[string]any{
			"Findable": c.Findable,
			"Items":    c.Items.ToNBT(),
		},
		c.Loot.ToNBT(), pair, customSize,
	)
}

func (c *Chest) FromNBT(x map[string]any) {
	c.Global.FromNBT(x)
	c.Findable = x["Findable"].(byte)
	c.Items.FromNBT(x["Items"].([]any))
	c.Loot.FromNBT(x)

	if pairlead, has := x["pairlead"].(byte); has {
		c.HasPair = 1
		c.Pairlead = pairlead
		c.Pairx = x["pairx"].(int32)
		c.Pairz = x["pairz"].(int32)
	} else {
		c.ForceUnpair, _ = x["ForceUnpair"].(byte)
	}
	if customSize, has := x["CustomSize"].(int16); has {
		c.CustomSize = protocol.Optional[int16]{Set: true, Val: customSize}
	}
}
