package cmd_sender

import (
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/nodes/defines"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"

	"github.com/google/uuid"
)

func init() {
	if false {
		func(sender neomega.CmdSender) {}(&EndPointCmdSender{})
	}
}

type EndPointCmdSender struct {
	*CmdSenderBasic
	node defines.APINode
}

func NewEndPointCmdSender(node defines.APINode, reactable neomega.ReactCore, interactable neomega.InteractCore) neomega.CmdSender {
	c := &EndPointCmdSender{
		CmdSenderBasic: NewCmdSenderBasic(reactable, interactable),
		node:           node,
	}
	return c
}

func (c *EndPointCmdSender) SendPlayerCmdNeedResponse(cmd string) *async_wrapper.AsyncWrapper[*packet.CommandOutput] {
	ud, _ := uuid.NewUUID()
	args := defines.FromString(cmd).Extend(defines.FromUUID(ud))
	return async_wrapper.NewAsyncWrapper(func(ac *async_wrapper.AsyncController[*packet.CommandOutput]) {
		c.cbByUUID.Set(ud.String(), func(co *packet.CommandOutput) {
			ac.SetResult(co)
		})
		ac.SetCancelHook(func() {
			c.cbByUUID.Delete(ud.String())
		})
		c.node.CallOmitResponse("send-player-command", args)
	}, false)
}

func (c *EndPointCmdSender) SendAICommandNeedResponse(runtimeid string, cmd string) *async_wrapper.AsyncWrapper[*packet.CommandOutput] {
	ud, _ := uuid.NewUUID()
	args := defines.FromString(runtimeid).Extend(defines.FromString(cmd), defines.FromUUID(ud))
	return async_wrapper.NewAsyncWrapper(func(ac *async_wrapper.AsyncController[*packet.CommandOutput]) {
		c.cbByUUID.Set(ud.String(), func(co *packet.CommandOutput) {
			ac.SetResult(co)
		})
		ac.SetCancelHook(func() {
			c.cbByUUID.Delete(ud.String())
		})
		c.node.CallOmitResponse("send-ai-command", args)
	}, false)
}
