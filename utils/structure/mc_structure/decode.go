package mc_structure

import (
	"time"

	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/utils/structure/pos_operations"
)

type DecodedStructure struct {
	Version    int32
	Size       define.CubePos
	Origin     define.CubePos
	ForeGround []uint32
	BackGround []uint32
	Nbts       map[define.CubePos]map[string]interface{}
}

func (d *DecodedStructure) NBTsInAbsolutePos() map[define.CubePos]map[string]interface{} {
	return d.Nbts
}

func (d *DecodedStructure) ForeGroundRtidNested() []uint32 {
	return d.ForeGround
}

func (d *DecodedStructure) BackGroundRtidNested() []uint32 {
	return d.BackGround
}

func (d *DecodedStructure) IndexOfRelativePos(pos define.CubePos) int {
	YZ := d.Size.Y() * d.Size.Z()
	return YZ*pos.X() + d.Size.Z()*pos.Y() + pos.Z()
}

func (d *DecodedStructure) BlockOfRelativePos(pos define.CubePos) (foreGround, backGround uint32) {
	idx := d.IndexOfRelativePos(pos)
	return d.ForeGround[idx], d.BackGround[idx]
}

func (structure *DecodedStructure) DumpToChunkProviderAbsolutePos(chunkProvider chunks.ChunkProvider) (err error) {
	background, foreground := structure.BackGround, structure.ForeGround
	rtid := uint32(0)
	chunkRangesX := pos_operations.RangeSplits(structure.Origin.X(), structure.Size[0], 16)
	chunkRangesZ := pos_operations.RangeSplits(structure.Origin.Z(), structure.Size[2], 16)
	var chunkData *chunks.ChunkWithAuxInfo
	for _, xRange := range chunkRangesX {
		for _, zRange := range chunkRangesZ {
			blockPos00 := define.CubePos{int(xRange[0]), int(structure.Origin.Y()), int(zRange[0])}
			chunkPos := define.ChunkPos{int32(blockPos00.X() >> 4), int32(blockPos00.Z() >> 4)}
			if chunkData != nil {
				if err = chunkProvider.Write(chunkData); err != nil {
					return err
				}
			}
			chunkData = chunkProvider.Get(chunkPos)
			if chunkData == nil {
				chunkData = &chunks.ChunkWithAuxInfo{
					Chunk:     chunk.New(blocks.AIR_RUNTIMEID, define.WorldRange),
					BlockNbts: make(map[define.CubePos]map[string]interface{}),
					SyncTime:  time.Now().Unix(),
					ChunkPos:  chunkPos,
				}
			}
			for x := xRange[0]; x < xRange[0]+xRange[1]; x++ {
				for z := zRange[0]; z < zRange[0]+zRange[1]; z++ {
					for y := structure.Origin.Y(); y < structure.Origin.Y()+structure.Size[1]; y++ {
						blockPos := define.CubePos{int(x), int(y), int(z)}
						iPos := blockPos.Sub(structure.Origin)
						i := iPos.Z() + int(iPos.Y())*structure.Size[2] + iPos.X()*(structure.Size[1]*structure.Size[2])
						rtid = background[i]
						if rtid != blocks.AIR_RUNTIMEID {
							chunkData.Chunk.SetBlock(uint8(blockPos.X())&0xF, int16(y), uint8(blockPos.Z())&0xF, 0, rtid)
						}
						rtid = foreground[i]
						if rtid != blocks.AIR_RUNTIMEID {
							chunkData.Chunk.SetBlock(uint8(blockPos.X())&0xF, int16(y), uint8(blockPos.Z())&0xF, 0, rtid)
						}
						// TODO: Check Block Offset or Block Pos
						nbt, found := structure.Nbts[blockPos]
						if found {
							chunkData.BlockNbts[blockPos] = nbt
						}
					}
				}
			}

		}
	}
	if chunkData != nil {
		if err = chunkProvider.Write(chunkData); err != nil {
			return err
		}
	}
	return nil
}
