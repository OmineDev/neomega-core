package packet

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"

	"github.com/go-gl/mathgl/mgl32"
)

// Netease packet
type LevelSoundEventV1 struct {
	// Netease
	SoundType uint8
	// Netease
	Posistion mgl32.Vec3
	// Netease
	ExtraData int32
	// Netease
	Pitch int32
	// Netease
	IsBabyMob bool
	// Netease
	IsGlobal bool
}

// ID ...
func (*LevelSoundEventV1) ID() uint32 {
	return IDLevelSoundEventV1
}

func (pk *LevelSoundEventV1) Marshal(io protocol.IO) {
	io.Uint8(&pk.SoundType)
	io.Vec3(&pk.Posistion)
	io.Varint32(&pk.ExtraData)
	io.Varint32(&pk.Pitch)
	io.Bool(&pk.IsBabyMob)
	io.Bool(&pk.IsGlobal)
}
