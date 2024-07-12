package block_actors

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	general "github.com/OmineDev/neomega-core/minecraft/protocol/block_actors/general_actors"
)

// 活塞臂
type PistonArm struct {
	general.BlockActor
	AttachedBlocks []int32 `mapstructure:"AttachedBlocks"` // TAG_List[TAG_Int] (9[4])
	BreakBlocks    []int32 `mapstructure:"BreakBlocks"`    // TAG_List[TAG_Int] (9[4])
	LastProgress   float32 `mapstructure:"LastProgress"`   // TAG_Float(6) = 0
	NewState       byte    `mapstructure:"NewState"`       // TAG_Byte(1) = 0
	Progress       float32 `mapstructure:"Progress"`       // TAG_Float(6) = 0
	State          byte    `mapstructure:"State"`          // TAG_Byte(1) = 0
	Sticky         byte    `mapstructure:"Sticky"`         // TAG_Byte(1) = 0
}

// ID ...
func (*PistonArm) ID() string {
	return IDPistonArm
}

func (p *PistonArm) Marshal(io protocol.IO) {
	protocol.Single(io, &p.BlockActor)
	io.Float32(&p.Progress)
	io.Float32(&p.LastProgress)
	protocol.NBTInt(&p.State, io.Varuint32)
	protocol.NBTInt(&p.NewState, io.Varuint32)
	io.Uint8(&p.Sticky)
	io.PistonAttachedBlocks(&p.AttachedBlocks)
	io.PistonAttachedBlocks(&p.BreakBlocks)
}
