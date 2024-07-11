package minimal_end_point_entry

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/OmineDev/neomega-core/i18n"
	standard_nbt "github.com/OmineDev/neomega-core/minecraft/nbt"
	"github.com/OmineDev/neomega-core/minecraft/protocol"

	// alter_nbt "github.com/OmineDev/neomega-core/neomega/alter/nbt"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/bundle"
	"github.com/OmineDev/neomega-core/neomega/chunks"
	"github.com/OmineDev/neomega-core/neomega/chunks/chunk"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/nodes"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/nodes/underlay_conn"
	"github.com/pterm/pterm"
)

const ENTRY_NAME = "omega_minimal_end_point"

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

func (c *ChunkRequester) RequestChunk(chunkPos define.ChunkPos) neomega.ChunkRequestResultHandler {
	return &ChunkRequestResultHandler{
		d:        c,
		ctx:      context.Background(),
		chunkPos: chunkPos,
	}
}

func SubChunkDecode(data []byte) (subChunkIndex int8, subChunk *chunk.SubChunk, nbts []map[string]interface{}, err error) {
	buf := bytes.NewBuffer(data)
	subChunkIndex, subChunk, err = SubChunkBlockDecode(buf)
	nbts = make([]map[string]interface{}, 0)
	if buf.Len() > 0 {
		nbtDecoder := standard_nbt.NewDecoder(buf)
		blockData := make(map[string]interface{})
		for buf.Len() != 0 {
			if err := nbtDecoder.Decode(&blockData); err != nil {
				pterm.Printfln("decode chunk nbt error %v", err)
				break
			}
			//fmt.Println(blockData)
			nbts = append(nbts, blockData)
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

		// normal
		allFound := true
		subIndex, subChunk, nbts, subChunkDecodeErr := SubChunkDecode(entry.RawPayload)
		if index != subIndex {
			err = fmt.Errorf("subchunk index mismatch: %v!=%v", index, subIndex)
			break
		}

		if subChunkDecodeErr == nil && allFound && len(nbts) == 0 {
			cd.Chunk.AssignSub(int(subIndex+4), subChunk)
			for _, nbt := range nbts {
				x, y, z, ok := define.GetPosFromNBT(nbt)
				if ok {
					cd.BlockNbts[define.CubePos{x, y, z}] = nbt
				}
			}
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
	}
	delete(c.chunkListeners, cp)
}

func NewChunkRequester(interact neomega.InteractCore, react neomega.ReactCore, info neomega.ExtendInfo) *ChunkRequester {
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

func Entry(args *Args) {
	var node defines.Node
	// ctx := context.Background()
	{
		client, err := underlay_conn.NewClientFromBasicNet(args.AccessPointAddr, time.Second)
		if err != nil {
			panic(err)
		}
		slave, err := nodes.NewZMQSlaveNode(client)
		if err != nil {
			panic(err)
		}
		node = nodes.NewGroup("neomega", slave, false)
		node.ListenMessage("reboot", func(msg defines.Values) {
			reason, _ := msg.ToString()
			fmt.Println(reason)
			os.Exit(3)
		}, false)
		if !node.CheckNetTag("access-point") {
			panic(i18n.T(i18n.S_no_access_point_in_network))
		}
		for {
			if node.CheckNetTag("access-point-ready") {
				break
			}
			time.Sleep(time.Second)
		}
	}

	omegaCore, err := bundle.NewEndPointMicroOmega(node)
	if err != nil {
		panic(err)
	}
	ret := omegaCore.GetGameControl().SendWebSocketCmdNeedResponse("tp @s 1024 100 1024").BlockGetResult()
	fmt.Println(ret)
	requestor := NewChunkRequester(omegaCore.GetGameControl(), omegaCore.GetReactCore(), omegaCore.GetMicroUQHolder())
	chunk, err := requestor.RequestChunk(define.ChunkPos{1024 >> 4, 1024 >> 4}).BlockGetResult()
	fmt.Println(chunk)
	fmt.Println(err)
	// omegaCore, err := bundle.NewEndPointMicroOmega(node)
	// if err != nil {
	// 	panic(err)
	// }
	// err = omegaCore.GetBotAction().HighLevelPickBlock(define.CubePos{1148, -60, 1029}, 3, 3)
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("pick complete")
	// time.Sleep(time.Second)
	// omegaCore.GetBotAction().DropItemFromHotBar(3)
	// time.Sleep(time.Second)
	// fmt.Println(omegaCore)
	// players := omegaCore.GetMicroUQHolder().GetAllOnlinePlayers()
	// for _, player := range players {
	// 	fmt.Println(player.GetUsername())
	// 	fmt.Println(player.GetEntityUniqueID())
	// }
	// provider := memory_hold_canvas.NewMemoryChunkCache(nil)

	// origBlock := define.CubePos{1149, -60, 1025}
	// targetBlock := define.CubePos{1149, -60, 1027}

	// block, err := omegaCore.GetStructureRequester().RequestStructure(origBlock, define.CubePos{1, 1, 1}, "block").BlockGetResult()
	// if err != nil {
	// 	panic(err)
	// }
	// decoded, err := block.Decode()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println("requested")
	// nbtData := decoded.Nbts[origBlock]
	// fmt.Println(nbtData)
	// rtid, _ := decoded.BlockOf(define.CubePos{0, 0, 0})
	// blockName, _ := blocks.RuntimeIDToBlockNameWithStateStr(rtid)
	// omegaCore.GetBotAction().SetBlockCmd(targetBlock, blockName).AsWebSocket().SendAndGetResponse().BlockGetResult()

	// containerInfo, _ := neomega.GenContainerItemsInfoFromItemsNbt(nbtData["Items"].([]any))
	// fmt.Println(containerInfo)
	// blockName, found := blocks.RuntimeIDToBlockNameWithStateStr(rtid)
	// if !found {
	// 	panic("block not found")
	// }
	// fmt.Println(blockName)
	// err = omegaCore.GetBotAction().HighLevelGenContainer(targetBlock, containerInfo, blockName)
	// fmt.Println(err)

	// itemNBT := nbtData["Item"]
	// rotationInfo, _ := nbtData["ItemRotation"].(float32)
	// rotation := (int(rotationInfo) / 45) + 1
	// item, err := neomega.GenItemInfoFromItemFrameNBT(itemNBT)
	// if err != nil {
	// 	panic(err)
	// }
	// err = omegaCore.GetBotAction().HighLevelMakeItem(item, 0, targetBlock.Add(define.CubePos{1, 0, -1}), targetBlock.Add(define.CubePos{1, 0, 1}))
	// if err != nil {
	// 	panic(err)
	// }
	// for rotation > 0 {
	// 	err = omegaCore.GetBotAction().HighLevelPlaceItemFrameItem(targetBlock, 0)
	// 	if err != nil {
	// 		break
	// 	}
	// 	rotation--
	// }

	// fg, _ := decoded.BlockOf(define.CubePos{0, 0, 0})

	// fmt.Println(blockNameWithState)
	// // 0~3 z+
	// // 4~7 x-
	// // 8~11 z-
	// // 12~16 x+
	// blk, _ := blocks.RuntimeIDToBlock(fg)
	// rot := int32(0)
	// if len(blk.States()) == 1 {
	// 	if blk.States()[0].Name == "facing_direction" {
	// 		if blk.States()[0].Value.Int32Val() == 2 {
	// 			rot = 8
	// 		}
	// 		if blk.States()[0].Value.Int32Val() == 3 {
	// 			rot = 0
	// 		}
	// 		if blk.States()[0].Value.Int32Val() == 4 {
	// 			rot = 4
	// 		}
	// 		if blk.States()[0].Value.Int32Val() == 5 {
	// 			rot = 12
	// 		}
	// 	}
	// 	if blk.States()[0].Name == "ground_sign_direction" {
	// 		rot = blk.States()[0].Value.Int32Val()
	// 	}
	// } else if len(blk.States()) == 4 {
	// 	if blk.States()[1].Value.Int32Val() == 2 {
	// 		rot = 8
	// 	}
	// 	if blk.States()[1].Value.Int32Val() == 3 {
	// 		rot = 0
	// 	}
	// 	if blk.States()[1].Value.Int32Val() == 4 {
	// 		rot = 4
	// 	}
	// 	if blk.States()[1].Value.Int32Val() == 5 {
	// 		rot = 12
	// 	}

	// }

	// font := define.CubePos{-2, 0, 2}
	// if rot >= 4 {
	// 	font = define.CubePos{-2, 0, -2}
	// }
	// if rot >= 8 {
	// 	font = define.CubePos{2, 0, -2}
	// }
	// if rot >= 12 {
	// 	font = define.CubePos{2, 0, 2}
	// }
	// back := define.CubePos{0, 0, 0}.Sub(font)
	// targetPos := define.CubePos{1051, -60, 982}
	// omegaCore.GetBotAction().SetBlockCmd(targetPos, blockNameWithState).Send()
	// omegaCore.GetBotAction().SelectHotBar(0)
	// omegaCore.GetBotAction().ReplaceHotBarItemCmd(0, "air").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// omegaCore.GetBotAction().UseHotBarItemOnBlock(targetPos, fg, 4, 0)
	// omegaCore.GetGameControl().SendPacket(&packet.BlockActorData{
	// 	Position: protocol.BlockPos{int32(targetPos.X()), int32(targetPos.Y()), int32(targetPos.Z())},
	// 	NBTData:  opt.ToNBT(),
	// })
	// if dyeName := opt.FrontText.GetDyeName(); dyeName != "" {
	// 	fmt.Println(font)
	// 	omegaCore.GetBotAction().ReplaceHotBarItemCmd(0, dyeName).SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// 	omegaCore.GetBotAction().UseHotBarItemOnBlockWithBotOffset(targetPos, font, fg, 0, 0)
	// }
	// if opt.FrontText.IgnoreLighting == 1 {
	// 	omegaCore.GetBotAction().ReplaceHotBarItemCmd(0, "glow_ink_sac").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// 	omegaCore.GetBotAction().UseHotBarItemOnBlockWithBotOffset(targetPos, font, fg, 0, 0)
	// }
	// if dyeName := opt.BackText.GetDyeName(); dyeName != "" {
	// 	omegaCore.GetBotAction().ReplaceHotBarItemCmd(0, dyeName).SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// 	omegaCore.GetBotAction().UseHotBarItemOnBlockWithBotOffset(targetPos, back, fg, 0, 0)
	// }

	// if opt.BackText.IgnoreLighting == 1 {
	// 	omegaCore.GetBotAction().ReplaceHotBarItemCmd(0, "glow_ink_sac").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// 	omegaCore.GetBotAction().UseHotBarItemOnBlockWithBotOffset(targetPos, back, fg, 0, 0)
	// }
	// if opt.IsWaxed == 1 {
	// 	omegaCore.GetBotAction().ReplaceHotBarItemCmd(0, "honeycomb").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// 	omegaCore.GetBotAction().UseHotBarItemOnBlock(targetPos, fg, 0, 0)
	// }

	// omegaCore.GetGameControl().SendPacket(&packet.RequestPermissions{
	// 	EntityUniqueID:       -4294967295,
	// 	PermissionLevel:      3,
	// 	RequestedPermissions: protocol.AbilityBuild,
	// })
	// for {
	// 	player, _ := omegaCore.GetPlayerInteract().GetPlayerKit("2401PT")
	// 	fmt.Println(player.IsOP())
	// 	time.Sleep(time.Second * 3)
	// }

	// player.SetBuildAbility(false)
	// player.SetDoorsAndSwitchesAbility(false)
	// go func() {
	// 	i := 0
	// 	for {
	// 		i++
	// 		time.Sleep(time.Second)
	// 		ret := omegaCore.GetGameControl().SendWebSocketCmdNeedResponse(fmt.Sprintf("tp @s ~~ %v", i)).BlockGetResult()
	// 		fmt.Println(ret)
	// 	}
	// }()

	// go func() {
	// 	i := 0
	// 	for {
	// 		i++
	// 		time.Sleep(time.Second)
	// 		pos, _ := omegaCore.GetMicroUQHolder().GetExtendInfo().GetBotPosition()
	// 		dimension, _ := omegaCore.GetMicroUQHolder().GetBotDimension()
	// 		fmt.Printf("%v %v %v %v %v\n", i, pos.X(), pos.Y(), pos.Z(), dimension)
	// 	}
	// }()

	// omegaCore.GetGameListener().SetTypedPacketCallBack(packet.IDItemStackResponse, func(p packet.Packet) {
	// 	pk := p.(*packet.ItemStackResponse)
	// 	for _, r := range pk.Responses {
	// 		fmt.Printf("status: %v id: %v\n", r.Status, r.RequestID)
	// 		for _, slot := range r.ContainerInfo {
	// 			bs, _ := json.Marshal(slot)
	// 			fmt.Printf("\t%v\n", string(bs))
	// 		}
	// 	}
	// }, true)

	// go func() {
	// 	for {
	// 		picked := <-pickedItemChan
	// 		bs, _ := json.Marshal(picked.NewItem)
	// 		fmt.Printf("type: %v window: %v slot: %v value: %v\n", picked.SourceType, picked.WindowID, picked.InventorySlot, string(bs))
	// 	}
	// }()

	// omegaCore.GetBotAction().ReplaceContainerItemFullCmd(define.CubePos{260, -60, 35}, 1, "planks", 5, 2, &neomega.ItemComponentsInGiveOrReplace{
	// 	CanPlaceOn:  []string{"grass"},
	// 	CanDestroy:  []string{"sand"},
	// 	ItemLock:    neomega.ItemLockPlaceSlot,
	// 	KeepOnDeath: true,
	// }).SendAndGetResponse().BlockGetResult()

	// omegaCore.GetBotAction().ReplaceContainerItemFullCmd(define.CubePos{260, -60, 35}, 4, "iron_sword", 5, 4, &neomega.ItemComponentsInGiveOrReplace{
	// 	CanPlaceOn:  []string{"grass"},
	// 	CanDestroy:  []string{"sand"},
	// 	ItemLock:    neomega.ItemLockPlaceSlot,
	// 	KeepOnDeath: true,
	// }).SendAndGetResponse().BlockGetResult()
	// time.Sleep(time.Second)

	// origBlock := define.CubePos{265, -59, 47}
	// block, err := omegaCore.GetStructureRequester().RequestStructure(origBlock, define.CubePos{1, 1, 1}, "block").BlockGetResult()
	// if err != nil {
	// 	panic(err)
	// }
	// decoded, err := block.Decode()
	// blockName, _ := blocks.RuntimeIDToBlockNameWithStateStr(decoded.ForeGround[0])
	// fmt.Println(blockName)
	// nbtData := decoded.Nbts[origBlock]["Item"]
	// fmt.Println(nbtData)
	// item, err := neomega.GenItemInfoFromItemFrameNBT(nbtData)
	// fmt.Println(item)
	// err = omegaCore.GetBotAction().HighLevelMakeItem(item, 0, origBlock.Add(define.CubePos{1, 0, -1}), origBlock.Add(define.CubePos{1, 0, 1}))
	// fmt.Println("make err", err)
	// err = omegaCore.GetBotAction().HighLevelPlaceItemFrameItem(define.CubePos{259, -60, 54}, 0)
	// fmt.Println("put err", err)
	// read out nbt data of a specific container
	// then re-make it
	// origChest := define.CubePos{264, -60, 46}
	// block, err := omegaCore.GetStructureRequester().RequestStructure(origChest, define.CubePos{1, 1, 1}, "block").BlockGetResult()
	// if err != nil {
	// 	panic(err)
	// }
	// decoded, err := block.Decode()
	// if err != nil {
	// 	panic(err)
	// }
	// fmt.Println(decoded)
	// nbtData := decoded.Nbts[origChest]["Items"]
	// containerInfo, _ := neomega.GenContainerItemsInfoFromItemsNbt(nbtData.([]any))
	// fmt.Println(containerInfo)
	// blockName, found := blocks.RuntimeIDToBlockNameWithStateStr(decoded.ForeGround[0])
	// if !found {nodes
	// 	panic(err)
	// }
	// fmt.Println(blockName)
	// err = omegaCore.GetBotAction().HighLevelGenContainer(define.CubePos{260, -60, 46}, containerInfo, blockName)
	// fmt.Println(err)

	// containerBlock.
	// 	omegaCore.B.HighLevelGenContainer()
	// HighLevelSetContainerItems(omegaCore, containerInfo, define.CubePos{257, -60, 36})
	// fmt.Println(containerInfo)

	// os.WriteFile(fmt.Sprintf("%v.nbt", decoded.ForeGround[0]), bs, 0755)
	// // set container item
	// result := omegaCore.GetBotAction().ReplaceContainerItemFullCmd(define.CubePos{264, -60, 46}, 0, "stone", 10, 0, &neomega.ItemComponentsInGiveOrReplace{
	// 	CanPlaceOn: []string{"grass", "stone"},
	// 	CanDestroy: []string{"sand"},
	// }).SendAndGetResponse().BlockGetResult()
	// fmt.Println(result)

	// // make item, enchant, and drop
	// result := omegaCore.GetBotAction().ReplaceBotHotBarItemFullCmd(2, "diamond_sword", 10, 0, &neomega.ItemComponentsInGiveOrReplace{
	// 	CanPlaceOn:  []string{"grass", "stone"},
	// 	CanDestroy:  []string{"sand"},
	// 	ItemLock:    neomega.ItemLockPlaceSlot,
	// 	KeepOnDeath: true,
	// }).SendAndGetResponse().BlockGetResult()
	// fmt.Println(result)
	// err = omegaCore.GetBotAction().HighLevelEnchantItem(2, map[string]int32{
	// 	"9":          2,
	// 	"30":         2,
	// 	"unbreaking": 2,
	// })
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// err = omegaCore.GetBotAction().DropItemFromHotBar(2)
	// if err != nil {
	// 	panic(err)
	// }

	// // put item in container
	// omegaCore.GetBotAction().ReplaceHotBarItemCmd(4, "oak_sign").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// omegaCore.GetBotAction().ReplaceHotBarItemCmd(5, "dark_oak_sign").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()

	// err = omegaCore.GetBotAction().HighLevelMoveItemToContainer(define.CubePos{264, -60, 46}, map[uint8]uint8{4: 7, 5: 10})
	// if err != nil {
	// 	fmt.Println(err.Error())
	// } else {
	// 	fmt.Println("move ok")
	// }

	// // rename item use anvil
	// omegaCore.GetBotAction().ReplaceHotBarItemCmd(0, "oak_sign").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	// err = omegaCore.GetBotAction().HighLevelRenameItemWithAutoGenAnvil(define.CubePos{260, -60, 41}, 0, "240")
	// if err != nil {
	// 	fmt.Println(err.Error())
	// }

	// // drop item from hotbar
	// err = omegaCore.GetBotAction().DropItemFromHotBar(0)
	// if err != nil {
	// 	fmt.Println(err.Error())
	// } else {
	// 	fmt.Println("drop ok")
	// }

	// //listen item picked
	// pickedItemChan, _, err := omegaCore.GetBotAction().HighLevelListenItemPicked(time.Minute * 10)
	// if err != nil {
	// 	panic(err)
	// }

	// // break and pick block
	// _, err = omegaCore.GetBotAction().HighLevelBlockBreakAndPickInHotBar(define.CubePos{266, -60, 35}, true, map[uint8]bool{5: false, 6: false, 7: false}, 2)
	// if err != nil {
	// 	fmt.Println(err)
	// } else {
	// 	fmt.Println("break and pick ok")
	// }

	// err = omegaCore.GetBotAction().DropItemFromHotBar(5)
	// if err != nil {
	// 	panic(err)
	// }
	// err = omegaCore.GetBotAction().DropItemFromHotBar(6)
	// if err != nil {
	// 	panic(err)
	// }
	// err = omegaCore.GetBotAction().DropItemFromHotBar(7)
	// if err != nil {
	// 	panic(err)
	// }

	// place sign
	// omegaCore.GetBotAction().HighLevelPlaceSign(
	// 	define.CubePos{262, -60, 49}, "2401!", true, "standing_sign",
	// )

	// // put command block
	// omegaCore.GetBotAction().HighLevelPlaceCommandBlock(&neomega.PlaceCommandBlockOption{
	// 	X: 263, Y: -60, Z: 49,
	// 	BlockName:    "command_block",
	// 	BlockState:   "0",
	// 	NeedRedStone: true,
	// 	Name:         "240!",
	// 	Command:      "say 240!",
	// }, 3)

	// go func() {
	// 	reader := bufio.NewReader(os.Stdin)
	// 	for {
	// 		time.Sleep(time.Second / 3)
	// 		fmt.Printf("> ")
	// 		line, err := reader.ReadString('\n')
	// 		if err != nil {
	// 			panic(err)
	// 		}
	// 		line = strings.TrimSpace(line)
	// 		if strings.HasPrefix(line, "/") {
	// 			omegaCore.GetGameControl().SendWebSocketCmdNeedResponse(line).AsyncGetResult(func(output *packet.CommandOutput) {
	// 				fmt.Println(output)
	// 			})
	// 			continue
	// 		}
	// 		if strings.HasPrefix(line, "player/") {
	// 			omegaCore.GetGameControl().SendPlayerCmdNeedResponse(strings.TrimPrefix(line, "player/")).AsyncGetResult(func(output *packet.CommandOutput) {
	// 				fmt.Println(output)
	// 			})
	// 			continue
	// 		}
	// 		if strings.HasPrefix(line, "#uq.") {
	// 			line = strings.TrimPrefix(line, "#uq.")
	// 			if line == "all_players" {
	// 				for i, player := range omegaCore.GetMicroUQHolder().GetAllOnlinePlayers() {
	// 					name, _ := player.GetUsername()
	// 					uuid, _ := player.GetUUIDString()
	// 					fmt.Printf("%v %v %v\n", i, name, uuid)
	// 				}
	// 			}
	// 			// if line == "command_permission_level" {
	// 			// 	for _, player := range omegaCore.GetMicroUQHolder().GetAllOnlinePlayers() {
	// 			// 		name, _ := player.GetUsername()
	// 			// 		level, found := player.GetCommandPermissionLevel()
	// 			// 		fmt.Printf("%v %v %v\n", name, level, found)
	// 			// 	}
	// 			// }
	// 			// if line == "op_permission_level" {
	// 			// 	for _, player := range omegaCore.GetMicroUQHolder().GetAllOnlinePlayers() {
	// 			// 		name, _ := player.GetUsername()
	// 			// 		level, found := player.GetOPPermissionLevel()
	// 			// 		fmt.Printf("%v %v %v\n", name, level, found)
	// 			// 	}
	// 			// }
	// 			continue
	// 		}
	// 	}
	// }()

	panic(<-node.WaitClosed())
}
