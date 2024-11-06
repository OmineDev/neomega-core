package bundle

import (
	"fmt"

	"github.com/OmineDev/neomega-core/i18n"
	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/nodes/defines"

	// "github.com/OmineDev/neomega-core/neomega/modules/block/placer"
	"sync"
	"time"

	"github.com/OmineDev/neomega-core/neomega/modules/area_request"
	"github.com/OmineDev/neomega-core/neomega/modules/bot_action"
	"github.com/OmineDev/neomega-core/neomega/modules/info_sender"
	"github.com/OmineDev/neomega-core/neomega/modules/player_interact"
)

func init() {
	if false {
		func(omega neomega.MicroOmega) {}(&MicroOmega{})
	}
}

type MicroOmega struct {
	neomega.ReactCore
	neomega.InteractCore
	neomega.InfoSender
	neomega.CmdSender
	neomega.MicroUQHolder
	// neomega.BlockPlacer
	neomega.PlayerInteract
	neomega.LowLevelAreaRequester
	neomega.CommandHelper
	neomega.BotAction
	neomega.BotActionHighLevel
	deferredActions []struct {
		cb   func()
		name string
	}
	mu sync.Mutex
}

func NewMicroOmega(
	interactCore neomega.InteractCore,
	reactCore neomega.UnStartedReactCore,
	microUQHolder neomega.MicroUQHolder,
	cmdSender neomega.CmdSender,
	node defines.Node,
	isAccessPoint bool,
) neomega.UnReadyMicroOmega {
	infoSender := info_sender.NewInfoSender(interactCore, cmdSender, microUQHolder.GetBotBasicInfo())
	playerInteract := player_interact.NewPlayerInteract(reactCore, microUQHolder.GetPlayersInfo(), microUQHolder.GetBotBasicInfo(), cmdSender, infoSender, interactCore)
	// asyncNbtBlockPlacer := placer.NewAsyncNbtBlockPlacer(reactCore, cmdSender, interactCore)
	areaRequester := area_request.NewAreaRequester(interactCore, reactCore, microUQHolder, microUQHolder)
	cmdHelper := bot_action.NewCommandHelper(cmdSender, microUQHolder)
	var botAction neomega.BotAction
	if isAccessPoint {
		botAction = bot_action.NewAccessPointBotActionWithPersistData(microUQHolder, interactCore, reactCore, cmdSender, node)
	} else {
		botAction = bot_action.NewEndPointBotAction(node, microUQHolder, interactCore)
	}

	botActionHighLevel := bot_action.NewBotActionHighLevel(microUQHolder, interactCore, reactCore, cmdSender, cmdHelper, areaRequester, botAction, node)

	omega := &MicroOmega{
		reactCore,
		interactCore,
		infoSender,
		cmdSender,
		microUQHolder,
		// asyncNbtBlockPlacer,
		playerInteract,
		areaRequester,
		cmdHelper,
		botAction,
		botActionHighLevel,
		make([]struct {
			cb   func()
			name string
		}, 0),
		sync.Mutex{},
	}

	if isAccessPoint {
		omega.PostponeActionsAfterChallengePassed("request tick update schedule", func() {
			go func() {
				for {
					clientTick := 0
					if tick, found := omega.GetMicroUQHolder().GetExtendInfo().GetCurrentTick(); found {
						clientTick = int(tick)
					}
					omega.GetGameControl().SendPacket(&packet.TickSync{
						ClientRequestTimestamp: int64(clientTick),
					})
					time.Sleep(time.Second * 5)
				}
			}()
		})
		omega.PostponeActionsAfterChallengePassed("auto respawn", func() {
			omega.GetGameListener().SetTypedPacketCallBack(packet.IDRespawn, func(p packet.Packet) {
				pkt := p.(*packet.Respawn)
				if pkt.State == packet.RespawnStateSearchingForSpawn {
					omega.SendPacket(&packet.Respawn{
						State:           packet.RespawnStateClientReadyToSpawn,
						EntityRuntimeID: omega.GetBotRuntimeID(),
					})
					omega.SendPacket(&packet.PlayerAction{
						EntityRuntimeID: omega.GetBotRuntimeID(),
						ActionType:      protocol.PlayerActionRespawn,
						BlockFace:       -1,
					})
				}
			}, true)
		})
		omega.PostponeActionsAfterChallengePassed("dial tick every 1/20 second", func() {
			go func() {
				startTime := time.Now()
				tickAdd := int64(0)
				for {
					// sleep in some platform (yes, you, windows!) is not very accurate
					tickToAdd := (time.Since(startTime).Milliseconds() / 50) - tickAdd
					if tickToAdd > 0 {
						tickAdd += tickToAdd
						if tick, found := omega.GetMicroUQHolder().GetExtendInfo().GetCurrentTick(); found {
							omega.GetMicroUQHolder().GetExtendInfo().UpdateFromPacket(&packet.TickSync{
								ClientRequestTimestamp:   0,
								ServerReceptionTimestamp: tick + tickToAdd,
							})
						}
					}
					time.Sleep(time.Second / 20)
				}
			}()
		})
		omega.PostponeActionsAfterChallengePassed("force reset dimension and pos", func() {
			e := &neomega.PosAndDimensionInfo{}
			if bot_action.RefreshPosAndDimensionInfo(e, omega) == nil {
				// fmt.Println(e)
				omega.MicroUQHolder.UpdateFromPacket(&packet.ChangeDimension{
					Dimension: int32(e.Dimension),
					Position:  e.HeadPosPrecise,
				})
			}
		})
	}

	if !isAccessPoint {
		omega.PostponeActionsAfterChallengePassed("check bot command status each 10s", func() {
			go func() {
				for {
					ret, err := omega.SendWebSocketCmdNeedResponse("errcmd").SetTimeout(time.Minute).BlockGetResult()
					if err != nil || ret == nil {
						panic("for some reason, end point cannot communicate with server, reload")
					} else {
						// fmt.Println(ret)
					}
					time.Sleep(time.Second * 10)
				}
			}()
		})
	}

	reactCore.Start()
	return omega
}

func (o *MicroOmega) GetGameControl() neomega.GameCtrl {
	return o
}

func (o *MicroOmega) GetReactCore() neomega.ReactCore {
	return o
}

func (o *MicroOmega) GetGameListener() neomega.PacketDispatcher {
	return o
}

func (o *MicroOmega) GetPlayerInteract() neomega.PlayerInteract {
	return o
}

func (o *MicroOmega) GetMicroUQHolder() neomega.MicroUQHolder {
	return o
}

func (o *MicroOmega) GetLowLevelAreaRequester() neomega.LowLevelAreaRequester {
	return o
}

func (o *MicroOmega) GetBotAction() neomega.BotActionComplex {
	return o
}

func (o *MicroOmega) NotifyChallengePassed() {
	for _, action := range o.deferredActions {
		fmt.Printf(i18n.T(i18n.S_starting_post_challenge_actions), action.name)
		action.cb()
	}
}

func (o *MicroOmega) PostponeActionsAfterChallengePassed(name string, action func()) {
	o.mu.Lock()
	defer o.mu.Unlock()
	o.deferredActions = append(o.deferredActions, struct {
		cb   func()
		name string
	}{action, name})
}
