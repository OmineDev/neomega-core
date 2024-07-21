package main

/*
#include <stdio.h>
#include <stdlib.h>
#include <string.h>
*/
import "C"
import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"sync"
	"time"
	"unsafe"

	"github.com/OmineDev/neomega-core/i18n"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/bundle"
	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/access_helper"
	"github.com/OmineDev/neomega-core/neomega/rental_server_impact/info_collect_utils"
	"github.com/OmineDev/neomega-core/nodes"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/nodes/underlay_conn"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
)

var GOmegaCore neomega.MicroOmega
var GPacketNameIDMapping map[string]uint32
var GPacketIDNameMapping map[uint32]string
var GPool packet.Pool

//export OmegaAvailable
func OmegaAvailable() bool {
	return true
}

const (
	EventTypeOmegaConnErr           = "OmegaConnErr"
	EventTypeCommandResponseCB      = "CommandResponseCB"
	EventTypeNewPacket              = "MCPacket"
	EventTypePlayerInterceptedInput = "PlayerInterceptInput"
	EventTypePlayerChange           = "PlayerChange"
	EventTypeChat                   = "Chat"
	EventTypeNamedCommandBlockMsg   = "NamedCommandBlockMsg"
)

type GEvent struct {
	EventType string
	// 调用语言可能希望将某个回调绑定到事件回调上，由于我们无法将事件传入 go 内部，
	// 因此此字段用以帮助调用语言找到目标回调
	RetrieverID string
	// Golang 有自己的GC，当一个数据从 GO 的视野中消失的时候，得假设它可能已经被回收了
	// 然而，实际上event的数据可以是任何类型的
	// 我们要求外部语言在通过 epoll 知道一个 event 发生（知道 event 的 emitter 和 retrieve id 后）
	// 立刻调用 omit event 忽略这个事件（这样 go 也可以顺利回收资源）
	// 或者 consume xxx 将这个事件立刻转为特定的类型
	Data any
}

var GEventsChan chan *GEvent
var GCurrentEvent *GEvent

//export EventPoll
func EventPoll() (EventType *C.char, RetrieverID *C.char) {
	e := <-GEventsChan
	GCurrentEvent = e
	return C.CString(e.EventType), C.CString(e.RetrieverID)
}

//export OmitEvent
func OmitEvent() {
	GCurrentEvent = nil
}

// Async Actions

//export ConsumeCommandResponseCB
func ConsumeCommandResponseCB() *C.char {
	p := (GCurrentEvent.Data).(*packet.CommandOutput)
	bs, _ := json.Marshal(p)
	return C.CString(string(bs))
}

//export SendWebSocketCommandNeedResponse
func SendWebSocketCommandNeedResponse(cmd *C.char, retrieverID *C.char) {
	GoRetrieverID := C.GoString(retrieverID)
	GOmegaCore.GetGameControl().SendWebSocketCmdNeedResponse(C.GoString(cmd)).AsyncGetResult(func(p *packet.CommandOutput, err error) {
		if err != nil {
			p = nil
		}
		GEventsChan <- &GEvent{EventTypeCommandResponseCB, GoRetrieverID, p}
	})
}

//export SendPlayerCommandNeedResponse
func SendPlayerCommandNeedResponse(cmd *C.char, retrieverID *C.char) {
	GoRetrieverID := C.GoString(retrieverID)
	GOmegaCore.GetGameControl().SendPlayerCmdNeedResponse(C.GoString(cmd)).AsyncGetResult(func(p *packet.CommandOutput, err error) {
		if err != nil {
			p = nil
		}
		GEventsChan <- &GEvent{EventTypeCommandResponseCB, GoRetrieverID, p}
	})
}

// One-Way Action

//export SendWOCommand
func SendWOCommand(cmd *C.char) {
	GOmegaCore.GetGameControl().SendWOCmd(C.GoString(cmd))
}

//export SendWebSocketCommandOmitResponse
func SendWebSocketCommandOmitResponse(cmd *C.char) {
	GOmegaCore.GetGameControl().SendWebSocketCmdOmitResponse(C.GoString(cmd))
}

//export SendPlayerCommandOmitResponse
func SendPlayerCommandOmitResponse(cmd *C.char) {
	GOmegaCore.GetGameControl().SendPlayerCmdOmitResponse(C.GoString(cmd))
}

//export SendGamePacket
func SendGamePacket(packetID int, jsonStr *C.char) (err *C.char) {
	pk := GPool[uint32(packetID)]()
	_err := json.Unmarshal([]byte(C.GoString(jsonStr)), &pk)
	if _err != nil {
		return C.CString(_err.Error())
	}
	GOmegaCore.GetGameControl().SendPacket(pk)
	return C.CString("")
}

type NoEOFByteReader struct {
	s []byte
	i int
}

func (nbr *NoEOFByteReader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return 0, nil
	}
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	n = copy(b, nbr.s[nbr.i:])
	nbr.i += n
	return
}

func (nbr *NoEOFByteReader) ReadByte() (b byte, err error) {
	if nbr.i >= len(nbr.s) {
		return 0, io.EOF
	}
	b = nbr.s[nbr.i]
	nbr.i++
	return b, nil
}

func bytesToCharArr(goByteSlice []byte) *C.char {
	ptr := C.malloc(C.size_t(len(goByteSlice)))
	C.memmove(ptr, (unsafe.Pointer)(&goByteSlice[0]), C.size_t(len(goByteSlice)))
	return (*C.char)(ptr)
}

//export JsonStrAsIsGamePacketBytes
func JsonStrAsIsGamePacketBytes(packetID int, jsonStr *C.char) (pktBytes *C.char, l int, err *C.char) {
	pk := GPool[uint32(packetID)]()
	_err := json.Unmarshal([]byte(C.GoString(jsonStr)), &pk)
	if _err != nil {
		return nil, 0, C.CString(_err.Error())
	}
	b := &bytes.Buffer{}
	w := protocol.NewWriter(b, 0)
	// hdr := pk.ID()
	// w.Varuint32(&hdr)
	pk.Marshal(w)
	bs := b.Bytes()
	l = len(bs)
	return bytesToCharArr(bs), l, nil
}

//export PlaceCommandBlock
func PlaceCommandBlock(option *C.char) {
	opt := neomega.PlaceCommandBlockOption{}
	json.Unmarshal([]byte(C.GoString(option)), &opt)
	GOmegaCore.GetBotAction().HighLevelPlaceCommandBlock(&opt, 3)
	// ba := GOmegaCore.GetGameControl().GenCommandBlockUpdateFromOption(&opt)
	// GOmegaCore.GetGameControl().AsyncPlaceCommandBlock(define.CubePos{
	// 	opt.X, opt.Y, opt.Z,
	// }, opt.BlockName, opt.BlockState, false, false, ba, func(done bool) {
	// 	if !done {
	// 		fmt.Printf("place command block @ [%v,%v,%v] fail\n", opt.X, opt.Y, opt.Z)
	// 	} else {
	// 		fmt.Printf("place command block @ [%v,%v,%v] ok\n", opt.X, opt.Y, opt.Z)
	// 	}
	// }, time.Second*10)
}

// listeners

// disconnect event

//export ConsumeOmegaConnError
func ConsumeOmegaConnError() *C.char {
	err := (GCurrentEvent.Data).(error)
	return C.CString(err.Error())
}

// packet event

var GAllPacketsListenerEnabled = false

//export ListenAllPackets
func ListenAllPackets() {
	if GAllPacketsListenerEnabled {
		panic("should only call ListenAllPackets once")
	}
	GAllPacketsListenerEnabled = true
	GOmegaCore.GetGameListener().SetAnyPacketCallBack(func(p packet.Packet) {
		GEventsChan <- &GEvent{
			EventType:   EventTypeNewPacket,
			RetrieverID: GPacketIDNameMapping[p.ID()],
			Data:        p,
		}
	}, true)
}

//export GetPacketNameIDMapping
func GetPacketNameIDMapping() *C.char {
	marshal, err := json.Marshal(GPacketNameIDMapping)
	if err != nil {
		panic(err)
	}
	return C.CString(string(marshal))
}

//export ConsumeMCPacket
func ConsumeMCPacket() (packetDataAsJsonStr *C.char, convertError *C.char) {
	p := (GCurrentEvent.Data).(packet.Packet)
	marshal, err := json.Marshal(p)
	packetDataAsJsonStr = C.CString(string(marshal))
	convertError = nil
	if err != nil {
		convertError = C.CString(string(err.Error()))
	}
	return
}

// Bot
//
//export GetClientMaintainedBotBasicInfo
func GetClientMaintainedBotBasicInfo() *C.char {
	basicInfo := GOmegaCore.GetMicroUQHolder().GetBotBasicInfo()
	basicInfoMap := map[string]any{
		"BotName":      basicInfo.GetBotName(),
		"BotRuntimeID": basicInfo.GetBotRuntimeID(),
		"BotUniqueID":  basicInfo.GetBotUniqueID(),
		"BotIdentity":  basicInfo.GetBotIdentity(),
		"BotUUIDStr":   basicInfo.GetBotUUIDStr(),
	}
	data, _ := json.Marshal(basicInfoMap)
	return C.CString(string(data))
}

//export GetClientMaintainedExtendInfo
func GetClientMaintainedExtendInfo() *C.char {
	extendInfo := GOmegaCore.GetMicroUQHolder().GetExtendInfo()
	extendInfoMap := map[string]any{}
	if worldName, found := extendInfo.GetWorldName(); found {
		extendInfoMap["WorldName"] = worldName
	}
	if worldSeed, found := extendInfo.GetWorldSeed(); found {
		extendInfoMap["WorldSeed"] = worldSeed
	}
	if worldGenerator, found := extendInfo.GetWorldGenerator(); found {
		extendInfoMap["WorldGenerator"] = worldGenerator
	}
	if levelID, found := extendInfo.GetLevelID(); found {
		extendInfoMap["LevelID"] = levelID
	}
	if thres, found := extendInfo.GetCompressThreshold(); found {
		extendInfoMap["CompressThreshold"] = thres
	}
	if worldGameMode, found := extendInfo.GetWorldGameMode(); found {
		extendInfoMap["WorldGameMode"] = worldGameMode
	}
	if worldDifficulty, found := extendInfo.GetWorldDifficulty(); found {
		extendInfoMap["WorldDifficulty"] = worldDifficulty
	}
	if time, found := extendInfo.GetTime(); found {
		extendInfoMap["Time"] = time
	}
	if dayTime, found := extendInfo.GetDayTime(); found {
		extendInfoMap["DayTime"] = dayTime
	}
	if timePercent, found := extendInfo.GetDayTimePercent(); found {
		extendInfoMap["TimePercent"] = timePercent
	}
	if gameRules, found := extendInfo.GetGameRules(); found {
		extendInfoMap["GameRules"] = gameRules
	}
	data, _ := json.Marshal(extendInfoMap)
	return C.CString(string(data))
}

// Player 描述单个的 neomega.PlayerKit
type Player struct {
	// 描述该结构体实际所携带的 PlayerKit 负载
	GPlayer neomega.PlayerKit
	// 描述该 Player 的引用计数
	UsingCount int
}

// 描述多个玩家的 PlayerKit 。
//
// 每当一个 Player 每新增一个使用者(如 Python 等)，
// 对应 Player 的引用计数都会加一。
//
// 当相应的使用者尝试释放一个 Player 时，
// UsingCount 将减一，
// 直到归零后，真正地被回收。
//
// 如果必要，可以考虑使用 ForceReleaseBindPlayer
// 进行强制回收，但这是危险的。
//
// 相关的引用计数不在 Go 处控制，
// 它们由使用者根据实际情况增加或减少，
// Go 处仅在引用计数归零后释放数据
type Players map[string]Player

// players
var GPlayers struct {
	Players
	sync.RWMutex
}

//export AddGPlayerUsingCount
func AddGPlayerUsingCount(uuid *C.char, delta int) {
	GPlayers.Lock()
	defer GPlayers.Unlock()

	uuidStr := C.GoString(uuid)
	player, found := GPlayers.Players[uuidStr]
	if !found {
		playerKit, found := GOmegaCore.GetPlayerInteract().GetPlayerKitByUUIDString(uuidStr)
		if !found {
			return
		}
		player = Player{GPlayer: playerKit}
	}

	player.UsingCount = player.UsingCount + delta
	GPlayers.Players[uuidStr] = player

	if player.UsingCount <= 0 {
		new := make(map[string]Player)
		delete(GPlayers.Players, uuidStr)
		for key, value := range GPlayers.Players {
			new[key] = value
		}
		GPlayers.Players = new
	}
}

//export ForceReleaseBindPlayer
func ForceReleaseBindPlayer(uuidStr *C.char) {
	GPlayers.Lock()
	defer GPlayers.Unlock()

	new := make(map[string]Player)
	delete(GPlayers.Players, C.GoString(uuidStr))
	for key, value := range GPlayers.Players {
		new[key] = value
	}
	GPlayers.Players = new
}

//export GetAllOnlinePlayers
func GetAllOnlinePlayers() *C.char {
	GPlayers.Lock()
	defer GPlayers.Unlock()

	players := GOmegaCore.GetPlayerInteract().ListAllPlayers()
	retPlayers := []string{}
	for _, player := range players {
		uuidStr, _ := player.GetUUIDString()
		retPlayers = append(retPlayers, uuidStr)
	}
	data, _ := json.Marshal(retPlayers)
	return C.CString(string(data))
}

//export GetPlayerByName
func GetPlayerByName(name *C.char) *C.char {
	GPlayers.Lock()
	defer GPlayers.Unlock()

	player, found := GOmegaCore.GetPlayerInteract().GetPlayerKit(C.GoString(name))
	if found {
		uuidStr, _ := player.GetUUIDString()
		return C.CString(uuidStr)
	}
	return C.CString("")
}

//export GetPlayerByUUID
func GetPlayerByUUID(uuid *C.char) *C.char {
	GPlayers.Lock()
	defer GPlayers.Unlock()

	player, found := GOmegaCore.GetPlayerInteract().GetPlayerKitByUUIDString(C.GoString(uuid))
	if found {
		uuidStr, _ := player.GetUUIDString()
		return C.CString(uuidStr)
	}
	return C.CString("")
}

//export PlayerName
func PlayerName(uuidStr *C.char) *C.char {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	name, _ := p.GPlayer.GetUsername()
	return C.CString(name)
}

//export PlayerEntityUniqueID
func PlayerEntityUniqueID(uuidStr *C.char) int64 {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	entityUniqueID, _ := p.GPlayer.GetEntityUniqueID()
	return entityUniqueID
}

//export PlayerLoginTime
func PlayerLoginTime(uuidStr *C.char) int64 {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	loginTime, _ := p.GPlayer.GetLoginTime()
	return loginTime.Unix()
}

//export PlayerPlatformChatID
func PlayerPlatformChatID(uuidStr *C.char) *C.char {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	name, _ := p.GPlayer.GetPlatformChatID()
	return C.CString(name)
}

//export PlayerBuildPlatform
func PlayerBuildPlatform(uuidStr *C.char) int32 {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	buildPlatform, _ := p.GPlayer.GetBuildPlatform()
	return buildPlatform
}

//export PlayerSkinID
func PlayerSkinID(uuidStr *C.char) *C.char {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	SkinID, _ := p.GPlayer.GetSkinID()
	return C.CString(SkinID)
}

// //export PlayerPropertiesFlag
// func PlayerPropertiesFlag(uuidStr *C.char) uint32 {
// 	GPlayers.RLock()
// 	defer GPlayers.RUnlock()
//
// 	p := GPlayers.Players[C.GoString(uuidStr)]
// 	PropertiesFlag, _ := p.GPlayer.GetPropertiesFlag()
// 	return PropertiesFlag
// }

// //export PlayerCommandPermissionLevel
// func PlayerCommandPermissionLevel(uuidStr *C.char) uint32 {
// 	GPlayers.RLock()
// 	defer GPlayers.RUnlock()
//
// 	p := GPlayers.Players[C.GoString(uuidStr)]
// 	CommandPermissionLevel, _ := p.GPlayer.GetCommandPermissionLevel()
// 	return CommandPermissionLevel
// }

// //export PlayerActionPermissions
// func PlayerActionPermissions(uuidStr *C.char) uint32 {
// 	GPlayers.RLock()
// 	defer GPlayers.RUnlock()
//
// 	p := GPlayers.Players[C.GoString(uuidStr)]
// 	ActionPermissions, _ := p.GPlayer.GetActionPermissions()
// 	return ActionPermissions
// }

// //export PlayerGetAbilityString
// func PlayerGetAbilityString(uuidStr *C.char) *C.char {
// 	GPlayers.RLock()
// 	defer GPlayers.RUnlock()
//
// 	p := GPlayers.Players[C.GoString(uuidStr)]
// 	adventureFlagsMap, actionPermissionMap, _ := p.GPlayer.GetAbilityString()
// 	abilityMap := map[string]map[string]bool{
// 		"AdventureFlagsMap":   adventureFlagsMap,
// 		"ActionPermissionMap": actionPermissionMap,
// 	}
// 	data, _ := json.Marshal(abilityMap)
// 	return C.CString(string(data))
// }

// //export PlayerOPPermissionLevel
// func PlayerOPPermissionLevel(uuidStr *C.char) uint32 {
// 	GPlayers.RLock()
// 	defer GPlayers.RUnlock()
//
// 	p := GPlayers.Players[C.GoString(uuidStr)]
// 	OPPermissionLevel, _ := p.GPlayer.GetOPPermissionLevel()
// 	return OPPermissionLevel
// }

// //export PlayerCustomStoredPermissions
// func PlayerCustomStoredPermissions(uuidStr *C.char) uint32 {
// 	GPlayers.RLock()
// 	defer GPlayers.RUnlock()
//
// 	p := GPlayers.Players[C.GoString(uuidStr)]
// 	CustomStoredPermissions, _ := p.GPlayer.GetCustomStoredPermissions()
// 	return CustomStoredPermissions
// }

//export PlayerCanBuild
func PlayerCanBuild(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.CanBuild()
	return hasAbility
}

//export PlayerSetBuild
func PlayerSetBuild(uuidStr *C.char, allow bool) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.SetBuildAbility(allow)
}

//export PlayerCanMine
func PlayerCanMine(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.CanMine()
	return hasAbility
}

//export PlayerSetMine
func PlayerSetMine(uuidStr *C.char, allow bool) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.SetMineAbility(allow)
}

//export PlayerCanDoorsAndSwitches
func PlayerCanDoorsAndSwitches(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.CanDoorsAndSwitches()
	return hasAbility
}

//export PlayerSetDoorsAndSwitches
func PlayerSetDoorsAndSwitches(uuidStr *C.char, allow bool) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.SetDoorsAndSwitchesAbility(allow)
}

//export PlayerCanOpenContainers
func PlayerCanOpenContainers(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.CanOpenContainers()
	return hasAbility
}

//export PlayerSetOpenContainers
func PlayerSetOpenContainers(uuidStr *C.char, allow bool) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.SetOpenContainersAbility(allow)
}

//export PlayerCanAttackPlayers
func PlayerCanAttackPlayers(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.CanAttackPlayers()
	return hasAbility
}

//export PlayerSetAttackPlayers
func PlayerSetAttackPlayers(uuidStr *C.char, allow bool) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.SetAttackPlayersAbility(allow)
}

//export PlayerCanAttackMobs
func PlayerCanAttackMobs(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.CanAttackMobs()
	return hasAbility
}

//export PlayerSetAttackMobs
func PlayerSetAttackMobs(uuidStr *C.char, allow bool) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.SetAttackMobsAbility(allow)
}

//export PlayerCanOperatorCommands
func PlayerCanOperatorCommands(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.CanOperatorCommands()
	return hasAbility
}

//export PlayerSetOperatorCommands
func PlayerSetOperatorCommands(uuidStr *C.char, allow bool) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.SetOperatorCommandsAbility(allow)
}

//export PlayerCanTeleport
func PlayerCanTeleport(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.CanTeleport()
	return hasAbility
}

//export PlayerSetTeleport
func PlayerSetTeleport(uuidStr *C.char, allow bool) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.SetTeleportAbility(allow)
}

//export PlayerStatusInvulnerable
func PlayerStatusInvulnerable(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.StatusInvulnerable()
	return hasAbility
}

//export PlayerStatusFlying
func PlayerStatusFlying(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.StatusFlying()
	return hasAbility
}

//export PlayerStatusMayFly
func PlayerStatusMayFly(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	hasAbility, _ := p.GPlayer.StatusMayFly()
	return hasAbility
}

//export PlayerDeviceID
func PlayerDeviceID(uuidStr *C.char) *C.char {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	name, _ := p.GPlayer.GetDeviceID()
	return C.CString(name)
}

//export PlayerEntityRuntimeID
func PlayerEntityRuntimeID(uuidStr *C.char) uint64 {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	EntityRuntimeID, _ := p.GPlayer.GetEntityRuntimeID()
	return EntityRuntimeID
}

//export PlayerEntityMetadata
func PlayerEntityMetadata(uuidStr *C.char) *C.char {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	entityMetadata, _ := p.GPlayer.GetEntityMetadata()
	data, _ := json.Marshal(entityMetadata)
	return C.CString(string(data))
}

//export PlayerIsOP
func PlayerIsOP(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	isOP, _ := p.GPlayer.IsOP()
	return isOP
}

//export PlayerOnline
func PlayerOnline(uuidStr *C.char) bool {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	return p.GPlayer.StillOnline()
}

//export PlayerChat
func PlayerChat(uuidStr *C.char, msg *C.char) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.Say(C.GoString(msg))
}

//export PlayerTitle
func PlayerTitle(uuidStr *C.char, title, subTitle *C.char) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.SubTitle(C.GoString(subTitle), C.GoString(title))
}

//export PlayerActionBar
func PlayerActionBar(uuidStr *C.char, actionBar *C.char) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	p.GPlayer.ActionBar(C.GoString(actionBar))
}

// //export SetPlayerAbility
// func SetPlayerAbility(uuidStr *C.char, jsonFlags *C.char) {
// 	GPlayers.RLock()
// 	defer GPlayers.RUnlock()
//
// 	p := GPlayers.Players[C.GoString(uuidStr)]
// 	// abilityMap := map[string]map[string]bool{
// 	// 	"AdventureFlagsMap":   adventureFlagsMap,
// 	// 	"ActionPermissionMap": actionPermissionMap,
// 	// }
// 	abilityMap := map[string]map[string]bool{}
// 	json.Unmarshal([]byte(C.GoString(jsonFlags)), &abilityMap)
// 	adventureFlagsMap := abilityMap["AdventureFlagsMap"]
// 	actionPermissionMap := abilityMap["ActionPermissionMap"]
// 	fmt.Println(adventureFlagsMap)
// 	fmt.Println(actionPermissionMap)
// 	p.GPlayer.SetAbilityString(adventureFlagsMap, actionPermissionMap)
// }

//export InterceptPlayerJustNextInput
func InterceptPlayerJustNextInput(uuidStr *C.char, retrieverID *C.char) {
	GPlayers.RLock()
	defer GPlayers.RUnlock()

	p := GPlayers.Players[C.GoString(uuidStr)]
	retrieverIDStr := C.GoString(retrieverID)
	p.GPlayer.GetInput().AsyncGetResult(func(chat *neomega.GameChat, err error) {
		GEventsChan <- &GEvent{
			EventType:   EventTypePlayerInterceptedInput,
			RetrieverID: retrieverIDStr,
			Data:        chat,
		}
	})
}

//export ConsumeChat
func ConsumeChat() *C.char {
	chat := GCurrentEvent.Data.(*neomega.GameChat)
	bs, _ := json.Marshal(chat)
	return C.CString(string(bs))
}

var GListenPlayerChangeListened = false

//export ListenPlayerChange
func ListenPlayerChange() {
	GPlayers.Lock()
	defer GPlayers.Unlock()

	if GListenPlayerChangeListened {
		panic("ListenPlayerChange should only called once")
	}
	GListenPlayerChangeListened = true
	GOmegaCore.GetPlayerInteract().ListenPlayerChange(func(player neomega.PlayerKit, action string) {
		uuidStr, _ := player.GetUUIDString()
		GEventsChan <- &GEvent{
			EventType:   EventTypePlayerChange,
			RetrieverID: uuidStr,
			Data:        action,
		}
	})
}

//export ConsumePlayerChange
func ConsumePlayerChange() (change *C.char) {
	return C.CString(GCurrentEvent.Data.(string))
}

var GListenChatListened = false

//export ListenChat
func ListenChat() {
	if GListenChatListened {
		panic("ListenPlayerChat should only called once")
	}
	GListenChatListened = true
	GOmegaCore.GetPlayerInteract().SetOnChatCallBack(func(chat *neomega.GameChat) {
		GEventsChan <- &GEvent{
			EventType:   EventTypeChat,
			RetrieverID: "",
			Data:        chat,
		}
	})
}

//export ListenCommandBlock
func ListenCommandBlock(name *C.char) {
	gName := C.GoString(name)
	GOmegaCore.GetPlayerInteract().SetOnSpecificCommandBlockTellCallBack(gName, func(chat *neomega.GameChat) {
		GEventsChan <- &GEvent{
			EventType:   EventTypeNamedCommandBlockMsg,
			RetrieverID: gName,
			Data:        chat,
		}
	})
}

// utils

//export FreeMem
func FreeMem(address unsafe.Pointer) {
	C.free(address)
}

func prepareOmegaAPIs(omegaCore neomega.MicroOmega) {
	GEventsChan = make(chan *GEvent, 1024)
	GOmegaCore = omegaCore
	GPacketNameIDMapping = GOmegaCore.GetGameListener().GetMCPacketNameIDMapping()
	{
		GPacketIDNameMapping = map[uint32]string{}
		for name, id := range GPacketNameIDMapping {
			GPacketIDNameMapping[id] = name
		}
	}
	GPool = packet.NewPool()
	go func() {
		err := <-omegaCore.Dead()
		GOmegaCore = nil
		GEventsChan <- &GEvent{
			EventTypeOmegaConnErr,
			"",
			err,
		}
	}()
	GPlayers.Players = make(Players)
}

//export ConnectOmega
func ConnectOmega(address *C.char) (Cerr *C.char) {
	if GOmegaCore != nil {
		return C.CString("connect has been established")
	}
	var node defines.Node
	// ctx := context.Background()
	{
		client, err := underlay_conn.NewClientFromBasicNet(C.GoString(address), time.Second)
		if err != nil {
			return C.CString(err.Error())
		}
		slave, err := nodes.NewSlaveNode(client)
		if err != nil {
			return C.CString(err.Error())
		}
		node = nodes.NewGroup("neomega", slave, false)
		if !node.CheckNetTag("access-point") {
			return C.CString(i18n.T(i18n.S_no_access_point_in_network))
		}
	}
	omegaCore, err := bundle.NewEndPointMicroOmega(node)
	if err != nil {
		return C.CString(err.Error())
	}
	prepareOmegaAPIs(omegaCore)
	return nil
}

//export StartOmega
func StartOmega(address *C.char, impactOptionsJson *C.char) (Cerr *C.char) {
	if GOmegaCore != nil {
		return C.CString("connect has been established")
	}
	var node defines.Node
	accessOption := access_helper.DefaultOptions()
	// ctx := context.Background()
	{
		impactOption := &access_helper.ImpactOption{}
		json.Unmarshal([]byte(C.GoString(impactOptionsJson)), &impactOption)
		if err := info_collect_utils.ReadUserInfoAndUpdateImpactOptions(impactOption); err != nil {
			return C.CString(err.Error())
		}

		accessOption.ImpactOption = impactOption
		accessOption.MakeBotCreative = true
		accessOption.DisableCommandBlock = false
		accessOption.ReasonWithPrivilegeStuff = true

		{
			server, err := underlay_conn.NewServerFromBasicNet(C.GoString(address))
			if err != nil {
				panic(err)
			}
			// server := nodes.NewSimpleNewMasterNodeServer(socket)
			master := nodes.NewMasterNode(server)
			node = nodes.NewGroup("neomega", master, false)
		}
	}
	ctx := context.Background()
	omegaCore, err := access_helper.ImpactServer(ctx, node, accessOption)
	if err != nil {
		return C.CString(err.Error())
	}
	prepareOmegaAPIs(omegaCore)
	return nil
}

func main() {
	//Windows: go build  -tags fbconn -o fbconn.dll -buildmode=c-shared main.go
	//Linux: go build -tags fbconn -o libfbconn.so -buildmode=c-shared main.go
	//Macos: go build -o omega_conn.dylib -buildmode=c-shared main.go
	//将生成的文件 (fbconn.dll 或 libfbconn.so 或 fbconn.dylib) 放在 conn.py 同一个目录下
}
