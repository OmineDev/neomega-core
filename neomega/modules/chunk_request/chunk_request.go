package chunk_request

import (
	"context"
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
)

type ChunkRequestResultHandler struct {
	d        *ChunkRequester
	ctx      context.Context
	chunkPos define.ChunkPos
}

func (cr *ChunkRequestResultHandler) SetContext(ctx context.Context) neomega.ChunkRequestResultHandler {
	cr.ctx = ctx
	return cr
}

func (cr *ChunkRequestResultHandler) SetTimeout(timeout time.Duration) neomega.ChunkRequestResultHandler {
	ctx, _ := context.WithTimeout(cr.ctx, timeout)
	cr.ctx = ctx
	return cr
}

func (cr *ChunkRequestResultHandler) AsyncGetResult(callback func(c *chunks.ChunkWithAuxInfo, err error)) {
	w := make(chan struct {
		cd  *chunks.ChunkWithAuxInfo
		err error
	}, 1)
	cr.d.mu.Lock()
	cr.d.chunkListeners[cr.chunkPos] = append(cr.d.chunkListeners[cr.chunkPos], func(cd *chunks.ChunkWithAuxInfo, err error) {
		w <- struct {
			cd  *chunks.ChunkWithAuxInfo
			err error
		}{
			cd, err,
		}
	})
	cr.d.mu.Unlock()
	cr.d.sendSubChunkRequest(cr.chunkPos)
	go func() {
		select {
		case response := <-w:
			callback(response.cd, response.err)
		case <-cr.ctx.Done():
			callback(nil, cr.ctx.Err())
		}
	}()
}

func (cr *ChunkRequestResultHandler) BlockGetResult() (c *chunks.ChunkWithAuxInfo, err error) {
	w := make(chan struct {
		cd  *chunks.ChunkWithAuxInfo
		err error
	}, 1)
	cr.d.mu.Lock()
	cr.d.chunkListeners[cr.chunkPos] = append(cr.d.chunkListeners[cr.chunkPos], func(cd *chunks.ChunkWithAuxInfo, err error) {
		w <- struct {
			cd  *chunks.ChunkWithAuxInfo
			err error
		}{
			cd, err,
		}
	})
	cr.d.mu.Unlock()
	cr.d.sendSubChunkRequest(cr.chunkPos)
	select {
	case response := <-w:
		return response.cd, response.err
	case <-cr.ctx.Done():
		return nil, cr.ctx.Err()
	}
}

type ChunkRequester struct {
	interact       neomega.InteractCore
	extendInfo     neomega.ExtendInfo
	chunkListeners map[define.ChunkPos][]func(cd *chunks.ChunkWithAuxInfo, err error)
	mu             sync.Mutex
}

func (c *ChunkRequester) sendSubChunkRequest(chunkPos define.ChunkPos) {
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
	c.interact.SendPacket(&packet.SubChunkRequest{
		Dimension: dim,
		Position:  protocol.SubChunkPos{chunkPos.X(), 0, chunkPos.Z()},
		Offsets:   subChunkOffsets,
	})
}

func (c *ChunkRequester) LowLevelRequestChunk(chunkPos define.ChunkPos) neomega.ChunkRequestResultHandler {
	return &ChunkRequestResultHandler{
		d:        c,
		ctx:      context.Background(),
		chunkPos: chunkPos,
	}
}

func (c *ChunkRequester) onSubChunk(pk *packet.SubChunk) {
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
	c.mu.Lock()
	defer c.mu.Unlock()
	if listeners, found := c.chunkListeners[cp]; found {
		for _, l := range listeners {
			if err != nil {
				l(nil, err)
			} else {
				l(cd, nil)
			}
		}
		delete(c.chunkListeners, cp)
	}
}

func NewChunkRequester(interact neomega.InteractCore, react neomega.ReactCore, info neomega.ExtendInfo) neomega.LowLevelChunkRequester {
	r := &ChunkRequester{
		interact:       interact,
		extendInfo:     info,
		mu:             sync.Mutex{},
		chunkListeners: make(map[define.ChunkPos][]func(cd *chunks.ChunkWithAuxInfo, err error)),
	}
	react.SetTypedPacketCallBack(packet.IDSubChunk, func(p packet.Packet) {
		pk := p.(*packet.SubChunk)
		r.onSubChunk(pk)
	}, true)

	return r
}
