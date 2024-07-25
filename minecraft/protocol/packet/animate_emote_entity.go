package packet

import "github.com/OmineDev/neomega-core/minecraft/protocol"

// Netease Packet
type AnimateEmoteEntity struct {
	// Netease
	Animation string
	// Netease
	NextState string
	// Netease
	StopExpression string
	// Netease
	StopExpressionVersion int32
	// Netease
	Controller string
	// Netease
	BlendOutTime float32
	// Netease
	RuntimeEntityIds []uint64
}

// ID ...
func (*AnimateEmoteEntity) ID() uint32 {
	return IDAnimateEmoteEntity
}

func (pk *AnimateEmoteEntity) Marshal(io protocol.IO) {
	io.String(&pk.Animation)
	io.String(&pk.NextState)
	io.String(&pk.StopExpression)
	io.Int32(&pk.StopExpressionVersion)
	io.String(&pk.Controller)
	io.Float32(&pk.BlendOutTime)
	protocol.FuncSliceVarint32Length(io, &pk.RuntimeEntityIds, io.Varuint64)
}
