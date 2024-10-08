package packet

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// Netease packet
type StoreBuySucc struct {
	// Netease
	Data []byte
}

// ID ...
func (*StoreBuySucc) ID() uint32 {
	return IDStoreBuySucc
}

func (pk *StoreBuySucc) Marshal(io protocol.IO) {
	io.ByteSlice(&pk.Data)
}
