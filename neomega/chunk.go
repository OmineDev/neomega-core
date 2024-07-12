package neomega

import (
	"context"
	"time"

	"github.com/OmineDev/neomega-core/neomega/chunks"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
)

type ChunkRequestResultHandler interface {
	BlockGetResult() (r *chunks.ChunkWithAuxInfo, err error)
	AsyncGetResult(callback func(r *chunks.ChunkWithAuxInfo, err error))
	SetContext(ctx context.Context) ChunkRequestResultHandler
	SetTimeout(timeout time.Duration) ChunkRequestResultHandler
}

// an low level api, can only request and retreive data without auto request gen or nbt translate
type LowLevelChunkRequester interface {
	// LowLevelRequestChunk will not check bot positon or translate nbt
	// so extra code (e.g. req generator & scheduler & postprocess & fallback control) is required
	LowLevelRequestChunk(chunkPos define.ChunkPos) ChunkRequestResultHandler
	// TODO: fine grained control
	// RequestSubChunks(chunkPos define.ChunkPos, subChunks []int8) ChunkRequestResultHandler
}