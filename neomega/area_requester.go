package neomega

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega/chunks"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type DecodedStructure interface {
	IndexOfRelativePos(pos define.CubePos) int
	BlockOfRelativePos(pos define.CubePos) (foreGround, backGround uint32)
	NBTsInAbsolutePos() map[define.CubePos]map[string]interface{}
	ForeGroundRtidNested() []uint32
	BackGroundRtidNested() []uint32
	DumpToChunkProviderAbsolutePos(chunkProvider chunks.ChunkProvider) (err error)
}

type StructureResponse interface {
	Raw() *packet.StructureTemplateDataResponse
	Decode() (DecodedStructure, error)
}

type SubChunkResult interface {
	SubCunkPos() protocol.SubChunkPos // X<<4,Y<<4,Z<<4 -> real world pos
	ChunkPos() define.ChunkPos        // define.ChunkPos{SubCunkPos().X(),SubCunkPos().Z()}
	Error() error                     // (ResultCode == protocol.SubChunkResultSuccessAllAir || ResultCode == protocol.SubChunkResultSuccess) && decodeErr()==nil
	ResultCode() byte
	NBTsInAbsolutePos() map[define.CubePos]map[string]interface{}
	SubChunk() *chunk.SubChunk
}

type SubChunkBatchResult interface {
	MapResults() map[protocol.SubChunkPos]SubChunkResult
	Results() []SubChunkResult
	AllOk() bool
	AllErrors() map[protocol.SubChunkPos]error
	ToChunks(
		optionalAlterFn func(r SubChunkResult) (*chunk.SubChunk, map[define.CubePos]map[string]interface{}),
	) map[define.ChunkPos]*chunks.ChunkWithAuxInfo
}

type SubChunkBatchReqHandler interface {
	OmitResult()
	GetResult() async_wrapper.AsyncResult[SubChunkBatchResult]
	AutoDimension() SubChunkBatchReqHandler
	InDimension(dim int32) SubChunkBatchReqHandler
	X(xOffset int8) SubChunkBatchReqHandler
	Z(zOffset int8) SubChunkBatchReqHandler
	Y(zOffset int8) SubChunkBatchReqHandler
	XRange(startOffset int8, endNotIncludedOffset int8) SubChunkBatchReqHandler
	YRange(startOffset int8, endNotIncludedOffset int8) SubChunkBatchReqHandler
	ZRange(startOffset int8, endNotIncludedOffset int8) SubChunkBatchReqHandler
	FullY() SubChunkBatchReqHandler
}

// a set of low level apis
// can only request and retreive data
// no frequency control
// no size control
// no translation or any other assemble control
type LowLevelAreaRequester interface {
	LowLevelRequestStructure(pos define.CubePos, size define.CubePos, structureName string) async_wrapper.AsyncResult[StructureResponse]
	LowLevelRequestStructureWithAutoName(pos define.CubePos, size define.CubePos) async_wrapper.AsyncResult[StructureResponse]

	LowLevelRequestChunk(baseChunkPos define.ChunkPos) SubChunkBatchReqHandler
	SetOnSubChunkResult(nonBlockingCallback func(SubChunkResult))
	AttachSubChunkResultListener(nonBlockingCallback func(SubChunkResult)) (detachFn func())
}
