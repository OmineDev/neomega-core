package block_nbt

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/block_nbt/general"
	"github.com/OmineDev/neomega-core/utils/slices"
)

// 移动的方块
type MovingBlock struct {
	MovingBlock      map[string]any                    `nbt:"movingBlock"`      // TAG_Compound(10)
	MovingBlockExtra map[string]any                    `nbt:"movingBlockExtra"` // TAG_Compound(10)
	PistonPosX       int32                             `nbt:"pistonPosX"`       // TAG_Int(4) = 0
	PistonPosY       int32                             `nbt:"pistonPosY"`       // TAG_Int(4) = 0
	PistonPosZ       int32                             `nbt:"pistonPosZ"`       // TAG_Int(4) = 0
	Expanding        byte                              `nbt:"expanding"`        // Not used; TAG_Byte(1) = 0 or 1 (Boolean)
	MovingEntity     protocol.Optional[map[string]any] `nbt:"movingEntity"`     // TAG_Compound(10)
	general.Global
}

// ID ...
func (*MovingBlock) ID() string {
	return IDMovingBlock
}

func (m *MovingBlock) Marshal(io protocol.IO) {
	m.Global.Marshal(io)
	io.NBTWithLength(&m.MovingBlock)
	io.NBTWithLength(&m.MovingBlockExtra)
	io.Varint32(&m.PistonPosX)
	io.Varint32(&m.PistonPosY)
	io.Varint32(&m.PistonPosZ)
	protocol.OptionalFunc(io, &m.MovingEntity, io.NBTWithLength)
}

func (m *MovingBlock) ToNBT() map[string]any {
	var temp map[string]any
	if movingEntity, has := m.MovingEntity.Value(); has {
		temp = map[string]any{
			"movingEntity": movingEntity,
		}
	}
	return slices.MergeMaps(
		m.Global.ToNBT(),
		map[string]any{
			"movingBlock":      m.MovingBlock,
			"movingBlockExtra": m.MovingBlockExtra,
			"pistonPosX":       m.PistonPosX,
			"pistonPosY":       m.PistonPosY,
			"pistonPosZ":       m.PistonPosZ,
			"expanding":        m.Expanding,
		},
		temp,
	)
}

func (m *MovingBlock) FromNBT(x map[string]any) {
	m.Global.FromNBT(x)
	m.MovingBlock = x["movingBlock"].(map[string]any)
	m.MovingBlockExtra = x["movingBlockExtra"].(map[string]any)
	m.PistonPosX = x["pistonPosX"].(int32)
	m.PistonPosY = x["pistonPosY"].(int32)
	m.PistonPosZ = x["pistonPosZ"].(int32)
	m.Expanding = x["expanding"].(byte)

	if movingEntity, has := x["movingEntity"].(map[string]any); has {
		m.MovingEntity = protocol.Optional[map[string]any]{Set: true, Val: movingEntity}
	}
}
