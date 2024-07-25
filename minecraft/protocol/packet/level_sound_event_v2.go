package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// Netease packet
type LevelSoundEventV2 struct {
	// Netease
	SoundType uint8
	// Netease
	Posistion mgl32.Vec3
	// Netease
	ExtraData int32
	// Netease
	EntityIdentifier string
	// Netease
	IsBabyMob bool
	// Netease
	IsGlobal bool
}

// ID ...
func (*LevelSoundEventV2) ID() uint32 {
	return IDLevelSoundEventV2
}

func (pk *LevelSoundEventV2) Marshal(io protocol.IO) {
	io.Uint8(&pk.SoundType)
	io.Vec3(&pk.Posistion)
	io.Varint32(&pk.ExtraData)
	io.String(&pk.EntityIdentifier)
	io.Bool(&pk.IsBabyMob)
	io.Bool(&pk.IsGlobal)
}
