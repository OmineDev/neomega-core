package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/fields"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
)

// 旗帜
type Banner struct {
	general.BlockActor
	Base     int32                   `mapstructure:"Base"`               // TAG_Int(4) = 0
	Patterns []fields.BannerPatterns `mapstructure:"Patterns,omitempty"` // TAG_List[TAG_Compound] (9[10])
	Type     int32                   `mapstructure:"Type"`               // TAG_Int(4) = 0
}

// ID ...
func (*Banner) ID() string {
	return IDBanner
}

func (b *Banner) Marshal(io protocol.IO) {
	protocol.Single(io, &b.BlockActor)
	protocol.NBTInt(&b.Base, io.Varuint32)
	io.Varint32(&b.Type)
	protocol.SliceVarint16Length(io, &b.Patterns)
}
