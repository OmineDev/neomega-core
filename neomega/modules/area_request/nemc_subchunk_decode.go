package area_request

import (
	"bytes"
	"errors"
	"fmt"

	standard_nbt "github.com/OmineDev/neomega-core/minecraft/nbt"
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/pterm/pterm"
)

var ErrCannotGetPosFromNBT = errors.New("cannot get pos from nbt")

// SubChunkDecode no longer need rtid translate since blocks has already align nemc and standard rtid
func SubChunkDecode(data []byte) (subChunkIndex int8, subChunk *chunk.SubChunk, nbts map[define.CubePos]map[string]interface{}, err error) {
	buf := bytes.NewBuffer(data)
	subChunkIndex, subChunk, err = SubChunkBlockDecode(buf)
	nbts = make(map[define.CubePos]map[string]interface{}, 0)
	if err != nil {
		return
	}
	if buf.Len() > 0 {
		nbtDecoder := standard_nbt.NewDecoder(buf)
		blockData := make(map[string]interface{})
		for buf.Len() != 0 {
			if err := nbtDecoder.Decode(&blockData); err != nil {
				pterm.Printfln("decode chunk nbt error %v", err)
				break
			}
			//fmt.Println(blockData)
			p, ok := define.GetCubePosFromNBT(blockData)
			if ok {
				nbts[p] = blockData
			} else {
				err = ErrCannotGetPosFromNBT
			}
		}
	}
	return subChunkIndex, subChunk, nbts, err
}

func SubChunkBlockDecode(buf *bytes.Buffer) (int8, *chunk.SubChunk, error) {
	ver, err := buf.ReadByte()
	Index := int8(127)
	if err != nil {
		return Index, nil, fmt.Errorf("error reading version: %w", err)
	}
	sub := chunk.NewSubChunk(blocks.AIR_RUNTIMEID)
	switch ver {
	default:
		return Index, nil, fmt.Errorf("unknown sub chunk version %v: can't decode", ver)
	case 9:
		// Version 8 allows up to 256 layers for one sub chunk.
		storageCount, err := buf.ReadByte()
		if err != nil {
			return Index, nil, fmt.Errorf("error reading storage count: %w", err)
		}
		uIndex, err := buf.ReadByte()
		Index = int8(uIndex)
		if err != nil {
			return Index, nil, fmt.Errorf("error reading subchunk index: %w", err)
		}
		// The index as written here isn't the actual index of the subchunk within the chunk. Rather, it is the Y
		// value of the subchunk. This means that we need to translate it to an index.
		sub.Storages = make([]*chunk.PalettedStorage, storageCount)

		for i := byte(0); i < storageCount; i++ {
			sub.Storages[i], err = decodePalettedStorage(buf)
			if err != nil {
				return Index, nil, err
			}
		}
	}
	return Index, sub, nil
}

func decodePalettedStorage(buf *bytes.Buffer) (*chunk.PalettedStorage, error) {
	blockSize, err := buf.ReadByte()
	if err != nil {
		return nil, fmt.Errorf("error reading block size: %w", err)
	}
	blockSize >>= 1
	if blockSize == 0x7f {
		return nil, nil
	}
	uint32Count := chunk.PaletteSize(blockSize).Uint32s()
	uint32s := make([]uint32, uint32Count)
	byteCount := uint32Count * 4

	data := buf.Next(byteCount)
	if len(data) != byteCount {
		return nil, fmt.Errorf("cannot read paletted storage (size=%v): not enough block data present: expected %v bytes, got %v", blockSize, byteCount, len(data))
	}
	for i := 0; i < uint32Count; i++ {
		// Explicitly don't use the binary package to greatly improve performance of reading the uint32s.
		uint32s[i] = uint32(data[i*4]) | uint32(data[i*4+1])<<8 | uint32(data[i*4+2])<<16 | uint32(data[i*4+3])<<24
	}
	p, err := decodePalette(buf, chunk.PaletteSize(blockSize))
	if err != nil {
		return nil, err
	}
	return chunk.NewPalettedStorage(uint32s, p), err
}

func decodePalette(buf *bytes.Buffer, blockSize chunk.PaletteSize) (*chunk.Palette, error) {
	var paletteCount int32 = 1
	if blockSize != 0 {
		if err := protocol.Varint32(buf, &paletteCount); err != nil {
			return nil, fmt.Errorf("error reading palette entry count: %w", err)
		}
		if paletteCount <= 0 {
			return nil, fmt.Errorf("invalid palette entry count %v", paletteCount)
		}
	}

	blocks, temp := make([]uint32, paletteCount), int32(0)
	for i := int32(0); i < paletteCount; i++ {
		if err := protocol.Varint32(buf, &temp); err != nil {
			return nil, fmt.Errorf("error decoding palette entry: %w", err)
		}
		blocks[i] = uint32(temp)
	}
	return &chunk.Palette{Values: blocks, Size: blockSize}, nil
}

type SubChunkResult struct {
	resultCode byte
	pos        protocol.SubChunkPos
	nbtsInMap  map[define.CubePos]map[string]interface{}
	subChunk   *chunk.SubChunk
	decodeErr  error
}

func (sr *SubChunkResult) ChunkPos() define.ChunkPos {
	return define.ChunkPos{sr.pos.X(), sr.pos.Z()}
}

func (sr *SubChunkResult) SubCunkPos() protocol.SubChunkPos {
	return sr.pos
}

func (sr *SubChunkResult) ResultCode() byte {
	return sr.resultCode
}

func (sr *SubChunkResult) Error() error {
	if sr.resultCode == protocol.SubChunkResultSuccessAllAir || sr.resultCode == protocol.SubChunkResultSuccess {
		return sr.decodeErr
	} else {
		return fmt.Errorf("server sub chunk response  err: %v", sr.resultCode)
	}
}

func (sr *SubChunkResult) AttachDecodeError(err error) {
	if err == nil {
		return
	}
	if sr.decodeErr == nil {
		sr.decodeErr = err
	} else {
		sr.decodeErr = fmt.Errorf("%v, %v", sr.decodeErr, err)
	}
}

func (sr *SubChunkResult) SubChunk() *chunk.SubChunk {
	return sr.subChunk
}

func (sr *SubChunkResult) NBTsInAbsolutePos() map[define.CubePos]map[string]interface{} {
	return sr.nbtsInMap
}
