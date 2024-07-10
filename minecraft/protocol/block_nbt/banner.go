package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/fields"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 旗帜
type Banner struct {
	Base     uint32                  `nbt:"Base"`     // * TAG_Int(4) = 0
	Patterns []fields.BannerPatterns `nbt:"Patterns"` // TAG_List[TAG_Compound] (9[10])
	Type     int32                   `nbt:"Type"`     // TAG_Int(4) = 0
	general.Global
}

// ID ...
func (*Banner) ID() string {
	return IDBanner
}

func (b *Banner) Marshal(io protocol.IO) {
	protocol.Single(io, &b.Global)
	io.Varuint32(&b.Base)
	io.Varint32(&b.Type)
	protocol.SliceVarint16Length(io, &b.Patterns)
}

func (b *Banner) ToNBT() map[string]any {
	var temp map[string]any
	if len(b.Patterns) > 0 {
		new := make([]any, len(b.Patterns))
		for key, value := range b.Patterns {
			new[key] = value.ToNBT()
		}
		temp = map[string]any{
			"Patterns": new,
		}
	}
	return slices_wrapper.MergeMaps(
		b.Global.ToNBT(),
		map[string]any{
			"Base": int32(b.Base),
			"Type": b.Type,
		},
		temp,
	)
}

func (b *Banner) FromNBT(x map[string]any) {
	b.Global.FromNBT(x)
	b.Base = uint32(x["Base"].(int32))
	b.Type = x["Type"].(int32)

	if patterns, has := x["Patterns"].([]any); has {
		b.Patterns = make([]fields.BannerPatterns, len(patterns))
		for key, value := range patterns {
			new := fields.BannerPatterns{}
			new.FromNBT(value.(map[string]any))
			b.Patterns[key] = new
		}
	}
}
