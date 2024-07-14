package neomega

import (
	"time"

	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type DecodedStructure struct {
	Version    int32
	Size       define.CubePos
	Origin     define.CubePos
	ForeGround []uint32
	BackGround []uint32
	Nbts       map[define.CubePos]map[string]interface{}
}

func (d *DecodedStructure) IndexOf(pos define.CubePos) int {
	YZ := d.Size.Y() * d.Size.Z()
	return YZ*pos.X() + d.Size.Z()*pos.Y() + pos.Z()
}

func (d *DecodedStructure) BlockOf(pos define.CubePos) (foreGround, backGround uint32) {
	idx := d.IndexOf(pos)
	return d.ForeGround[idx], d.BackGround[idx]
}

func RangeSplits(start int, len int, algin int) (ranges [][2]int) {
	ranges = make([][2]int, 0, (len+algin-1)/algin)
	currentStart := start
	for currentStart < start+len {
		currentEnd := (currentStart / algin) * algin
		if currentEnd <= currentStart {
			currentEnd += algin
		}
		if currentEnd > start+len {
			currentEnd = start + len
		}
		ranges = append(ranges, [2]int{currentStart, currentEnd - currentStart})
		currentStart = currentEnd
	}
	return ranges
}

func (structure *DecodedStructure) DumpToChunkProvider(chunkProvider chunks.ChunkProvider) (err error) {
	background, foreground := structure.BackGround, structure.ForeGround
	rtid := uint32(0)
	chunkRangesX := RangeSplits(structure.Origin.X(), structure.Size[0], 16)
	chunkRangesZ := RangeSplits(structure.Origin.Z(), structure.Size[2], 16)
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

type StructureResponse interface {
	Raw() *packet.StructureTemplateDataResponse
	Decode() (*DecodedStructure, error)
}

// a set of low level apis
// can only request and retreive data
// no frequency control
// no size control
// no translation or any other assemble control
type LowLevelAreaRequester interface {
	// TODO: Auto Requester to chunkProvider
	// Aim to support all dimensions (nbt translater should be in chunks.ChunkProvider)
	// RequestArea(pos define.CubePos, size define.CubePos, target chunks.ChunkProvider) *async_wrapper.NoRetAsyncWrapper

	LowLevelRequestStructure(pos define.CubePos, size define.CubePos, structureName string) *async_wrapper.AsyncWrapper[StructureResponse]
	LowLevelRequestStructureWithAutoName(pos define.CubePos, size define.CubePos) *async_wrapper.AsyncWrapper[StructureResponse]

	// LowLevelRequestChunk will not check bot positon or translate nbt
	// so extra code (e.g. req generator & scheduler & postprocess & fallback control) is required
	LowLevelRequestChunk(chunkPos define.ChunkPos) *async_wrapper.AsyncWrapper[*chunks.ChunkWithAuxInfo]
	// TODO: fine grained control
	// RequestSubChunks(chunkPos define.ChunkPos, subChunks []int8) ChunkRequestResultHandler
}
