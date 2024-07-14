package area_request

import (
	"fmt"

	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"

	"github.com/mitchellh/mapstructure"
)

type StructureResponse struct {
	raw              *packet.StructureTemplateDataResponse
	decodedStructure *neomega.DecodedStructure
}

func newStructureResponse(r *packet.StructureTemplateDataResponse) neomega.StructureResponse {
	return &StructureResponse{
		raw: r,
	}
}

func (sr *StructureResponse) Raw() *packet.StructureTemplateDataResponse {
	return sr.raw
}

type StructureContent struct {
	Version   int32    `mapstructure:"format_version" nbt:"format_version"`
	Size      [3]int32 `mapstructure:"size" nbt:"size"`
	Origin    [3]int32 `mapstructure:"structure_world_origin" nbt:"structure_world_origin"`
	Structure struct {
		BlockIndices [2]interface{} `mapstructure:"block_indices" nbt:"block_indices"`
		//Entities     []map[string]interface{} `mapstructure:"entities"`
		Palette struct {
			Default struct {
				BlockPositionData map[string]struct {
					Nbt map[string]interface{} `mapstructure:"block_entity_data" nbt:"block_entity_data"`
				} `mapstructure:"block_position_data" nbt:"block_position_data"`
				BlockPalette []struct {
					Name   string                 `mapstructure:"name" nbt:"name"`
					States map[string]interface{} `mapstructure:"states" nbt:"states"`
					Value  int16                  `mapstructure:"val" nbt:"val"`
				} `mapstructure:"block_palette" nbt:"block_palette"`
			} `mapstructure:"default" nbt:"default"`
		} `mapstructure:"palette" nbt:"palette"`
	} `mapstructure:"structure" nbt:"structure"`
	decoded *neomega.DecodedStructure
}

func (structure *StructureContent) FromNBT(nbt map[string]any) error {
	err := mapstructure.Decode(nbt, &structure)
	if err != nil {
		return err
	}
	return nil
}
func (structure *StructureContent) Decode() *neomega.DecodedStructure {
	nbts := map[define.CubePos]map[string]interface{}{}
	for _, blockNbt := range structure.Structure.Palette.Default.BlockPositionData {
		x, y, z, ok := define.GetPosFromNBT(blockNbt.Nbt)
		if ok {
			nbts[define.CubePos{x, y, z}] = blockNbt.Nbt
		}
	}
	// BlockPalettes := make(map[string]*neomega.BlockPalettes)
	paletteLookUp := make([]uint32, len(structure.Structure.Palette.Default.BlockPalette))
	for paletteIdx, palette := range structure.Structure.Palette.Default.BlockPalette {
		rtid, _ := blocks.BlockNameAndStateToRuntimeID(palette.Name, palette.States)
		paletteLookUp[paletteIdx] = rtid
		// hashName := fmt.Sprintf("%v[%v]", palette.Name, chunk.PropsToStateString(palette.States, false))
		// BlockPalettes[hashName] = &neomega.BlockPalettes{
		// 	Name:   palette.Name,
		// 	States: palette.States,
		// 	Value:  palette.Value,
		// 	RTID:   rtid,
		// 	// NemcRtid:    chunk.AirRID,
		// }
	}
	var foreground, background []uint32
	{
		BlockIndices0, BlockIndices1 := (structure.Structure.BlockIndices[0]).([]int32), (structure.Structure.BlockIndices[1]).([]int32)
		foreground = make([]uint32, len(BlockIndices0))
		background = make([]uint32, len(BlockIndices1))
		_v := int32(0)
		for i, v := range BlockIndices0 {
			_v = v
			if _v != -1 {
				foreground[i] = paletteLookUp[_v]
			} else {
				foreground[i] = blocks.AIR_RUNTIMEID
			}
		}
		for i, v := range BlockIndices1 {
			_v = v
			if _v != -1 {
				background[i] = paletteLookUp[_v]
			} else {
				background[i] = blocks.AIR_RUNTIMEID
			}
		}
	}
	decodeStructure := &neomega.DecodedStructure{
		Version: structure.Version,
		Size: define.CubePos{
			int(structure.Size[0]), int(structure.Size[1]), int(structure.Size[2]),
		},
		Origin: define.CubePos{
			int(structure.Origin[0]), int(structure.Origin[1]), int(structure.Origin[2]),
		},
		ForeGround: foreground,
		BackGround: background,
		Nbts:       nbts,
		// BlockPalettes: BlockPalettes,
	}
	structure.decoded = decodeStructure
	return decodeStructure
}

func (sr *StructureResponse) Decode() (s *neomega.DecodedStructure, err error) {
	if !sr.raw.Success {
		return nil, fmt.Errorf("response get fail result")
	}
	if sr.decodedStructure != nil {
		return sr.decodedStructure, nil
	}
	structureData := sr.raw.StructureTemplate
	structure := &StructureContent{}
	err = structure.FromNBT(structureData)
	if err != nil {
		return nil, err
	}
	decodeStructure := structure.Decode()
	sr.decodedStructure = decodeStructure
	return decodeStructure, nil

}
