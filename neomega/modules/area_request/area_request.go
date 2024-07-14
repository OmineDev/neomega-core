package area_request

import (
	"fmt"
	"sync"
	"time"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
	"github.com/OmineDev/neomega-core/utils/string_wrapper"

	"github.com/google/uuid"
)

type AreaRequester struct {
	ctrl               neomega.GameIntractable
	react              neomega.PacketDispatcher
	entityUniqueID     int64
	extendInfo         neomega.ExtendInfo
	structuresMu       sync.Mutex
	structureListeners map[string][]func(neomega.StructureResponse)
	chunkMu            sync.Mutex
	chunkListeners     map[define.ChunkPos][]func(cd *chunks.ChunkWithAuxInfo, err error)
}

func (a *AreaRequester) onStructureResponse(pk *packet.StructureTemplateDataResponse) {
	a.structuresMu.Lock()
	listeners, ok := a.structureListeners[pk.StructureName]
	if ok {
		delete(a.structureListeners, pk.StructureName)
		a.structuresMu.Unlock()
		r := newStructureResponse(pk)
		for _, v := range listeners {
			v(r)
		}
	} else {
		a.structuresMu.Unlock()
	}
}

func (o *AreaRequester) requestStructure(pos define.CubePos, size define.CubePos, structureName string) {
	o.ctrl.SendPacket(&packet.StructureTemplateDataRequest{
		StructureName: structureName,
		Position:      protocol.BlockPos{int32(pos.X()), int32(pos.Y()), int32(pos.Z())},
		Settings: protocol.StructureSettings{
			PaletteName:               "default",
			IgnoreEntities:            true,
			IgnoreBlocks:              false,
			Size:                      protocol.BlockPos{int32(size.X()), int32(size.Y()), int32(size.Z())},
			Offset:                    protocol.BlockPos{0, 0, 0},
			LastEditingPlayerUniqueID: o.entityUniqueID,
			Rotation:                  0,
			Mirror:                    0,
			Integrity:                 100,
			Seed:                      0,
			AllowNonTickingChunks:     false,
		},
		RequestType: packet.StructureTemplateRequestExportFromSave,
	})
}

func (o *AreaRequester) LowLevelRequestStructure(pos define.CubePos, size define.CubePos, structureName string) *async_wrapper.AsyncWrapper[neomega.StructureResponse] {
	return async_wrapper.NewAsyncWrapper(func(ac *async_wrapper.AsyncController[neomega.StructureResponse]) {
		o.structuresMu.Lock()
		_, ok := o.structureListeners[structureName]
		if !ok {
			o.structureListeners[structureName] = make([]func(neomega.StructureResponse), 0, 1)
		}
		o.structureListeners[structureName] = append(o.structureListeners[structureName], func(sr neomega.StructureResponse) {
			if ac.Context().Err() == nil {
				ac.SetResult(sr)
			}
		})
		o.structuresMu.Unlock()
		o.requestStructure(pos, size, structureName)
	}, false)
}

func (o *AreaRequester) LowLevelRequestStructureWithAutoName(pos define.CubePos, size define.CubePos) *async_wrapper.AsyncWrapper[neomega.StructureResponse] {
	name := string_wrapper.ReplaceWithUnfilteredLetter(uuid.New().String())
	return o.LowLevelRequestStructure(pos, size, name)
}

func (c *AreaRequester) onSubChunk(pk *packet.SubChunk) {
	cp := define.ChunkPos{pk.Position.X(), pk.Position.Z()}
	cd := &chunks.ChunkWithAuxInfo{
		Chunk:     chunk.New(blocks.AIR_RUNTIMEID, define.WorldRange),
		BlockNbts: make(map[define.CubePos]map[string]interface{}),
		SyncTime:  time.Now().Unix(),
		ChunkPos:  cp,
	}
	var err error
	for _, entry := range pk.SubChunkEntries {
		index := entry.Offset[1]
		if entry.Result == protocol.SubChunkResultSuccessAllAir {
			continue
		}
		if entry.Result != protocol.SubChunkResultSuccess {
			err = fmt.Errorf("subchunk result unsuccessful: %v", entry.Result)
			break
		}

		subIndex, subChunk, nbts, subChunkDecodeErr := SubChunkDecode(entry.RawPayload)
		if index != subIndex {
			err = fmt.Errorf("subchunk index mismatch: %v!=%v", index, subIndex)
			break
		}

		if subChunkDecodeErr == nil {
			cd.Chunk.AssignSub(int(subIndex+4), subChunk)
			for _, nbt := range nbts {
				x, y, z, ok := define.GetPosFromNBT(nbt)
				if ok {
					cd.BlockNbts[define.CubePos{x, y, z}] = nbt
				}
			}
		} else {
			err = subChunkDecodeErr
		}
	}
	c.chunkMu.Lock()
	if listeners, found := c.chunkListeners[cp]; found {
		delete(c.chunkListeners, cp)
		c.chunkMu.Unlock()
		for _, l := range listeners {
			if err != nil {
				l(nil, err)
			} else {
				l(cd, nil)
			}
		}
	} else {
		c.chunkMu.Unlock()
	}
}

func (c *AreaRequester) requestSubChunk(chunkPos define.ChunkPos) {
	subChunkOffsets := make([]protocol.SubChunkOffset, 0, 24)
	for i := int8(-4); i <= 19; i++ {
		subChunkOffsets = append(subChunkOffsets, [3]int8{0, i, 0})
	}
	// TODO: auto decide offsets base on dimension
	dim, found := c.extendInfo.GetBotDimension()
	if !found {
		dim = 0
	}
	dim = 0 // TODO: auto decide offsets base on dimension
	c.ctrl.SendPacket(&packet.SubChunkRequest{
		Dimension: dim,
		Position:  protocol.SubChunkPos{chunkPos.X(), 0, chunkPos.Z()},
		Offsets:   subChunkOffsets,
	})
}

func (o *AreaRequester) LowLevelRequestChunk(chunkPos define.ChunkPos) *async_wrapper.AsyncWrapper[*chunks.ChunkWithAuxInfo] {
	return async_wrapper.NewAsyncWrapper[*chunks.ChunkWithAuxInfo](func(ac *async_wrapper.AsyncController[*chunks.ChunkWithAuxInfo]) {
		o.chunkMu.Lock()
		_, ok := o.chunkListeners[chunkPos]
		if !ok {
			o.chunkListeners[chunkPos] = make([]func(*chunks.ChunkWithAuxInfo, error), 0, 1)
		}
		o.chunkListeners[chunkPos] = append(o.chunkListeners[chunkPos], func(cd *chunks.ChunkWithAuxInfo, err error) {
			if ac.Context().Err() == nil {
				ac.SetResultAndErr(cd, err)
			}
		})
		o.chunkMu.Unlock()
		o.requestSubChunk(chunkPos)
	}, false)
}

func NewAreaRequester(ctrl neomega.GameIntractable, react neomega.PacketDispatcher, uq neomega.MicroUQHolder, info neomega.ExtendInfo) *AreaRequester {
	d := &AreaRequester{
		ctrl:               ctrl,
		react:              react,
		entityUniqueID:     uq.GetBotBasicInfo().GetBotUniqueID(),
		extendInfo:         info,
		structuresMu:       sync.Mutex{},
		structureListeners: make(map[string][]func(neomega.StructureResponse)),
		chunkMu:            sync.Mutex{},
		chunkListeners:     make(map[define.ChunkPos][]func(cd *chunks.ChunkWithAuxInfo, err error)),
	}
	d.react.SetTypedPacketCallBack(packet.IDStructureTemplateDataResponse, func(p packet.Packet) {
		pk := p.(*packet.StructureTemplateDataResponse)
		d.onStructureResponse(pk)
	}, true)
	react.SetTypedPacketCallBack(packet.IDSubChunk, func(p packet.Packet) {
		pk := p.(*packet.SubChunk)
		d.onSubChunk(pk)
	}, true)
	return d
}
