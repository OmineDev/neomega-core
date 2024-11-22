package protocol

// Netease
type NeteaseBiomeData struct {
	// Netease
	Flag uint8
	// Netease
	Unknown1 int32
	// Netease
	Unknown2 string
	// Netease
	Pos1 BlockPos
	// Netease
	Pos2 BlockPos
}

// Marshal encodes/decodes an NeteaseBiomeData.
func (x *NeteaseBiomeData) Marshal(r IO) {
	r.Uint8(&x.Flag)
	r.Varint32(&x.Unknown1)
	r.String(&x.Unknown2)
	r.UBlockPos(&x.Pos1)
	if x.Flag == 1 {
		r.UBlockPos(&x.Pos2)
	}
}
