package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

// Netease Packet
type SyncSkin struct {
	// Netease: skin info
	PlayerData []protocol.NeteasePlayerData
	// Netease: player uids
	PlayerUIDs []string
	// Netease
	Unknown []string
	// Netease: skin item ids
	SkinItemIDs []string
	// Netease: skin
	Skin protocol.Skin
}

// ID ...
func (*SyncSkin) ID() uint32 {
	return IDSyncSkin
}

func (pk *SyncSkin) Marshal(io protocol.IO) {
	protocol.Slice(io, &pk.PlayerData)
	protocol.FuncSliceOfLen(io, uint32(len(pk.PlayerData)), &pk.PlayerUIDs, io.String)
	protocol.FuncSliceOfLen(io, uint32(len(pk.PlayerData)), &pk.Unknown, io.String)
	protocol.FuncSliceOfLen(io, uint32(len(pk.PlayerData)), &pk.SkinItemIDs, io.String)
	protocol.Single(io, &pk.Skin)
}
