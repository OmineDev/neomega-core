package neomega

import (
	"time"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega/chunks"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
	"github.com/OmineDev/neomega-core/neomega/supported_nbt_data"
	"github.com/OmineDev/neomega-core/neomega/supported_nbt_data/supported_item"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"

	"github.com/go-gl/mathgl/mgl32"
)

type PosAndDimensionInfo struct {
	Dimension      int
	InOverWorld    bool
	InNether       bool
	InEnd          bool
	HeadPosPrecise mgl32.Vec3
	FeetPosPrecise mgl32.Vec3
	HeadBlockPos   define.CubePos
	FeetBlockPos   define.CubePos
	YRot           float64
}

// information safe to be set from outside & when setting, no bot action module is required
// only information which is shared and *read only* by different modules should be placed here
type BotActionInfo struct {
	*PosAndDimensionInfo
	TempStructureName string
}

type CmdCannotGetResponse interface {
	Send()
}

type CmdCanGetResponse interface {
	CmdCannotGetResponse
	SendAndGetResponse() async_wrapper.AsyncResult[*packet.CommandOutput]
}

type GeneralCommand interface {
	Send()
	AsPlayer() CmdCanGetResponse
	AsWebSocket() CmdCanGetResponse
}

// HighLevel means it contains complex steps
// and most importantly, not all details are disclosed!
type BotActionHighLevel interface {
	HighLevelPlaceSign(targetPos define.CubePos, signBlock string, data *supported_nbt_data.SignBlockSupportedData) (err error)
	HighLevelPlaceCommandBlock(targetPos define.CubePos, option *supported_nbt_data.CommandBlockSupportedData, maxRetry int) error
	HighLevelMoveItemToContainer(pos define.CubePos, moveOperations map[uint8]uint8) error
	HighLevelEnsureBotNearby(pos define.CubePos, threshold float32) error
	HighLevelRemoveSpecificBlockSideEffect(pos define.CubePos, backupName string) (deferFunc func(), err error)
	HighLevelRenameItemWithAnvil(pos define.CubePos, slot uint8, newName string, autoGenAnvil bool) (err error)
	HighLevelEnchantItem(slot uint8, enchants supported_item.Enchants) (err error)
	HighLevelListenItemPicked(timeout time.Duration) (actionChan chan protocol.InventoryAction, cancel func(), err error)
	HighLevelPickBlock(pos define.CubePos, targetHotBar uint8, retryTimes int) error
	HighLevelBlockBreakAndPickInHotBar(pos define.CubePos, recoverBlock bool, targetSlot uint8, maxRetriesTotal int) (err error)
	HighLevelSetContainerContent(pos define.CubePos, containerInfo map[uint8]*supported_item.ContainerSlotItemStack) (err error)
	HighLevelGenContainer(pos define.CubePos, containerInfo map[uint8]*supported_item.ContainerSlotItemStack, block string) (err error)
	HighLevelWriteBook(slotID uint8, pages []string) (err error)
	HighLevelWriteBookAndClose(slotID uint8, pages []string, bookTitle string, bookAuthor string) (err error)
	HighLevelPlaceItemFrameItem(pos define.CubePos, slotID uint8) error
	HighLevelMakeItem(item *supported_item.Item, slotID uint8, anvilPos, nextContainerPos define.CubePos) error
	HighLevelRequestLargeArea(startPos define.CubePos, size define.CubePos, dst chunks.ChunkProvider, withMove bool) error
}

type BotAction interface {
	SelectHotBar(slotID uint8) error
	SleepTick(ticks int)
	UseHotBarItem(slot uint8) (err error)
	UseHotBarItemOnBlock(blockPos define.CubePos, blockNEMCRuntimeID uint32, face int32, slot uint8) (err error)
	UseHotBarItemOnBlockWithBotOffset(blockPos define.CubePos, botOffset define.CubePos, blockNEMCRuntimeID uint32, face int32, slot uint8) (err error)
	// TapBlockUsingHotBarItem(blockPos define.CubePos, blockNEMCRuntimeID uint32, slotID uint8) (err error)
	MoveItemFromInventoryToEmptyContainerSlots(pos define.CubePos, blockNemcRtid uint32, blockName string, moveOperations map[uint8]uint8) error
	UseAnvil(pos define.CubePos, blockNemcRtid uint32, slot uint8, newName string) error
	DropItemFromHotBar(slot uint8) error
	MoveItemInsideHotBarOrInventory(sourceSlot, targetSlot, count uint8) error
	SetStructureBlockData(pos define.CubePos, settings *supported_nbt_data.StructureBlockSupportedData)
	GetInventoryContent(windowID uint32, slotID uint8) (instance *protocol.ItemInstance, found bool)
}

type CommandHelper interface {
	// if in overworld, send in WO manner, otherwise send in websocket manner
	ConstructDimensionLimitedWOCommand(cmd string) CmdCannotGetResponse
	// if in overworld, Send() in WO manner, otherwise Send() in websocket manner
	ConstructDimensionLimitedGeneralCommand(cmd string) GeneralCommand
	ConstructGeneralCommand(cmd string) GeneralCommand
	// an uuid string but replaced by special chars
	GenAutoUnfilteredUUID() string
	ReplaceHotBarItemCmd(slotID int32, item string) CmdCanGetResponse
	ReplaceBotHotBarItemFullCmd(slotID int32, itemName string, count uint8, value int32, components *supported_item.ItemPropsInGiveOrReplace) CmdCanGetResponse
	ReplaceContainerBlockItemCmd(pos define.CubePos, slotID int32, item string) CmdCanGetResponse
	ReplaceContainerItemFullCmd(pos define.CubePos, slotID int32, itemName string, count uint8, value int32, components *supported_item.ItemPropsInGiveOrReplace) CmdCanGetResponse

	BackupStructureWithGivenNameCmd(start define.CubePos, size define.CubePos, name string) CmdCanGetResponse
	BackupStructureWithAutoNameCmd(start define.CubePos, size define.CubePos) (name string, cmd CmdCanGetResponse)
	RevertStructureWithGivenNameCmd(start define.CubePos, name string) CmdCanGetResponse

	SetBlockCmd(pos define.CubePos, blockString string) GeneralCommand
	SetBlockRelativeCmd(pos define.CubePos, blockString string) GeneralCommand
	FillBlocksWithRangeCmd(startPos define.CubePos, endPos define.CubePos, blockString string) GeneralCommand
	FillBlocksWithSizeCmd(startPos define.CubePos, size define.CubePos, blockString string) GeneralCommand
}

type BotActionComplex interface {
	CommandHelper
	BotAction
	BotActionHighLevel
}
