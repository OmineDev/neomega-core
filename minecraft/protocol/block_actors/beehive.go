package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/fields"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
)

// 蜂箱
type Beehive struct {
	general.BlockActor `mapstructure:",squash"`
	Occupants          []any `mapstructure:"Occupants,omitempty"` // TAG_List[TAG_Compound] (9[10])
	ShouldSpawnBees    byte  `mapstructure:"ShouldSpawnBees"`     // TAG_Byte(1) = 0
}

// ID ...
func (*Beehive) ID() string {
	return IDBeehive
}

func (b *Beehive) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
	protocol.NBTSliceVarint16Length(io, &b.Occupants, &[]fields.BeehiveOccupants{})
	io.Uint8(&b.ShouldSpawnBees)
}
