package general

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// 描述部分容器的 战利品表
type Loot struct {
	LootTable     string `nbt:"LootTable"`     // TAG_String(8) = ""
	LootTableSeed int64  `nbt:"LootTableSeed"` // * TAG_Int(4) = 0
}

func (l *Loot) Marshal(r protocol.IO) {
	r.String(&l.LootTable)
	if len(l.LootTable) > 0 {
		r.Varint64(&l.LootTableSeed)
	}
}

func (l *Loot) ToNBT() map[string]any {
	if len(l.LootTable) > 0 {
		return map[string]any{
			"LootTable":     l.LootTable,
			"LootTableSeed": int32(l.LootTableSeed),
		}
	}
	return map[string]any{}
}

func (l *Loot) FromNBT(x map[string]any) {
	if lootTable, has := x["LootTable"].(string); has {
		l.LootTable = lootTable
		l.LootTableSeed = int64(x["LootTableSeed"].(int32))
	}
}
