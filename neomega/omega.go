package neomega

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/minecraft_neo/can_close"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
	"github.com/OmineDev/neomega-core/utils/pressure_metric"

	"github.com/google/uuid"
)

// 可以向游戏发送数据包
type GameIntractable interface {
	SendPacket(packet.Packet)
}

// NoBlockAndDetachablePacketCallBack 表示没有阻塞的数据处理函数类型
// 当不需要继续读取数据时，返回 err
type NoBlockAndDetachablePacketCallback func(pk packet.Packet) error

type PacketDispatcher interface {
	SetAnyPacketCallBack(callback func(packet.Packet), newGoroutine bool)
	SetTypedPacketCallBack(packetID uint32, callback func(packet.Packet), newGoroutine bool)
	// SetOneTimeTypedPacketNoBlockCallBack(uint32, func(packet.Packet) (next bool))
	// GetMCPacketNameIDMapping 返回协议包名字与 ID 的映射关系
	GetMCPacketNameIDMapping() map[string]uint32
	// TranslateStringWantsToIDSet 将字符串形式的协议包名字列表转换为对应的协议包 ID 集合
	TranslateStringWantsToIDSet(want []string) map[uint32]bool
	AddNewNoBlockAndDetachablePacketCallBack(wants map[uint32]bool, cb NoBlockAndDetachablePacketCallback)
}

type ReactCore interface {
	can_close.CanCloseWithError
	PacketDispatcher
}

type UnStartedReactCore interface {
	ReactCore
	Start()
}

type InteractCore interface {
	GameIntractable
}

type CmdSender interface {
	SendWOCmd(cmd string)
	SendWebSocketCmdOmitResponse(cmd string)
	SendPlayerCmdOmitResponse(cmd string)
	SendAICommandOmitResponse(runtimeid string, cmd string)

	SendWebSocketCmdNeedResponse(cmd string) async_wrapper.AsyncResult[*packet.CommandOutput]
	SendPlayerCmdNeedResponse(cmd string) async_wrapper.AsyncResult[*packet.CommandOutput]
	SendAICommandNeedResponse(runtimeid string, cmd string) async_wrapper.AsyncResult[*packet.CommandOutput]
}

type InfoSender interface {
	BotSay(msg string)
	SayTo(target string, msg string)
	RawSayTo(target string, msg string)
	ActionBarTo(target string, msg string)
	TitleTo(target string, msg string)
	SubTitleTo(target string, subTitle string, title string)
}

type GameChat struct {
	// 玩家名（去除前缀, e.g. <乱七八糟的前缀> 张三 -> 张三）
	Name string
	// 分割后的消息 (e.g. "前往 1 2 3" ["前往", "1", "2", "3"])
	Msg []string
	// e.g. packet.TextTypeChat
	Type byte
	// 原始消息，未分割
	RawMsg string
	// 原始玩家名，未分割
	RawName string
	// 原始参数， packet.Text 中的 Parameters
	RawParameters []string
	// 附加信息，用于传递额外的信息，一般为空，也可能是 packet.Text 数据包，不要过于依赖这个字段
	Aux any
}

type PlayerMsgListener interface {
	GetInput(playerName string, breakOnLeave bool) async_wrapper.AsyncResult[*GameChat]
	SetOnChatCallBack(func(chat *GameChat))
	SetOnSpecificCommandBlockTellCallBack(commandBlockName string, cb func(chat *GameChat))
	SetOnSpecificItemMsgCallBack(itemName string, cb func(chat *GameChat))
}

type PlayerInteract interface {
	PlayerMsgListener
	ListAllPlayers() []PlayerKit
	ListenPlayerChange(cb func(player PlayerKit, action string))
	GetPlayerKit(name string) (playerKit PlayerKit, found bool)
	GetPlayerKitByUUID(ud uuid.UUID) (playerKit PlayerKit, found bool)
	GetPlayerKitByUUIDString(ud string) (playerKit PlayerKit, found bool)
	GetPlayerKitByUniqueID(uniqueID int64) (playerKit PlayerKit, found bool)
}

type QueryResult struct {
	Dimension int `json:"dimension"`
	Position  *struct {
		X float64 `json:"x"`
		Y float64 `json:"y"`
		Z float64 `json:"z"`
	} `json:"position"`
	UUID string  `json:"uniqueId"`
	YRot float64 `json:"yRot"`
}

type PlayerKit interface {
	PlayerUQReader
	Say(msg string)
	RawSay(msg string)
	ActionBar(msg string)
	Title(msg string)
	SubTitle(subTitle string, title string)
	GetInput(breakOnLeave bool) async_wrapper.AsyncResult[*GameChat]
	// e.g. CheckCondition(func(ok),"m=c","tag=op")
	CheckCondition(onResult func(bool), conditions ...string)
	Query(onResult func([]QueryResult), conditions ...string)
	// SetAbility(adventureFlagsUpdateMap, actionPermissionUpdateMap map[uint32]bool) (sent bool)
	// SetAbilityString(adventureFlagsUpdateMap, actionPermissionUpdateMap map[string]bool) (sent bool)

	SetBuildAbility(allow bool)
	SetMineAbility(allow bool)
	SetDoorsAndSwitchesAbility(allow bool)
	SetOpenContainersAbility(allow bool)
	SetAttackPlayersAbility(allow bool)
	SetAttackMobsAbility(allow bool)
	SetOperatorCommandsAbility(allow bool)
	SetTeleportAbility(allow bool)
}

type GameCtrl interface {
	InteractCore
	CmdSender
	InfoSender
	// BlockPlacer
}

type GameCtrlBox struct {
	GameCtrl
	*pressure_metric.FreqMetric
}

func (b GameCtrlBox) SendWebSocketCmdOmitResponse(cmd string) {
	b.GameCtrl.SendWebSocketCmdOmitResponse(cmd)
	b.FreqMetric.Record()
}

func (b GameCtrlBox) SendPlayerCmdOmitResponse(cmd string) {
	b.GameCtrl.SendPlayerCmdOmitResponse(cmd)
	b.FreqMetric.Record()
}

func (b GameCtrlBox) SendAICommandOmitResponse(runtimeid string, cmd string) {
	b.GameCtrl.SendAICommandOmitResponse(runtimeid, cmd)
	b.FreqMetric.Record()
}

func (b GameCtrlBox) SendWebSocketCmdNeedResponse(cmd string) async_wrapper.AsyncResult[*packet.CommandOutput] {
	b.FreqMetric.Record()
	return b.GameCtrl.SendWebSocketCmdNeedResponse(cmd)
}

func (b GameCtrlBox) SendPlayerCmdNeedResponse(cmd string) async_wrapper.AsyncResult[*packet.CommandOutput] {
	b.FreqMetric.Record()
	return b.GameCtrl.SendPlayerCmdNeedResponse(cmd)
}

func (b GameCtrlBox) SendAICommandNeedResponse(runtimeid string, cmd string) async_wrapper.AsyncResult[*packet.CommandOutput] {
	b.FreqMetric.Record()
	return b.GameCtrl.SendAICommandNeedResponse(runtimeid, cmd)
}

func NewGameCtrlBox(c GameCtrl, m *pressure_metric.FreqMetric) GameCtrlBox {
	return GameCtrlBox{
		GameCtrl:   c,
		FreqMetric: m,
	}
}

type MicroOmega interface {
	can_close.CanCloseWithError
	GetGameControl() GameCtrl
	GetReactCore() ReactCore
	GetGameListener() PacketDispatcher
	GetMicroUQHolder() MicroUQHolder
	GetPlayerInteract() PlayerInteract
	GetLowLevelAreaRequester() LowLevelAreaRequester
	GetBotAction() BotActionComplex
}

type MicroOmegaCmdBox struct {
	GameCtrlBox
	MicroOmega
}

func (b MicroOmegaCmdBox) GetGameControl() GameCtrl {
	return b.GameCtrlBox
}

func NewMicroOmegaCmdBox(o MicroOmega, m *pressure_metric.FreqMetric) MicroOmegaCmdBox {
	return MicroOmegaCmdBox{
		GameCtrlBox: NewGameCtrlBox(o.GetGameControl(), m),
		MicroOmega:  o,
	}
}

type UnReadyMicroOmega interface {
	MicroOmega
	// i dont wanto add this, but sometimes we need cmd to init stuffs
	NotifyChallengePassed()
	PostponeActionsAfterChallengePassed(name string, action func())
}
