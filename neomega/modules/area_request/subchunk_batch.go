package area_request

import (
	"time"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

type SubChunkBatchReqHandler struct {
	baseChunkPos define.ChunkPos
	xGen         func() []int8
	zGen         func() []int8
	yGen         func(dim int32) []int8
	getDim       func() int32
	ar           *AreaRequester
	finalDim     *int32
}

func autoDim(info neomega.ExtendInfo) int32 {
	dim, _ := info.GetBotDimension()
	return dim
}

func fullY(dim int32) []int8 {
	s, e := int8(-4), int8(20)
	if dim == 0 {

	} else if dim == 1 {
		s, e = 0, 8
	} else if dim == 2 {
		s, e = 0, 16
	} else {
		// seems same as overworld, not suggest
	}
	ys := make([]int8, 0, 24)
	for i := int8(s); i < e; i++ {
		ys = append(ys, i)
	}
	return ys
}

func fixR(off int8) func() []int8 {
	return func() []int8 { return []int8{off} }
}

func rangeR(startOff, endOffNotIncluded int8) func() []int8 {
	return func() []int8 {
		r := make([]int8, 0, 24)
		for i := int8(startOff); i < endOffNotIncluded; i++ {
			r = append(r, i)
		}
		return r
	}
}

func (h *SubChunkBatchReqHandler) OmitResult() {
	subChunkOffsets := make([]protocol.SubChunkOffset, 0, 24)
	if h.finalDim == nil {
		d := h.getDim()
		h.finalDim = &d
	}

	for _, x := range h.xGen() {
		for _, z := range h.zGen() {
			for _, y := range h.yGen(*h.finalDim) {
				subChunkOffsets = append(subChunkOffsets, [3]int8{x, y, z})
			}
		}
	}
	h.ar.ctrl.SendPacket(&packet.SubChunkRequest{
		Dimension: *h.finalDim,
		Position:  protocol.SubChunkPos{h.baseChunkPos.X(), 0, h.baseChunkPos.Z()},
		Offsets:   subChunkOffsets,
	})
}

type SubChunkBatchResult struct {
	results  map[protocol.SubChunkPos]neomega.SubChunkResult
	finalDim int32
}

func (r *SubChunkBatchResult) MapResults() map[protocol.SubChunkPos]neomega.SubChunkResult {
	return r.results
}

func (r *SubChunkBatchResult) Results() []neomega.SubChunkResult {
	rs := make([]neomega.SubChunkResult, 0, len(r.results))
	for _, e := range r.results {
		rs = append(rs, e)
	}
	return rs
}

func (r *SubChunkBatchResult) AllOk() bool {
	for _, e := range r.results {
		if e.Error() != nil {
			return false
		}
	}
	return true
}

func (r *SubChunkBatchResult) AllErrors() map[protocol.SubChunkPos]error {
	es := make(map[protocol.SubChunkPos]error, len(r.results))
	for k, e := range r.results {
		es[k] = e.Error()
	}
	return es
}

func (r *SubChunkBatchResult) ToChunks(optionalAlterFn func(r neomega.SubChunkResult) (*chunk.SubChunk, map[define.CubePos]map[string]interface{})) map[define.ChunkPos]*chunks.ChunkWithAuxInfo {
	chunkSet := map[define.ChunkPos]*chunks.ChunkWithAuxInfo{}
	for pos, sc := range r.results {
		cp := define.ChunkPos{pos.X(), pos.Z()}
		ySubChunk := pos.Y()
		if r.finalDim == 0 {
			ySubChunk += 4
		}
		var c *chunks.ChunkWithAuxInfo
		var found bool
		if c, found = chunkSet[cp]; !found {
			c = &chunks.ChunkWithAuxInfo{
				Chunk:     chunk.New(blocks.AIR_RUNTIMEID, define.WorldRange),
				BlockNbts: make(map[define.CubePos]map[string]interface{}),
				SyncTime:  time.Now().Unix(),
				ChunkPos:  cp,
			}
			chunkSet[cp] = c
		}
		finalSC, nbts := sc.SubChunk(), sc.NBTsInAbsolutePos()
		if optionalAlterFn != nil {
			finalSC, nbts = optionalAlterFn(sc)
		}
		c.Chunk.AssignSub(int(ySubChunk), finalSC)
		for p, n := range nbts {
			c.BlockNbts[p] = n
		}
	}
	return chunkSet
}

func newSubChunkBatchResult(dim int32, slots []protocol.SubChunkPos) *SubChunkBatchResult {
	placeHolderResult := map[protocol.SubChunkPos]neomega.SubChunkResult{}
	for _, slot := range slots {
		placeHolderResult[slot] = &SubChunkResult{
			resultCode: protocol.SubChunkResultChunkNotFound,
			pos:        slot,
			nbtsInMap:  make(map[define.CubePos]map[string]interface{}),
			subChunk:   chunk.NewSubChunk(blocks.AIR_RUNTIMEID),
			decodeErr:  nil,
		}
	}
	return &SubChunkBatchResult{
		results:  placeHolderResult,
		finalDim: dim,
	}
}

func (h *SubChunkBatchReqHandler) GetResult() *async_wrapper.AsyncWrapper[neomega.SubChunkBatchResult] {
	return async_wrapper.NewAsyncWrapper(func(ac *async_wrapper.AsyncController[neomega.SubChunkBatchResult]) {
		if h.finalDim == nil {
			d := h.getDim()
			h.finalDim = &d
		}
		slots := make([]protocol.SubChunkPos, 0)
		hit := map[protocol.SubChunkPos]bool{}
		for _, x := range h.xGen() {
			for _, z := range h.zGen() {
				for _, y := range h.yGen(*h.finalDim) {
					sp := protocol.SubChunkPos{h.baseChunkPos.X() + int32(x), int32(y), h.baseChunkPos.Z() + int32(z)}
					slots = append(slots, sp)
					hit[sp] = false
				}
			}
		}
		result := newSubChunkBatchResult(*h.finalDim, slots)
		var detachFn func()
		ac.SetUnfinishedResult(result)
		ac.SetCancelHook(func() {
			detachFn()
		})
		detachFn = h.ar.AttachSubChunkResultListener(func(scr neomega.SubChunkResult) {
			if _, found := hit[scr.SubCunkPos()]; found {
				hit[scr.SubCunkPos()] = true
				result.results[scr.SubCunkPos()] = scr
				delete(hit, scr.SubCunkPos())
			}
			if len(hit) == 0 {
				detachFn()
				ac.SetResult(result)
			}
		})
		h.OmitResult()
	}, true)
}

func (h *SubChunkBatchReqHandler) AutoDimension() neomega.SubChunkBatchReqHandler {
	h.getDim = func() int32 { return autoDim(h.ar.extendInfo) }
	return h
}

func (h *SubChunkBatchReqHandler) InDimension(dim int32) neomega.SubChunkBatchReqHandler {
	h.getDim = func() int32 { return dim }
	return h
}

func (h *SubChunkBatchReqHandler) X(xOffset int8) neomega.SubChunkBatchReqHandler {
	h.xGen = fixR(xOffset)
	return h
}

func (h *SubChunkBatchReqHandler) Z(zOffset int8) neomega.SubChunkBatchReqHandler {
	h.zGen = fixR(zOffset)
	return h
}

func (h *SubChunkBatchReqHandler) Y(yOffset int8) neomega.SubChunkBatchReqHandler {
	h.yGen = func(dim int32) []int8 { return []int8{yOffset} }
	return h
}

func (h *SubChunkBatchReqHandler) XRange(startOffset int8, endNotIncludedOffset int8) neomega.SubChunkBatchReqHandler {
	h.xGen = rangeR(startOffset, endNotIncludedOffset)
	return h
}

func (h *SubChunkBatchReqHandler) ZRange(startOffset int8, endNotIncludedOffset int8) neomega.SubChunkBatchReqHandler {
	h.zGen = rangeR(startOffset, endNotIncludedOffset)
	return h
}
func (h *SubChunkBatchReqHandler) YRange(startOffset int8, endNotIncludedOffset int8) neomega.SubChunkBatchReqHandler {
	h.yGen = func(dim int32) []int8 { return rangeR(startOffset, endNotIncludedOffset)() }
	return h
}

func (h *SubChunkBatchReqHandler) FullY() neomega.SubChunkBatchReqHandler {
	h.yGen = fullY
	return h
}

func (ar *AreaRequester) LowLevelRequestChunk(chunkPos define.ChunkPos) neomega.SubChunkBatchReqHandler {
	return &SubChunkBatchReqHandler{
		baseChunkPos: chunkPos,
		xGen:         fixR(0),
		zGen:         fixR(0),
		yGen:         fullY,
		getDim:       func() int32 { return autoDim(ar.extendInfo) },
		ar:           ar,
	}
}
