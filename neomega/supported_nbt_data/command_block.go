package supported_nbt_data

import (
	"fmt"
	"strings"

	"github.com/OmineDev/neomega-core/minecraft/protocol"
	"github.com/OmineDev/neomega-core/minecraft/protocol/packet"
	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/OmineDev/neomega-core/neomega/chunks/define"
)

type CommandBlockSupportedData struct {
	BlockName          string
	BlockState         string
	NeedRedStone       bool
	Conditional        bool
	Command            string
	Name               string
	TickDelay          int
	ShouldTrackOutput  bool
	ExecuteOnFirstTick bool
}

type CommandBlockSupportedDataWithPos struct {
	*CommandBlockSupportedData
	Pos define.CubePos
}

func (opt *CommandBlockSupportedDataWithPos) String() string {
	if opt == nil {
		return ""
	}
	return fmt.Sprintf("[%v,%v,%v]", opt.Pos.X(), opt.Pos.Y(), opt.Pos.Z()) + opt.CommandBlockSupportedData.String()
}

func (opt *CommandBlockSupportedData) String() string {
	if opt == nil {
		return ""
	}
	describe := ""
	if opt.Name != "" {
		describe += fmt.Sprintf("\n  名字: %v", opt.Name)
	}
	if opt.Command != "" {
		describe += fmt.Sprintf("\n  指令: %v", opt.Command)
	}
	options := fmt.Sprintf("\n  红石=%v,有条件=%v,显示输出=%v,执行第一个已选项=%v,延迟=%v", opt.NeedRedStone, opt.Conditional, opt.ShouldTrackOutput, opt.ExecuteOnFirstTick, opt.TickDelay)
	describe += strings.ReplaceAll(strings.ReplaceAll(options, "true", "是"), "false", "否")
	return describe
}

func (opt *CommandBlockSupportedData) GenCommandBlockUpdateFromOption(targetPos define.CubePos) *packet.CommandBlockUpdate {
	var mode uint32
	if opt.BlockName == "command_block" {
		mode = packet.CommandBlockImpulse
	} else if opt.BlockName == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	} else if opt.BlockName == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else {
		opt.BlockName = "command_block"
		mode = packet.CommandBlockImpulse
	}
	return &packet.CommandBlockUpdate{
		Block:              true,
		Position:           protocol.BlockPos{int32(targetPos.X()), int32(targetPos.Y()), int32(targetPos.Z())},
		Mode:               mode,
		NeedsRedstone:      opt.NeedRedStone,
		Conditional:        opt.Conditional,
		Command:            opt.Command,
		LastOutput:         "",
		Name:               opt.Name,
		TickDelay:          uint32(opt.TickDelay),
		ExecuteOnFirstTick: opt.ExecuteOnFirstTick,
		ShouldTrackOutput:  opt.ShouldTrackOutput,
	}
}

func NewCommandBlockSupportedDataFromNBT(blockNameAndState string, nbt map[string]interface{}) (o *CommandBlockSupportedData, err error) {
	rtid, found := blocks.BlockStrToRuntimeID(blockNameAndState)
	if !found {
		return nil, fmt.Errorf("cannot recognize this block %v", blockNameAndState)
	}
	return NewCommandBlockSupportedDataFromNBTAndRtid(rtid, nbt)
}

func NewCommandBlockSupportedDataFromNBTAndRtid(rtid uint32, nbt map[string]interface{}) (o *CommandBlockSupportedData, err error) {
	_, exist := nbt["__tag"]
	if exist {
		return nil, fmt.Errorf("flatten nemc nbt, cannot handle")
	}
	if nbt == nil {
		return nil, fmt.Errorf("nbt is empty, cannot handle")
	}
	var mode uint32
	defer func() {
		r := recover()
		if r != nil {
			err = fmt.Errorf("cannot gen place command block option %v", r)
		}
	}()
	block, _ := blocks.RuntimeIDToBlock(rtid)
	if block.ShortName() == "command_block" {
		mode = packet.CommandBlockImpulse
	} else if block.ShortName() == "repeating_command_block" {
		mode = packet.CommandBlockRepeating
	} else if block.ShortName() == "chain_command_block" {
		mode = packet.CommandBlockChain
	} else {
		return nil, fmt.Errorf("this block %v%v is not command block", block.ShortName(), block.States().BedrockString(true))
	}
	mode = mode // just make compiler happy
	cmd, _ := nbt["Command"].(string)
	constumeName, _ := nbt["CustomName"].(string)
	exeft, _ := nbt["ExecuteOnFirstTick"].(uint8)
	tickdelay, _ := nbt["TickDelay"].(int32)     //*/
	aut, _ := nbt["auto"].(uint8)                //!needrestone
	trackoutput, _ := nbt["TrackOutput"].(uint8) //
	var conditionalmode uint8
	// lo, _ := nbt["LastOutput"].(string)
	// conditionalmode, ok := nbt["conditionalMode"].(uint8)
	// if !ok {
	// 	conditionalmode = block.States().ToNBT()["conditional_bit"].(uint8)
	// }
	for _, p := range block.States() {
		if p.Name == "conditional_bit" {
			conditionalmode = p.Value.Uint8Val()
		}
	}
	//conditionalmode := nbt["conditionalMode"].(uint8)
	var executeOnFirstTickBit bool
	if exeft == 0 {
		executeOnFirstTickBit = false
	} else {
		executeOnFirstTickBit = true
	}
	var trackOutputBit bool
	if trackoutput == 1 {
		trackOutputBit = true
	} else {
		trackOutputBit = false
	}
	var needRedStoneBit bool
	if aut == 1 {
		needRedStoneBit = false
		//REVERSED!!
	} else {
		needRedStoneBit = true
	}
	var conditionalBit bool
	if conditionalmode == 1 {
		conditionalBit = true
	} else {
		conditionalBit = false
	}
	o = &CommandBlockSupportedData{
		BlockName:          block.ShortName(),
		BlockState:         block.States().BedrockString(true),
		NeedRedStone:       needRedStoneBit,
		Conditional:        conditionalBit,
		Command:            cmd,
		Name:               constumeName,
		TickDelay:          int(tickdelay),
		ShouldTrackOutput:  trackOutputBit,
		ExecuteOnFirstTick: executeOnFirstTickBit,
	}
	return o, nil
}
