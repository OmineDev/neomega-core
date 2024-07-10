package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 熔炉
type Furnace struct {
	BurnDuration int16             `nbt:"BurnDuration"` // TAG_Short(3) = 0
	BurnTime     int16             `nbt:"BurnTime"`     // TAG_Short(3) = 0
	CookTime     int16             `nbt:"CookTime"`     // TAG_Short(3) = 0
	Items        protocol.ItemList `nbt:"Items"`        // TAG_List[TAG_Compound] (9[10])
	StoredXPInt  int32             `nbt:"StoredXPInt"`  // TAG_Int(4) = 0
	general.Global
}

// ID ...
func (*Furnace) ID() string {
	return IDFurnace
}

func (f *Furnace) Marshal(io protocol.IO) {
	f.Global.Marshal(io)
	io.Varint16(&f.BurnTime)
	io.Varint16(&f.CookTime)
	io.Varint16(&f.BurnDuration)
	io.Varint32(&f.StoredXPInt)
	f.Items.Marshal(io)
}

func (f *Furnace) ToNBT() map[string]any {
	return utils.MergeMaps(
		f.Global.ToNBT(),
		map[string]any{
			"BurnDuration": f.BurnDuration,
			"BurnTime":     f.BurnTime,
			"CookTime":     f.CookTime,
			"Items":        f.Items.ToNBT(),
			"StoredXPInt":  f.StoredXPInt,
		},
	)
}

func (f *Furnace) FromNBT(x map[string]any) {
	f.Global.FromNBT(x)
	f.BurnDuration = x["BurnDuration"].(int16)
	f.BurnTime = x["BurnTime"].(int16)
	f.CookTime = x["CookTime"].(int16)
	f.Items.FromNBT(x["Items"].([]any))
	f.StoredXPInt = x["StoredXPInt"].(int32)
}
