package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 可疑的方块
type BrushableBlock struct {
	LootTable      string `nbt:"LootTable"`       // Not used; TAG_String(8) = "loot_tables/entities/empty_brushable_block.json"
	LootTableSeed  int32  `nbt:"LootTableSeed"`   // Not used; TAG_Int(4) = 0
	BrushCount     int32  `nbt:"brush_count"`     // Not used; TAG_Int(4) = 0
	BrushDirection byte   `nbt:"brush_direction"` // Not used; AG_Byte(1) = 6
	general.Global
}

// ID ...
func (*BrushableBlock) ID() string {
	return IDBrushableBlock
}

func (b *BrushableBlock) Marshal(io protocol.IO) {
	protocol.Single(io, &b.Global)
}

func (b *BrushableBlock) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		b.Global.ToNBT(),
		map[string]any{
			"LootTable":       b.LootTable,
			"LootTableSeed":   b.LootTableSeed,
			"brush_count":     b.BrushCount,
			"brush_direction": b.BrushDirection,
		},
	)
}

func (b *BrushableBlock) FromNBT(x map[string]any) {
	b.Global.FromNBT(x)
	b.LootTable = x["LootTable"].(string)
	b.LootTableSeed = x["LootTableSeed"].(int32)
	b.BrushCount = x["brush_count"].(int32)
	b.BrushDirection = x["brush_direction"].(byte)
}
