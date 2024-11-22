package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
)

// Netease Packet
type ChangeBiome struct {
	// Netease
	Unknown1 string
	// Netease
	Unknown2 float32
	// Netease
	Unknown3 float32
	// Netease
	Unknown4 float32
	// Netease
	Unknown5 float32
	// Netease
	Unknown6 bool
}

// ID ...
func (*ChangeBiome) ID() uint32 {
	return IDChangeBiome
}

func (pk *ChangeBiome) Marshal(io protocol.IO) {
	io.String(&pk.Unknown1)
	io.Float32(&pk.Unknown2)
	io.Float32(&pk.Unknown3)
	io.Float32(&pk.Unknown4)
	io.Float32(&pk.Unknown5)
	io.Bool(&pk.Unknown6)
}
