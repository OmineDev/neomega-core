package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

// Netease packet
type ConfirmSkin struct {
	// Netease: skin info
	SkinInfo []protocol.ConfirmSkinUnknownEntry
	// Netease: launcher uids
	Uids []string
	// Netease
	Unknown3 []string
}

// ID ...
func (*ConfirmSkin) ID() uint32 {
	return IDConfirmSkin
}

func (pk *ConfirmSkin) Marshal(io protocol.IO) {
	protocol.SliceVaruint32Length(io, &pk.SkinInfo)
	protocol.FuncSliceOfLen(io, uint32(len(pk.SkinInfo)), &pk.Uids, io.String)
	protocol.FuncSliceOfLen(io, uint32(len(pk.SkinInfo)), &pk.Unknown3, io.String)
}
