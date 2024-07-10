package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices_wrapper"
)

// 头颅
type Skull struct {
	DoingAnimation byte    `nbt:"DoingAnimation"` // * TAG_Byte(1) = 0
	MouthTickCount uint16  `nbt:"MouthTickCount"` // * TAG_Int(4) = 0
	Rotation       float32 `nbt:"Rotation"`       // TAG_Float(6) = 0
	SkullType      uint16  `nbt:"SkullType"`      // * TAG_Byte(1) = 0
	general.Global
}

// ID ...
func (*Skull) ID() string {
	return IDSkull
}

func (s *Skull) Marshal(io protocol.IO) {
	protocol.Single(io, &s.Global)
	io.Varuint16(&s.SkullType)
	io.Float32(&s.Rotation)
	io.Uint8(&s.DoingAnimation)
	io.Varuint16(&s.MouthTickCount)
}

func (s *Skull) ToNBT() map[string]any {
	return slices_wrapper.MergeMaps(
		s.Global.ToNBT(),
		map[string]any{
			"DoingAnimation": s.DoingAnimation,
			"MouthTickCount": int32(s.MouthTickCount),
			"Rotation":       s.Rotation,
			"SkullType":      byte(s.SkullType),
		},
	)
}

func (s *Skull) FromNBT(x map[string]any) {
	s.Global.FromNBT(x)
	s.DoingAnimation = x["DoingAnimation"].(byte)
	s.MouthTickCount = uint16(x["MouthTickCount"].(int32))
	s.Rotation = x["Rotation"].(float32)
	s.SkullType = uint16(x["SkullType"].(byte))
}
