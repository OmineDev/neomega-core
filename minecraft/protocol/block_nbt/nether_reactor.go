package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/utils"
)

// 下界反应核
type NetherReactor struct {
	HasFinished   byte  `nbt:"HasFinished"`   // TAG_Byte(1) = 0
	IsInitialized byte  `nbt:"IsInitialized"` // TAG_Byte(1) = 0
	Progress      int16 `nbt:"Progress"`      // TAG_Short(3) = 0
	general.Global
}

// ID ...
func (*NetherReactor) ID() string {
	return IDNetherReactor
}

func (n *NetherReactor) Marshal(io protocol.IO) {
	n.Global.Marshal(io)
	io.Uint8(&n.IsInitialized)
	io.Varint16(&n.Progress)
	io.Uint8(&n.HasFinished)
}

func (n *NetherReactor) ToNBT() map[string]any {
	return utils.MergeMaps(
		n.Global.ToNBT(),
		map[string]any{
			"HasFinished":   n.HasFinished,
			"IsInitialized": n.IsInitialized,
			"Progress":      n.Progress,
		},
	)
}

func (n *NetherReactor) FromNBT(x map[string]any) {
	n.Global.FromNBT(x)
	n.HasFinished = x["HasFinished"].(byte)
	n.IsInitialized = x["IsInitialized"].(byte)
	n.Progress = x["Progress"].(int16)
}
