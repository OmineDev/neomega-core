package chunks

import (
	"time"

	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
)

var TimeStampNotFound = time.Unix(0, 0).Unix()

// ChunkWithAuxInfo 包含一个区块的方块数据，Nbt信息，
// 收到/保存/读取该区块时区块所在的位置 ChunkX/ChunkZ (ChunkX=X>>4)
// 以及区块收到/保存的时间 (Unix Second)
type ChunkWithAuxInfo struct {
	Chunk     *chunk.Chunk
	BlockNbts map[define.CubePos]map[string]interface{}
	SyncTime  int64
	ChunkPos  define.ChunkPos
}

func (c *ChunkWithAuxInfo) GetBlock(pos define.CubePos) (rtid uint32) {
	if c.Chunk == nil {
		return blocks.AIR_RUNTIMEID
	}
	return c.Chunk.Block(uint8(pos.X()&0xf), int16(pos.Y()), uint8(pos.Z()&0xf), 0)
}

func (c *ChunkWithAuxInfo) GetBlockWithNbt(pos define.CubePos) (rtid uint32, nbt map[string]interface{}) {
	var n map[string]interface{}
	if c.BlockNbts != nil {
		n = c.BlockNbts[pos]
	}
	return c.GetBlock(pos), n
}

type DimensiondChunkWithAuxInfo struct {
	*ChunkWithAuxInfo
	Dim define.Dimension
}

func (cd *ChunkWithAuxInfo) GetSyncTime() time.Time {
	return time.Unix(cd.SyncTime, 0)
}

func (cd *ChunkWithAuxInfo) SetSyncTime(t time.Time) {
	cd.SyncTime = t.Unix()
}

type RidBlockWithNbt struct {
	Rid uint32
	Nbt map[string]interface{}
}

type ChunkWriter interface {
	Write(data *ChunkWithAuxInfo) error
}

// 没有该数据时应该返回 nil
// GetWithDeadline(pos ChunkPos, deadline time.Time) 若在 deadline 前无法获得数据，那么应该返回 nil
type ChunkReader interface {
	Get(ChunkPos define.ChunkPos) (data *ChunkWithAuxInfo)
}

// 可以读写区块
type ChunkProvider interface {
	ChunkReader
	ChunkWriter
}

type MultiDimChunkWriter interface {
	WriteWithDim(data *DimensiondChunkWithAuxInfo) error
}

// 没有该数据时应该返回 nil
// GetWithDeadline(pos ChunkPos, deadline time.Time) 若在 deadline 前无法获得数据，那么应该返回 nil
type MultiDimChunkReader interface {
	GetWithDim(dim define.Dimension, ChunkPos define.ChunkPos) (data *DimensiondChunkWithAuxInfo)
}

// 可以读写区块
type MultiDimChunkProvider interface {
	MultiDimChunkReader
	MultiDimChunkWriter
}
