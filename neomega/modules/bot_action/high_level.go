package bot_action

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/neomega/supported_nbt_data"
	"github.com/OmineDev/neomega-core/neomega/supported_nbt_data/supported_item"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/string_wrapper"
	"github.com/OmineDev/neomega-core/utils/structure/pos_operations"

	"github.com/go-gl/mathgl/mgl32"
)

type BotActionHighLevel struct {
	uq               neomega.MicroUQHolder
	ctrl             neomega.InteractCore
	cmdSender        neomega.CmdSender
	cmdHelper        neomega.CommandHelper
	areaRequester    neomega.LowLevelAreaRequester
	microAction      neomega.BotAction
	pickedItemChan   chan protocol.InventoryAction
	playerHotBarChan chan *packet.PlayerHotBar
	muChan           chan struct{}
	node             defines.Node
	// asyncNBTBlockPlacer neomega.AsyncNBTBlockPlacer
	nameCount int
}

func NewBotActionHighLevel(
	uq neomega.MicroUQHolder,
	ctrl neomega.InteractCore,
	react neomega.ReactCore,
	cmdSender neomega.CmdSender,
	cmdHelper neomega.CommandHelper,
	areaRequester neomega.LowLevelAreaRequester,
	microAction neomega.BotAction,
	node defines.Node,
) neomega.BotActionHighLevel {
	muChan := make(chan struct{}, 1)
	muChan <- struct{}{}
	bah := &BotActionHighLevel{
		uq:            uq,
		ctrl:          ctrl,
		cmdSender:     cmdSender,
		cmdHelper:     cmdHelper,
		areaRequester: areaRequester,
		// asyncNBTBlockPlacer: asyncNBTBlockPlacer,
		microAction: microAction,
		muChan:      muChan,
		node:        node,
	}

	react.SetTypedPacketCallBack(packet.IDPlayerHotBar, func(p packet.Packet) {
		// isDroppedItem := false
		pk := p.(*packet.PlayerHotBar)
		if bah.playerHotBarChan == nil {
			return
		}
		select {
		case bah.playerHotBarChan <- pk:
			break
		default:
		}
	}, false)

	react.SetTypedPacketCallBack(packet.IDInventoryTransaction, func(p packet.Packet) {
		// isDroppedItem := false
		pk := p.(*packet.InventoryTransaction)
		for _, _value := range pk.Actions {
			c := bah.pickedItemChan
			if c == nil {
				continue
			}
			// bs, _ := json.Marshal(_value)
			// fmt.Println(string(bs))
			// always slot 1, StackNetworkID 0, when using "give" command
			// protocol.InventoryActionSourceCreative
			// always slot 1, StackNetworkID 0, when item is picked from the world
			// protocol.InventoryActionSourceWorld

			// seems this confusing part is fixed by nemc, now pick a single item will not send two packets
			if _value.SourceType == protocol.InventoryActionSourceContainer {
				// isDroppedItem = true
				value := _value
				select {
				case c <- value:
					break
				default:
				}
			}
			// else if _value.SourceType == protocol.InventoryActionSourceContainer && isDroppedItem {
			// 	value := _value
			// 	select {
			// 	case c <- value:
			// 	default:
			// 	}
			// }
		}
	}, false)

	return bah
}

func (o *BotActionHighLevel) nextCountName() string {
	o.nameCount += 1
	return string_wrapper.ReplaceWithUnfilteredLetter(fmt.Sprintf("%v", o.nameCount))
}

func (o *BotActionHighLevel) occupyBot(timeout time.Duration) (release func(), err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	select {
	case <-ctx.Done():
		o.microAction.SleepTick(1)
		return nil, fmt.Errorf("cannot acquire bot (high level)")
	case <-o.muChan:
		if !o.node.TryLock("bot-action-high", time.Second*2) {
			o.muChan <- struct{}{} // give back bot control
			return nil, fmt.Errorf("cannot acquire bot (high level, distribute)")
		}
		stopChan := make(chan struct{})
		go func() {
			for {
				select {
				case <-stopChan:
					return
				case <-time.NewTimer(time.Second).C:
					o.node.ResetLockTime("bot-action-high", time.Second*2)
				}
			}
		}()
		return func() {
			close(stopChan)
			o.node.ResetLockTime("bot-action-high", 0)
			o.microAction.SleepTick(1)
			o.muChan <- struct{}{} // give back bot control
		}, nil
	}
}

func (o *BotActionHighLevel) highLevelEnsureBotNearby(pos define.CubePos, threshold float32) error {
	botPos, _ := o.uq.GetBotPosition()
	if botPos.Sub(mgl32.Vec3{float32(pos.X()), float32(pos.Y()), float32(pos.Z())}).Len() > threshold {
		ret, err := o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z())).SetTimeout(time.Second * 3).BlockGetResult()
		if ret == nil || err != nil {
			return fmt.Errorf("cannot move to target pos")
		}
	}
	return nil
}

func (o *BotActionHighLevel) HighLevelEnsureBotNearby(pos define.CubePos, threshold float32) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelEnsureBotNearby(pos, threshold)
}

// if we want the pos to be air when use it, recoverFromAir=true
// e.g. 如果我们希望把某个方块位置临时变为 air, 则 wantAir=true
// 如果希望把某个方块位置临时变为某个非空气方块, 则 wantAir=false
func (o *BotActionHighLevel) HighLevelRemoveSpecificBlockSideEffect(pos define.CubePos, backupName string) (deferFunc func(), err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return func() {}, err
	}
	defer release()
	return o.highLevelRemoveSpecificBlockSideEffect(pos, backupName)
}

func (o *BotActionHighLevel) highLevelRemoveSpecificBlockSideEffect(pos define.CubePos, backupName string) (deferFunc func(), err error) {
	deferFunc, err = o.highLevelGetAndRemoveSpecificBlockSideEffect(pos, backupName)
	return deferFunc, err
}

func (o *BotActionHighLevel) highLevelGetAndRemoveSpecificBlockSideEffect(pos define.CubePos, backupName string) (deferFunc func(), err error) {
	o.highLevelEnsureBotNearby(pos.Add(define.CubePos{0, 2, 0}), 3)
	// enlargedStart := pos.Sub(define.CubePos{1, 1, 1})
	ret, err := o.cmdHelper.BackupStructureWithGivenNameCmd(pos, define.CubePos{1, 1, 1}, backupName).SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	if ret == nil || err != nil {
		return func() {}, fmt.Errorf("cannot backup block for revert")
	}
	deferFunc = func() {
		o.cmdHelper.RevertStructureWithGivenNameCmd(pos, backupName).Send()
		o.microAction.SleepTick(1)
	}
	o.cmdHelper.SetBlockCmd(pos, "structure_void").AsWebSocket().SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	return deferFunc, nil
}

func (o *BotActionHighLevel) HighLevelPlaceSign(targetPos define.CubePos, signBlock string, opt *supported_nbt_data.SignBlockSupportedData) (err error) {
	if opt == nil {
		return nil
	}
	o.highLevelEnsureBotNearby(targetPos.Add(define.CubePos{0, 2, 0}), 3)
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelPlaceSign(targetPos, signBlock, opt)
}

func (o *BotActionHighLevel) highLevelPlaceSign(targetPos define.CubePos, signBlock string, opt *supported_nbt_data.SignBlockSupportedData) (err error) {
	rtid, ok := blocks.BlockStrToRuntimeID(signBlock)
	if !ok {
		return fmt.Errorf("sign block not found")
	}
	blockNameWithState, _ := blocks.RuntimeIDToBlockNameWithStateStr(rtid)
	blk, _ := blocks.RuntimeIDToBlock(rtid)
	rot := int32(0)
	if len(blk.States()) == 1 {
		if blk.States()[0].Name == "facing_direction" {
			if blk.States()[0].Value.Int32Val() == 2 {
				rot = 8
			}
			if blk.States()[0].Value.Int32Val() == 3 {
				rot = 0
			}
			if blk.States()[0].Value.Int32Val() == 4 {
				rot = 4
			}
			if blk.States()[0].Value.Int32Val() == 5 {
				rot = 12
			}
		}
		if blk.States()[0].Name == "ground_sign_direction" {
			rot = blk.States()[0].Value.Int32Val()
		}
	} else if len(blk.States()) == 4 {
		if blk.States()[1].Value.Int32Val() == 2 {
			rot = 8
		}
		if blk.States()[1].Value.Int32Val() == 3 {
			rot = 0
		}
		if blk.States()[1].Value.Int32Val() == 4 {
			rot = 4
		}
		if blk.States()[1].Value.Int32Val() == 5 {
			rot = 12
		}

	}

	font := define.CubePos{-2, 0, 2}
	if rot >= 4 {
		font = define.CubePos{-2, 0, -2}
	}
	if rot >= 8 {
		font = define.CubePos{2, 0, -2}
	}
	if rot >= 12 {
		font = define.CubePos{2, 0, 2}
	}
	back := define.CubePos{0, 0, 0}.Sub(font)
	o.cmdHelper.SetBlockCmd(targetPos, blockNameWithState).AsWebSocket().SendAndGetResponse().BlockGetResult()
	o.microAction.SelectHotBar(0)
	o.microAction.SleepTick(2)
	o.cmdHelper.ReplaceHotBarItemCmd(0, "air").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	o.microAction.UseHotBarItemOnBlock(targetPos, rtid, 4, 0)
	o.ctrl.SendPacket(&packet.BlockActorData{
		Position: protocol.BlockPos{int32(targetPos.X()), int32(targetPos.Y()), int32(targetPos.Z())},
		NBTData:  opt.ToNBT(),
	})
	if dyeName := opt.FrontText.GetDyeName(); dyeName != "" {
		o.cmdHelper.ReplaceHotBarItemCmd(0, dyeName).SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
		o.microAction.UseHotBarItemOnBlockWithBotOffset(targetPos, font, rtid, 0, 0)
	}
	if opt.FrontText.IgnoreLighting == 1 {
		o.cmdHelper.ReplaceHotBarItemCmd(0, "glow_ink_sac").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
		o.microAction.UseHotBarItemOnBlockWithBotOffset(targetPos, font, rtid, 0, 0)
	}
	if dyeName := opt.BackText.GetDyeName(); dyeName != "" {
		o.cmdHelper.ReplaceHotBarItemCmd(0, dyeName).SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
		o.microAction.UseHotBarItemOnBlockWithBotOffset(targetPos, back, rtid, 0, 0)
	}

	if opt.BackText.IgnoreLighting == 1 {
		o.cmdHelper.ReplaceHotBarItemCmd(0, "glow_ink_sac").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
		o.microAction.UseHotBarItemOnBlockWithBotOffset(targetPos, back, rtid, 0, 0)
	}
	if opt.IsWaxed == 1 {
		o.cmdHelper.ReplaceHotBarItemCmd(0, "honeycomb").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
		o.microAction.UseHotBarItemOnBlock(targetPos, rtid, 0, 0)
	}
	return nil
}

func (o *BotActionHighLevel) HighLevelPlaceCommandBlock(targetPos define.CubePos, option *supported_nbt_data.CommandBlockSupportedData, maxRetry int) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelPlaceCommandBlock(targetPos, option, maxRetry)
}

func (o *BotActionHighLevel) highLevelPlaceCommandBlock(targetPos define.CubePos, option *supported_nbt_data.CommandBlockSupportedData, maxRetry int) error {
	if err := o.highLevelEnsureBotNearby(targetPos.Add(define.CubePos{0, 2, 0}), 3); err != nil {
		return err
	}
	updateOption := option.GenCommandBlockUpdateFromOption(targetPos)
	sleepTime := 1
	for maxRetry > 0 {
		maxRetry--
		cmd := fmt.Sprintf("setblock %v %v %v %v %v", targetPos.X(), targetPos.Y(), targetPos.Z(), strings.Replace(option.BlockName, "minecraft:", "", 1), option.BlockState)
		o.cmdSender.SendWebSocketCmdNeedResponse(cmd).SetTimeout(time.Second * 3).BlockGetResult()
		o.ctrl.SendPacket(updateOption)
		time.Sleep(100 * time.Millisecond)
		r, err := o.areaRequester.LowLevelRequestStructure(targetPos, define.CubePos{1, 1, 1}, "_temp").BlockGetResult()
		if err != nil {
		} else {
			d, err := r.Decode()
			if err == nil {
				if len(d.NBTsInAbsolutePos()) > 0 {
					for _, tnbt := range d.NBTsInAbsolutePos() {
						ok := true
						if tnbt["id"].(string) != "CommandBlock" {
							ok = false
						} else if strings.TrimSpace(tnbt["Command"].(string)) != strings.TrimSpace(option.Command) {
							ok = false
							if strings.TrimSpace(tnbt["Command"].(string)) == "***" {
								return fmt.Errorf("netease make command to ***")
							}
						} else if tnbt["CustomName"].(string) != option.Name {
							ok = false
						}
						if ok {
							return nil
						}
						break
					}
				}
			}
		}
		sleepTime++
	}
	return fmt.Errorf("cannot successfully place commandblock in given limit")
}

func (o *BotActionHighLevel) HighLevelMoveItemToContainer(pos define.CubePos, moveOperations map[uint8]uint8) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelMoveItemToContainer(pos, moveOperations)
}

func (o *BotActionHighLevel) highLevelMoveItemToContainer(pos define.CubePos, moveOperations map[uint8]uint8) error {
	if err := o.highLevelEnsureBotNearby(pos.Add(define.CubePos{0, 2, 0}), 3); err != nil {
		return err
	}
	structureResponse, err := o.areaRequester.LowLevelRequestStructure(pos.Sub(define.CubePos{1, 0, 1}), define.CubePos{3, 1, 3}, o.nextCountName()).SetTimeout(time.Second * 3).BlockGetResult()
	if err != nil {
		return err
	}
	structure, err := structureResponse.Decode()
	if err != nil {
		return err
	}
	containerRuntimeID := structure.ForeGroundRtidNested()[4]
	// containerNEMCRuntimeID := chunk.StandardRuntimeIDToNEMCRuntimeID(containerRuntimeID)
	if containerRuntimeID == blocks.AIR_RUNTIMEID {
		return fmt.Errorf("block of %v (nemc) not found", containerRuntimeID)
	}
	block, found := blocks.RuntimeIDToBlock(containerRuntimeID)
	if !found {
		panic(fmt.Errorf("block of %v not found", containerRuntimeID))
	}
	_, found = getContainerIDMappingByBlockBaseName(block.ShortName())
	if !found {
		return fmt.Errorf("block %v is not a supported container", block.ShortName())
	}
	for _, targetSlot := range moveOperations {
		o.cmdHelper.ReplaceContainerBlockItemCmd(pos, int32(targetSlot), "air").Send()
	}
	deferAction := func() {}
	if strings.Contains(block.ShortName(), "shulker_box") {
		blockerPos := pos
		face := byte(255)
		if len(structure.NBTsInAbsolutePos()[pos]) > 0 {
			if facing_origin, ok := structure.NBTsInAbsolutePos()[pos]["facing"]; ok {
				face, ok = facing_origin.(byte)
				if !ok {
					face = 255
				}
			}
		}
		if face != 255 {
			switch face {
			case 0:
				blockerPos[1] = blockerPos[1] - 1
			case 1:
				blockerPos[1] = blockerPos[1] + 1
			case 2:
				blockerPos[2] = blockerPos[2] - 1
			case 3:
				blockerPos[2] = blockerPos[2] + 1
			case 4:
				blockerPos[0] = blockerPos[0] - 1
			case 5:
				blockerPos[0] = blockerPos[0] + 1
			}
			deferAction, err = o.highLevelRemoveSpecificBlockSideEffect(blockerPos, o.nextCountName())
			if err != nil {
				return err
			}
		}
	} else if strings.Contains(block.ShortName(), "chest") {
		deferAction, err = o.highLevelRemoveSpecificBlockSideEffect(pos.Add(define.CubePos{0, 1, 0}), o.nextCountName())
		if err != nil {
			return err
		}
	}
	defer deferAction()
	o.microAction.SleepTick(1)
	return o.microAction.MoveItemFromInventoryToEmptyContainerSlots(pos, containerRuntimeID, block.ShortName(), moveOperations)
}

func (o *BotActionHighLevel) HighLevelRenameItemWithAnvil(pos define.CubePos, slot uint8, newName string, autoGenAnvil bool) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelRenameItemWithAnvil(pos, slot, newName, autoGenAnvil)
}

func (o *BotActionHighLevel) highLevelRenameItemWithAnvil(pos define.CubePos, slot uint8, newName string, autoGenAnvil bool) (err error) {
	if err := o.highLevelEnsureBotNearby(pos.Add(define.CubePos{0, 2, 0}), 3); err != nil {
		return err
	}
	deferActionStand := func() {}
	deferAction := func() {}
	if autoGenAnvil {
		deferActionStand, err = o.highLevelRemoveSpecificBlockSideEffect(pos.Add(define.CubePos{0, -1, 0}), o.nextCountName())
		if err != nil {
			return err
		}
		if ret, err := o.cmdHelper.SetBlockCmd(pos.Add(define.CubePos{0, -1, 0}), "glass").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult(); ret == nil || err != nil {
			return fmt.Errorf("cannot place anvil for operation")
		}
		deferAction, err = o.highLevelRemoveSpecificBlockSideEffect(pos, o.nextCountName())
		if err != nil {
			return err
		}
		if ret, err := o.cmdHelper.SetBlockCmd(pos, "anvil").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult(); ret == nil || err != nil {
			return fmt.Errorf("cannot place anvil for operation")
		}
		o.microAction.SleepTick(6)
	}
	// wait until anvil place then get runtime id
	structureResponse, err := o.areaRequester.LowLevelRequestStructure(pos, define.CubePos{1, 1, 1}, "_temp").BlockGetResult()
	if err != nil {
		return err
	}
	structure, err := structureResponse.Decode()
	if err != nil {
		return err
	}
	containerRuntimeID := structure.ForeGroundRtidNested()[0]
	// containerNEMCRuntimeID := chunk.StandardRuntimeIDToNEMCRuntimeID(containerRuntimeID)
	if containerRuntimeID == blocks.AIR_RUNTIMEID {
		return fmt.Errorf("block of %v @ %v (nemc) not found, should be anvil", pos, containerRuntimeID)
	}
	defer deferActionStand()
	defer deferAction()
	o.microAction.SleepTick(1)
	return o.microAction.UseAnvil(pos, containerRuntimeID, slot, newName)
}

func (o *BotActionHighLevel) HighLevelEnchantItem(slot uint8, enchants supported_item.Enchants) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelEnchantItem(slot, enchants)
}

func (o *BotActionHighLevel) highLevelEnchantItem(slot uint8, enchants supported_item.Enchants) (err error) {
	o.microAction.SelectHotBar(slot)
	o.microAction.SleepTick(1)
	results := make(chan *packet.CommandOutput, len(enchants))
	for nameOrID, level := range enchants {
		o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("enchant @s %d %v", nameOrID, level)).SetTimeout(time.Second * 3).AsyncGetResult(func(output *packet.CommandOutput, err error) {
			if err != nil {
				output = nil
			}
			results <- output
		})
	}
	for i := 0; i < len(enchants); i++ {
		r := <-results
		if r == nil || r.SuccessCount == 0 {
			err = fmt.Errorf("some enchant command fail")
		}
	}
	return
}

func (o *BotActionHighLevel) highLevelListenItemPicked(ctx context.Context) (actionChan chan protocol.InventoryAction, err error) {
	go func() {
		<-ctx.Done()
		o.pickedItemChan = nil
	}()
	o.pickedItemChan = make(chan protocol.InventoryAction, 64)
	return o.pickedItemChan, nil
}

func (o *BotActionHighLevel) HighLevelListenItemPicked(timeout time.Duration) (actionChan chan protocol.InventoryAction, cancel func(), err error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return nil, func() {}, err
	}
	go func() {
		<-ctx.Done()
		o.pickedItemChan = nil
		release()
	}()
	o.pickedItemChan = make(chan protocol.InventoryAction, 64)
	return o.pickedItemChan, cancel, nil
}

func (o *BotActionHighLevel) HighLevelPickBlock(pos define.CubePos, targetHotBar uint8, retryTimes int) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelPickBlock(pos, targetHotBar, retryTimes)
}

func (o *BotActionHighLevel) highLevelPickBlock(pos define.CubePos, targetHotBar uint8, retryTimes int) error {
	if err := o.highLevelEnsureBotNearby(pos.Add(define.CubePos{0, 2, 0}), 3); err != nil {
		return err
	}
	defer func() {
		o.playerHotBarChan = nil
	}()
	o.playerHotBarChan = make(chan *packet.PlayerHotBar, 64)
	for i := 0; i < retryTimes; i++ {
		o.microAction.SelectHotBar(targetHotBar)
		o.ctrl.SendPacket(&packet.BlockPickRequest{
			Position:    protocol.BlockPos{int32(pos.X()), int32(pos.Y()), int32(pos.Z())},
			AddBlockNBT: true,
			HotBarSlot:  targetHotBar,
		})
		select {
		case pk := <-o.playerHotBarChan:
			actualSlot := pk.SelectedHotBarSlot
			if actualSlot == uint32(targetHotBar) {
				return nil
			} else {
				if err := o.microAction.MoveItemInsideHotBarOrInventory(uint8(actualSlot), targetHotBar, 1); err == nil {
					return nil
				}
			}
			break
		case <-time.NewTimer(time.Second).C:
			break
		}
	}
	return fmt.Errorf("cannot pick block within specific retry times")
}

func (o *BotActionHighLevel) HighLevelBlockBreakAndPickInHotBar(pos define.CubePos, recoverBlock bool, targetSlot uint8, maxRetriesTotal int) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelBlockBreakAndPickInHotBar(pos, recoverBlock, targetSlot, maxRetriesTotal)
}

func (o *BotActionHighLevel) highLevelBlockBreakAndPickInHotBar(pos define.CubePos, recoverBlock bool, targetSlot uint8, maxRetriesTotal int) (err error) {
	if err := o.highLevelEnsureBotNearby(pos.Add(define.CubePos{0, 2, 0}), 3); err != nil {
		return err
	}
	o.cmdSender.SendWebSocketCmdNeedResponse("clear @s").BlockGetResult()
	o.microAction.SelectHotBar(0)
	o.microAction.SleepTick(2)
	// split targets
	o.microAction.SleepTick(2)
	if ret, err := o.cmdSender.SendWebSocketCmdNeedResponse(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z())).SetTimeout(time.Second * 3).BlockGetResult(); ret == nil || err != nil {
		return fmt.Errorf("cannot make bot to target position")
	}
	if ret, err := o.cmdSender.SendWebSocketCmdNeedResponse("tp @e[type=item,r=9] ~ -100 ~").SetTimeout(time.Second * 3).BlockGetResult(); ret == nil || err != nil {
		return fmt.Errorf("cannot clean bot nearby items")
	}

	ctx, cancelListen := context.WithCancel(context.Background())
	defer cancelListen()
	actionChan, err := o.highLevelListenItemPicked(ctx)
	if err != nil {
		return err
	}
	recoverAction, err := o.highLevelRemoveSpecificBlockSideEffect(pos, "_temp_break"+o.nextCountName())
	if err != nil {
		return err
	}
	defer func() {
		if err != nil || recoverBlock {
			recoverAction()
		}
	}()
	totalTimes := 1 + maxRetriesTotal
	for tryTime := 0; tryTime < totalTimes; tryTime++ {
		thisTimeOk := false
		pickedSlot := -1
		// move and break immediately
		o.cmdSender.SendWOCmd(fmt.Sprintf("tp @s %v %v %v", pos.X(), pos.Y(), pos.Z()))
		o.cmdSender.SendWOCmd(fmt.Sprintf("setblock %v %v %v air 0 destroy", pos.X(), pos.Y(), pos.Z()))
		// listen block picked (wait at most 3s)
		select {
		case <-time.NewTicker(time.Second * 1).C:
			// get no item
		case pickedAction := <-actionChan:
			// item acquired
			pickedSlot = int(pickedAction.InventorySlot)
			// fmt.Println(pickedSlot)
			// we need to check if anything not wanted also dropped into inventory
			o.microAction.SleepTick(5)
			// eat up all unwanted actions
			hasInreleventItem := false
			for {
				notWantedExist := false
				select {
				case <-actionChan:
					notWantedExist = true
					hasInreleventItem = true
					break
				default:
				}
				if !notWantedExist {
					break
				}
			}
			if !hasInreleventItem {
				thisTimeOk = true
			} else {
				return fmt.Errorf("this is not a simple block or a shulker box, but a block with contents, which cannot be get in single slot")
			}
		}

		if thisTimeOk {
			// check if item is in slot we want
			if pickedSlot == int(targetSlot) {
				return
				// lucky
				// fmt.Println("get in slot ", pickedSlot)
			} else {
				// move item
				// fmt.Printf("move %v -> %v\n", pickedSlot, targetSlot)
				if err = o.microAction.MoveItemInsideHotBarOrInventory(uint8(pickedSlot), uint8(targetSlot), 1); err != nil {
					// maybe something block this slot
					o.cmdHelper.ReplaceHotBarItemCmd(int32(targetSlot), "air").SendAndGetResponse().SetTimeout(time.Second).BlockGetResult()
					o.microAction.SleepTick(1)
					if err = o.microAction.MoveItemInsideHotBarOrInventory(uint8(pickedSlot), uint8(targetSlot), 1); err != nil {
						// oh no
						thisTimeOk = false
					} else {
						// ok
						return nil
					}
				}
			}
		}

		// fmt.Printf("do time: %v ok: %v slot: %v\n", tryTime, thisTimeOk, pickedSlot)
		if tryTime == totalTimes-1 {
			break
		}
		recoverAction()
		if !thisTimeOk {
			if ret, err := o.cmdSender.SendWebSocketCmdNeedResponse("tp @e[type=item,r=9] ~ -100 ~").SetTimeout(time.Second * 3).BlockGetResult(); ret == nil || err != nil {
				return fmt.Errorf("cannot clean bot nearby items")
			}
			o.cmdHelper.ReplaceHotBarItemCmd(int32(targetSlot), "air").SendAndGetResponse().BlockGetResult()
		}
		o.microAction.SleepTick(10)
	}
	o.microAction.SleepTick(5)
	return fmt.Errorf("not all slots successfully get block")
}

func (o *BotActionHighLevel) HighLevelWriteBook(slotID uint8, pages []string) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelWriteBook(slotID, pages)
}

func (o *BotActionHighLevel) highLevelWriteBook(slotID uint8, pages []string) (err error) {
	rest, err := o.cmdHelper.ReplaceHotBarItemCmd(int32(slotID), "writable_book").SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult()
	if rest == nil || err != nil {
		return fmt.Errorf("cannot get writable_book in slot")
	}
	o.microAction.SleepTick(1)
	o.microAction.UseHotBarItem(slotID)
	o.microAction.SleepTick(1)
	for page, data := range pages {
		o.ctrl.SendPacket(&packet.BookEdit{
			ActionType:    packet.BookActionReplacePage,
			InventorySlot: slotID,
			Text:          data,
			PageNumber:    uint8(page),
		})
	}
	o.microAction.SleepTick(1)
	return nil
}

func (o *BotActionHighLevel) HighLevelWriteBookAndClose(slotID uint8, pages []string, bookTitle string, bookAuthor string) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	// its wired, we must do this or bot cannot generate book
	o.microAction.SelectHotBar(0)
	return o.highLevelWriteBookAndClose(slotID, pages, bookTitle, bookAuthor)
}

func (o *BotActionHighLevel) highLevelWriteBookAndClose(slotID uint8, pages []string, bookTitle string, bookAuthor string) (err error) {
	if err := o.highLevelWriteBook(slotID, pages); err != nil {
		return err
	}
	o.ctrl.SendPacket(&packet.BookEdit{
		ActionType:    packet.BookActionSign,
		InventorySlot: slotID,
		Title:         bookTitle,
		Author:        bookAuthor,
	})
	return nil
}

func (o *BotActionHighLevel) HighLevelPlaceItemFrameItem(pos define.CubePos, slotID uint8, rotation int) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelPlaceItemFrameItem(pos, slotID, rotation)
}

func (o *BotActionHighLevel) highLevelPlaceItemFrameItem(pos define.CubePos, slotID uint8, rotation int) (err error) {
	var clickCount int

	o.highLevelEnsureBotNearby(pos.Add(define.CubePos{0, 0, 0}), 0)
	o.microAction.SleepTick(2)
	block, err := o.areaRequester.LowLevelRequestStructure(pos, define.CubePos{1, 1, 1}, o.nextCountName()).SetTimeout(time.Second * 3).BlockGetResult()
	if err != nil {
		return err
	}
	decoded, err := block.Decode()
	if err != nil {
		panic(err)
	}
	runtimeID := decoded.ForeGroundRtidNested()[0]
	// nemcRuntimeID := chunk.StandardRuntimeIDToNEMCRuntimeID(runtimeID)
	if len(decoded.ForeGroundRtidNested()) == 0 {
		return fmt.Errorf("item frame nbt not found")
	}
	_, states, _ := blocks.RuntimeIDToState(decoded.ForeGroundRtidNested()[0])
	face, ok := states["facing_direction"]
	if !ok {
		return fmt.Errorf("facing not found")
	}
	facing, ok := face.(int32)
	if !ok {
		return fmt.Errorf("facing not found")
	}
	o.microAction.SleepTick(1)

	if rotation < 45 {
		clickCount = 1 + rotation
	} else {
		clickCount = 1 + int(rotation/45)
	}

	for i := 0; i < clickCount; i++ {
		_ = o.microAction.UseHotBarItemOnBlock(pos, runtimeID, facing, slotID)
	}

	return nil
}

func (o *BotActionHighLevel) HighLevelSetContainerContent(pos define.CubePos, containerInfo map[uint8]*supported_item.ContainerSlotItemStack) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	o.highLevelEnsureBotNearby(pos.Add(define.CubePos{0, 2, 0}), 3)
	return o.highLevelSetContainerItems(pos, containerInfo)
}

func (o *BotActionHighLevel) HighLevelGenContainer(pos define.CubePos, containerInfo map[uint8]*supported_item.ContainerSlotItemStack, block string) (err error) {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	o.highLevelEnsureBotNearby(pos.Add(define.CubePos{0, 2, 0}), 3)

	/*
		Solve chest and trapped_chest related problems,
		because the settings command is async.

		We need ensure that the current block we will place is
		after the last block we placed.

		Note
			1. To split a large chest/trapped_chest to single,
			   we usually place a trapped_chest(chest) block
			   before the chest(trapped_chest) block.
			2. These features are come from PhoenixBuilder,
			   but not neomega-builder.

		--Happy2018new
	*/
	if strings.Contains(block, "chest") {
		o.microAction.SleepTick(5)
	}

	if ret, err := o.cmdHelper.SetBlockCmd(pos, block).AsWebSocket().SendAndGetResponse().SetTimeout(time.Second * 3).BlockGetResult(); ret == nil || err != nil {
		return fmt.Errorf("cannot set container")
	}
	return o.highLevelSetContainerItems(pos, containerInfo)
}

func (o *BotActionHighLevel) HighLevelMakeItem(item *supported_item.Item, slotID uint8, anvilPos, nextContainerPos define.CubePos) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelMakeItem(item, slotID, anvilPos, nextContainerPos)
}

func (o *BotActionHighLevel) highLevelMakeItem(item *supported_item.Item, slotID uint8, anvilPos, nextContainerPos define.CubePos) error {
	typeDescription := item.GetTypeDescription()
	if !typeDescription.IsComplexBlock() {
		if typeDescription.KnownItem() == supported_item.KnownItemWritableBook {
			if err := o.highLevelWriteBook(uint8(slotID), item.SpecificKnownNonBlockItemData.Pages); err != nil {
				return err
			}
		} else if typeDescription.KnownItem() == supported_item.KnownItemWrittenBook {
			if err := o.highLevelWriteBookAndClose(uint8(slotID), item.SpecificKnownNonBlockItemData.Pages, item.SpecificKnownNonBlockItemData.BookName, item.SpecificKnownNonBlockItemData.BookAuthor); err != nil {
				return err
			}
		} else {
			if ret, err := o.cmdHelper.ReplaceBotHotBarItemFullCmd(int32(slotID), item.Name, 1, int32(item.Value), item.BaseProps).SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult(); ret == nil || err != nil {
				return fmt.Errorf("cannot put simple block/item in container %v %v %v %v", item.Name, 1, int32(item.Value), item.BaseProps)
			}
		}
		if item.DisplayName != "" {
			o.highLevelRenameItemWithAnvil(anvilPos, slotID, item.DisplayName, true)
		}
		if len(item.Enchants) > 0 {
			if err := o.highLevelEnchantItem(slotID, item.Enchants); err != nil {
				return err
			}
		}
	} else {
		deferActionWorkspace, _ := o.highLevelRemoveSpecificBlockSideEffect(nextContainerPos, o.nextCountName())
		defer deferActionWorkspace()
		o.cmdHelper.SetBlockCmd(nextContainerPos, fmt.Sprintf("%v %v", item.Name, item.RelatedBlockBedrockStateString)).AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
		if err := o.highLevelSetContainerItems(nextContainerPos, item.RelateComplexBlockData.Container); err != nil {
			return err
		}
		if err := o.highLevelPickBlock(nextContainerPos, slotID, 3); err != nil {
			return err
		}
		// if err := o.highLevelBlockBreakAndPickInHotBar(nextContainerPos, false, slotID, 2); err != nil {
		// 	return err
		// }
		// give complex block enchant and name
		if len(item.Enchants) > 0 {
			if err := o.highLevelEnchantItem(slotID, item.Enchants); err != nil {
				return err
			}
		}
		if item.DisplayName != "" {
			if err := o.highLevelRenameItemWithAnvil(anvilPos, slotID, item.DisplayName, true); err != nil {
				return err
			}
		}
	}
	return nil
}

func (o *BotActionHighLevel) highLevelSetContainerItems(pos define.CubePos, containerInfo map[uint8]*supported_item.ContainerSlotItemStack) (err error) {
	o.highLevelEnsureBotNearby(pos.Add(define.CubePos{0, 2, 0}), 3)
	updateErr := func(newErr error) {
		if newErr == nil {
			return
		}
		if err == nil {
			err = newErr
		} else {
			err = fmt.Errorf("%v\n%v", err.Error(), newErr.Error())
		}
	}
	targetContainerPos := pos
	possibleBlockerPos := targetContainerPos.Add(define.CubePos{0, 1, 0})
	anvilPos := targetContainerPos.Add(define.CubePos{1, 0, -1})
	nextContainerPos := targetContainerPos.Add(define.CubePos{1, 0, 1})
	recoverPossibleBlocker, _ := o.highLevelRemoveSpecificBlockSideEffect(possibleBlockerPos, o.nextCountName())
	defer recoverPossibleBlocker()
	// put simple block/item in container first
	for slot, stack := range containerInfo {
		if stack.Item.GetTypeDescription().IsSimple() {
			if ret, err := o.cmdHelper.ReplaceContainerItemFullCmd(targetContainerPos, int32(slot), stack.Item.Name, stack.Count, int32(stack.Item.Value), stack.Item.BaseProps).SendAndGetResponse().BlockGetResult(); ret == nil || err != nil {
				updateErr(fmt.Errorf("cannot put simple block/item in container %v %v %v %v", stack.Item.Name, stack.Count, int32(stack.Item.Value), stack.Item.BaseProps))
			}
		}
	}
	// put block/item needs only enchant in container
	hotBarSlotID := 0
	slotAndEnchant := map[uint8]*supported_item.ContainerSlotItemStack{}
	targetSlots := map[uint8]uint8{}
	flush := func() {
		if len(targetSlots) == 0 {
			return
		}
		// wait for a very short time
		o.microAction.SleepTick(5)
		// enchant & rename

		deferActionStand := func() {}
		deferAction := func() {}

		for _, stack := range slotAndEnchant {
			if stack.Item.DisplayName != "" {
				deferActionStand, _ = o.highLevelRemoveSpecificBlockSideEffect(anvilPos.Add(define.CubePos{0, -1, 0}), o.nextCountName())
				o.cmdHelper.SetBlockCmd(anvilPos.Add(define.CubePos{0, -1, 0}), "glass").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
				deferAction, _ = o.highLevelRemoveSpecificBlockSideEffect(anvilPos, o.nextCountName())
				o.cmdHelper.SetBlockCmd(anvilPos, "anvil").AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
				break
			}
		}

		defer deferActionStand()
		defer deferAction()

		for hotBarSlot, stack := range slotAndEnchant {
			if len(stack.Item.Enchants) > 0 {
				updateErr(o.highLevelEnchantItem(hotBarSlot, stack.Item.Enchants))
			}
			if stack.Item.DisplayName != "" {
				updateErr(o.highLevelRenameItemWithAnvil(anvilPos, hotBarSlot, stack.Item.DisplayName, false))
			}
		}
		// swap
		updateErr(o.highLevelMoveItemToContainer(targetContainerPos, targetSlots))
		// reset
		hotBarSlotID = 0
		slotAndEnchant = map[uint8]*supported_item.ContainerSlotItemStack{}
		targetSlots = map[uint8]uint8{}
		o.microAction.SleepTick(5)
	}
	for _slot, _stack := range containerInfo {
		slot, stack := _slot, _stack
		typeDescription := stack.Item.GetTypeDescription()
		if typeDescription.NeedHotbar() && !typeDescription.IsComplexBlock() {
			if typeDescription.KnownItem() == supported_item.KnownItemWritableBook {
				updateErr(o.highLevelWriteBook(uint8(hotBarSlotID), stack.Item.SpecificKnownNonBlockItemData.Pages))
			} else if typeDescription.KnownItem() == supported_item.KnownItemWrittenBook {
				updateErr(o.highLevelWriteBookAndClose(uint8(hotBarSlotID), stack.Item.SpecificKnownNonBlockItemData.Pages, stack.Item.SpecificKnownNonBlockItemData.BookName, stack.Item.SpecificKnownNonBlockItemData.BookAuthor))
			} else {
				if ret, err := o.cmdHelper.ReplaceBotHotBarItemFullCmd(int32(hotBarSlotID), stack.Item.Name, stack.Count, int32(stack.Item.Value), stack.Item.BaseProps).SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult(); ret == nil || err != nil {
					updateErr(fmt.Errorf("cannot put simple block/item in container %v %v %v %v", stack.Item.Name, stack.Count, int32(stack.Item.Value), stack.Item.BaseProps))
				}
			}
			slotAndEnchant[uint8(hotBarSlotID)] = stack
			targetSlots[uint8(hotBarSlotID)] = slot
			hotBarSlotID++
			if hotBarSlotID == 8 {
				flush()
			}
		}
	}
	flush()
	for _slot, _stack := range containerInfo {
		slot, stack := _slot, _stack
		if stack.Item.GetTypeDescription().IsComplexBlock() {
			o.microAction.SleepTick(5)
			deferActionWorkspace, _ := o.highLevelRemoveSpecificBlockSideEffect(nextContainerPos, o.nextCountName())
			defer deferActionWorkspace()
			o.cmdHelper.SetBlockCmd(nextContainerPos, fmt.Sprintf("%v %v", stack.Item.Name, stack.Item.RelatedBlockBedrockStateString)).AsWebSocket().SendAndGetResponse().SetTimeout(3 * time.Second).BlockGetResult()
			o.microAction.SleepTick(5)
			updateErr(o.highLevelSetContainerItems(nextContainerPos, stack.Item.RelateComplexBlockData.Container))
			err := o.highLevelPickBlock(nextContainerPos, 0, 3)
			// err := o.highLevelBlockBreakAndPickInHotBar(nextContainerPos, false, 0, 3)
			updateErr(err)
			// give complex block enchant and name
			if len(stack.Item.Enchants) > 0 {
				updateErr(o.highLevelEnchantItem(0, stack.Item.Enchants))
			}
			if stack.Item.DisplayName != "" {
				updateErr(o.highLevelRenameItemWithAnvil(anvilPos, 0, stack.Item.DisplayName, true))
			}
			updateErr(o.highLevelMoveItemToContainer(targetContainerPos, map[uint8]uint8{0: slot}))
		}
	}
	return
}

func (o *BotActionHighLevel) HighLevelRequestLargeArea(startPos define.CubePos, size define.CubePos, dst chunks.ChunkProvider, withMove bool) error {
	release, err := o.occupyBot(time.Second * 3)
	if err != nil {
		return err
	}
	defer release()
	return o.highLevelRequestLargeArea(startPos, size, dst, withMove)
}

func (o *BotActionHighLevel) highLevelRequestLargeArea(startPos define.CubePos, size define.CubePos, dst chunks.ChunkProvider, withMove bool) error {
	chunkRangesX := pos_operations.RangeSplits(startPos.X(), size.X(), 16)
	chunkRangesZ := pos_operations.RangeSplits(startPos.Z(), size.Z(), 16)
	for _, xRange := range chunkRangesX {
		startX := xRange[0]
		for _, zRange := range chunkRangesZ {
			startZ := zRange[0]
			if withMove {
				o.highLevelEnsureBotNearby(define.CubePos{startX, 320, startZ}, 16)
			}
			var err error
			for i := 0; i < 3; i++ {
				var resp neomega.StructureResponse
				var structure neomega.DecodedStructure
				if err != nil {
					time.Sleep(time.Second)
				}
				resp, err = o.areaRequester.LowLevelRequestStructure(define.CubePos{startX, startPos.Y(), startZ}, define.CubePos{xRange[1], size.Y(), zRange[1]}, "_tmp"+o.nextCountName()).BlockGetResult()
				if err != nil {
					continue
				}
				structure, err = resp.Decode()
				if err != nil {
					continue
				}
				err = structure.DumpToChunkProviderAbsolutePos(dst)
				if err != nil {
					return err
				}
			}
			if err != nil {
				return err
			}
		}
	}
	return nil
}
