package general

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 描述 熔炉、高炉、烟熏炉 的通用字段
type Furnace struct {
	BurnDuration int16             `nbt:"BurnDuration"` // TAG_Short(3) = 0
	BurnTime     int16             `nbt:"BurnTime"`     // TAG_Short(3) = 0
	CookTime     int16             `nbt:"CookTime"`     // TAG_Short(3) = 0
	Items        protocol.ItemList `nbt:"Items"`        // TAG_List[TAG_Compound] (9[10])
	StoredXPInt  int32             `nbt:"StoredXPInt"`  // TAG_Int(4) = 0
	Global
}

func (f *Furnace) Marshal(r protocol.IO) {
	f.Global.Marshal(r)
	r.Varint16(&f.BurnTime)
	r.Varint16(&f.CookTime)
	r.Varint16(&f.BurnDuration)
	r.Varint32(&f.StoredXPInt)
	f.Items.Marshal(r)
}

func (f *Furnace) ToNBT() map[string]any {
	return slices.MergeMaps(
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
