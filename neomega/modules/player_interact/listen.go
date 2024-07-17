package player_interact

import (
	"strings"

	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega"
	"github.com/OmineDev/neomega-core/neomega/uqholder"
	"github.com/OmineDev/neomega-core/utils/async_wrapper"
)

func (i *PlayerInteract) onTextPacket(pk *packet.Text) {
	if pk.Message == "" {
		return
	}
	if pk.SourceName == i.botBasicUQ.GetBotName() {
		return
	}
	splitMessage := strings.Split(pk.Message, " ")
	cleanedMessage := make([]string, 0, len(splitMessage))
	for _, v := range splitMessage {
		if v != "" {
			cleanedMessage = append(cleanedMessage, v)
		}
	}
	chat := &neomega.GameChat{
		Name:          uqholder.ToPlainName(pk.SourceName),
		Msg:           cleanedMessage,
		Type:          pk.TextType,
		RawMsg:        pk.Message,
		RawName:       pk.SourceName,
		RawParameters: pk.Parameters,
		Aux:           nil,
	}
	i.onChat(chat)
}

func (i *PlayerInteract) SetOnChatCallBack(cb func(chat *neomega.GameChat)) {
	i.chatCbs = append(i.chatCbs, cb)
}

func (i *PlayerInteract) SetOnSpecificCommandBlockTellCallBack(commandBlockName string, cb func(chat *neomega.GameChat)) {
	commandBlockName = strings.TrimSuffix(commandBlockName, "§r")
	commandBlockName += "§r"
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.commandBlockTellCbs[commandBlockName]; !ok {
		i.commandBlockTellCbs[commandBlockName] = make([]func(*neomega.GameChat), 0)
	}
	i.commandBlockTellCbs[commandBlockName] = append(i.commandBlockTellCbs[commandBlockName], cb)
}

func (i *PlayerInteract) SetOnSpecificItemMsgCallBack(itemName string, cb func(chat *neomega.GameChat)) {
	i.mu.Lock()
	defer i.mu.Unlock()
	if _, ok := i.specificItemMsgCbs[itemName]; !ok {
		i.specificItemMsgCbs[itemName] = make([]func(*neomega.GameChat), 0)
	}
	i.specificItemMsgCbs[itemName] = append(i.specificItemMsgCbs[itemName], cb)
}

func (i *PlayerInteract) GetInput(playerName string) async_wrapper.AsyncResult[*neomega.GameChat] {
	return async_wrapper.NewAsyncWrapper(func(ac *async_wrapper.AsyncController[*neomega.GameChat]) {
		var c chan *neomega.GameChat
		i.mu.Lock()
		found := false
		if c, found = i.nextMsgListenerChan[playerName]; !found {
			c = make(chan *neomega.GameChat)
			i.nextMsgListenerChan[playerName] = c
		}
		i.mu.Unlock()
		select {
		case chat := <-c:
			ac.SetResult(chat)
		case <-ac.Context().Done():
			return
		}
	}, true)
}

func (i *PlayerInteract) onChat(chat *neomega.GameChat) {
	i.mu.Lock()
	defer i.mu.Unlock()
	// specific item msg
	if cbs, ok := i.specificItemMsgCbs[chat.RawName]; ok {
		for _, cb := range cbs {
			go cb(chat)
		}
		return
	}
	_, isPlayer := i.playersUQ.GetPlayerByName(chat.Name)
	// command block tell
	if strings.HasSuffix(chat.RawName, "§r") && !isPlayer {
		if chat.Type == packet.TextTypeWhisper {
			if cbs, ok := i.commandBlockTellCbs[chat.RawName]; ok {
				for _, cb := range cbs {
					go cb(chat)
				}
			}
		}
		return
	}
	if ch, ok := i.nextMsgListenerChan[chat.Name]; ok {
		select {
		case ch <- chat:
			return
		default:
		}
	}
	for _, cb := range i.chatCbs {
		go cb(chat)
	}
}
