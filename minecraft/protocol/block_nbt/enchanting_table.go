package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 附魔台
type EnchantingTable struct {
	Rotation float32 `nbt:"rott"`       // TAG_Float(6) = 0
	Name     string  `nbt:"CustomName"` // TAG_String(8) = ""
	general.Global
}

// ID ...
func (*EnchantingTable) ID() string {
	return IDEnchantingTable
}

func (e *EnchantingTable) Marshal(io protocol.IO) {
	protocol.Single(io, &e.Global)
	io.String(&e.Name)
	io.Float32(&e.Rotation)
}

func (e *EnchantingTable) ToNBT() map[string]any {
	if len(e.Name) > 0 {
		temp := e.CustomName
		defer func() {
			e.CustomName = temp
		}()
		e.CustomName = e.Name
	}
	return slices.MergeMaps(
		e.Global.ToNBT(),
		map[string]any{
			"rott": e.Rotation,
		},
	)
}

func (e *EnchantingTable) FromNBT(x map[string]any) {
	e.Global.FromNBT(x)
	e.Rotation = x["rott"].(float32)
}
