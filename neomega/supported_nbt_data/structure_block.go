package supported_nbt_data

import (
	"strings"

	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/mitchellh/mapstructure"
)

type StructureBlockSupportedData struct {
	DataField         string  `mapstructure:"dataField"`
	IgnoreEntities    uint8   `mapstructure:"ignoreEntities"`
	IncludePlayers    uint8   `mapstructure:"includePlayers"`
	Integrity         float32 `mapstructure:"integrity"`
	Mirror            uint8   `mapstructure:"mirror"`
	RedstoneSaveMode  int32   `mapstructure:"redstoneSaveMode"`
	IgnoreBlocks      int32   `mapstructure:"removeBlocks"`
	Rotation          uint8   `mapstructure:"rotation"`
	ShowBoundingBox   uint8   `mapstructure:"showBoundingBox"`
	AnimationMode     uint8   `mapstructure:"animationMode"`
	AnimationDuration float32 `mapstructure:"animationSeconds"`
	Seed              int64   `mapstructure:"seed"`

	XStructureSize int32 `mapstructure:"xStructureSize"`
	YStructureSize int32 `mapstructure:"yStructureSize"`
	ZStructureSize int32 `mapstructure:"zStructureSize"`

	XStructureOffset int32 `mapstructure:"xStructureOffset"`
	YStructureOffset int32 `mapstructure:"yStructureOffset"`
	ZStructureOffset int32 `mapstructure:"zStructureOffset"`

	StructureName      string `mapstructure:"structureName"`
	StructureBlockType int32
}

func (s *StructureBlockSupportedData) FromNBT(nbt map[string]any) {
	mapstructure.Decode(nbt, s)
}

func (s *StructureBlockSupportedData) LoadTypeFromString(state string) {
	if strings.Contains(state, "save") {
		s.StructureBlockType = packet.StructureBlockSave
	} else if strings.Contains(state, "load") {
		s.StructureBlockType = packet.StructureBlockLoad
	} else if strings.Contains(state, "corner") {
		s.StructureBlockType = packet.StructureBlockCorner
	} else if strings.Contains(state, "data") {
		s.StructureBlockType = packet.StructureBlockData
	}
}
