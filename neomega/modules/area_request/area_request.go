package area_request

import (
	"fmt"
	"sync"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
	"github.com/OmineDev/neomega-core/utils/string_wrapper"
	"github.com/OmineDev/neomega-core/utils/sync_wrapper"
	"github.com/google/uuid"
)

type AreaRequester struct {
	ctrl               neomega.GameIntractable
	react              neomega.PacketDispatcher
	entityUniqueID     int64
	extendInfo         neomega.ExtendInfo
	structuresMu       sync.Mutex
	structureListeners map[string][]func(neomega.StructureResponse)
	subChunkListeners  *sync_wrapper.HybridListener[neomega.SubChunkResult]
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
			ac.SetResult(sr)
		})
		o.structuresMu.Unlock()
		o.requestStructure(pos, size, structureName)
	}, false)
}

func (o *AreaRequester) LowLevelRequestStructureWithAutoName(pos define.CubePos, size define.CubePos) *async_wrapper.AsyncWrapper[neomega.StructureResponse] {
	name := string_wrapper.ReplaceWithUnfilteredLetter(uuid.New().String())
	return o.LowLevelRequestStructure(pos, size, name)
}

func (c *AreaRequester) onSubChunkPacket(pk *packet.SubChunk) {
	if c.subChunkListeners.Len() == 0 {
		return
	}
	for _, entry := range pk.SubChunkEntries {
		xOff, yOff, zOff := entry.Offset[0], entry.Offset[1], entry.Offset[2]
		subChunkPos := protocol.SubChunkPos{pk.Position.X() + int32(xOff), pk.Position.Y() + int32(yOff), pk.Position.Z() + int32(zOff)}
		ret := &SubChunkResult{
			resultCode: entry.Result,
			pos:        subChunkPos,
			nbtsInMap:  make(map[define.CubePos]map[string]interface{}),
			subChunk:   chunk.NewSubChunk(blocks.AIR_RUNTIMEID),
		}
		if entry.Result == protocol.SubChunkResultSuccessAllAir {
			c.subChunkListeners.Call(ret)
			continue
		}
		if entry.Result != protocol.SubChunkResultSuccess {
			c.subChunkListeners.Call(ret)
			continue
		}

		subIndex, subChunk, nbts, subChunkDecodeErr := SubChunkDecode(entry.RawPayload)
		if int8(pk.Position.Y())+yOff != subIndex {
			panic(fmt.Errorf("subchunk index mismatch: (%v+%v)!=%v", pk.Position.Y(), yOff, subIndex))
		}
		ret.subChunk = subChunk
		ret.nbtsInMap = nbts
		ret.decodeErr = subChunkDecodeErr
		c.subChunkListeners.Call(ret)
	}
}

func (c *AreaRequester) SetOnSubChunkResult(nonBlockingCallback func(neomega.SubChunkResult)) {
	c.subChunkListeners.SetNonBlockingFixListener(nonBlockingCallback)
}
func (c *AreaRequester) AttachSubChunkResultListener(nonBlockingCallback func(neomega.SubChunkResult)) (detachFn func()) {
	return c.subChunkListeners.AttachNonBlockingDetachableListener(nonBlockingCallback)
}

func NewAreaRequester(ctrl neomega.GameIntractable, react neomega.PacketDispatcher, uq neomega.MicroUQHolder, info neomega.ExtendInfo) *AreaRequester {
	d := &AreaRequester{
		ctrl:               ctrl,
		react:              react,
		entityUniqueID:     uq.GetBotBasicInfo().GetBotUniqueID(),
		extendInfo:         info,
		structuresMu:       sync.Mutex{},
		structureListeners: make(map[string][]func(neomega.StructureResponse)),
		subChunkListeners:  sync_wrapper.NewHybridListener[neomega.SubChunkResult](),
	}
	d.react.SetTypedPacketCallBack(packet.IDStructureTemplateDataResponse, func(p packet.Packet) {
		pk := p.(*packet.StructureTemplateDataResponse)
		d.onStructureResponse(pk)
	}, true)
	react.SetTypedPacketCallBack(packet.IDSubChunk, func(p packet.Packet) {
		pk := p.(*packet.SubChunk)
		d.onSubChunkPacket(pk)
	}, true)
	return d
}
