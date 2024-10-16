package blocks

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/OmineDev/neomega-core/neomega/blocks/describe"
)

func RuntimeIDToBlock(runtimeID uint32) (block *describe.Block, found bool) {
	Init()
	block = MC_CURRENT.BlockByRtid(runtimeID)
	return block, block != nil
}

func LegacyBlockToRuntimeID(name string, data uint16) (runtimeID uint32, found bool) {
	Init()
	return DefaultAnyToNemcConvertor.TryBestSearchByLegacyValue(describe.BlockNameForSearch(name), data)
}

func RuntimeIDToState(runtimeID uint32) (baseName string, properties map[string]any, found bool) {
	Init()
	block, found := RuntimeIDToBlock(runtimeID)
	if !found {
		return "air", nil, false
	}
	return block.ShortName(), block.States().ToNBT(), true
}

// coral_block ["coral_color":"yellow", "dead_bit":false] true
func RuntimeIDToBlockNameWithStateStr(runtimeID uint32) (blockNameWithState string, found bool) {
	Init()
	block, found := RuntimeIDToBlock(runtimeID)
	if !found {
		return "air []", false
	}
	return block.BedrockString(), true
}

func RuntimeIDToBlockNameAndStateStr(runtimeID uint32) (blockName, blockState string, found bool) {
	Init()
	block, found := RuntimeIDToBlock(runtimeID)
	if !found {
		return "air", "[]", false
	}
	return block.ShortName(), block.States().BedrockString(true), true
}

func BlockNameAndStateToRuntimeID(name string, properties map[string]any) (runtimeID uint32, found bool) {
	Init()
	props, err := describe.PropsForSearchFromNbt(properties)
	if err != nil {
		// legacy capability
		fmt.Println(err)
		return uint32(AIR_RUNTIMEID), false
	}
	rtid, _, found := DefaultAnyToNemcConvertor.TryBestSearchByState(describe.BlockNameForSearch(name), props)
	return rtid, found
}

func BlockNameAndStateStrToRuntimeID(name string, stateStr string) (runtimeID uint32, found bool) {
	Init()
	props, err := describe.PropsForSearchFromStr(stateStr)
	if err != nil {
		// legacy capability
		fmt.Println(err)
		return uint32(AIR_RUNTIMEID), false
	}
	rtid, _, found := DefaultAnyToNemcConvertor.TryBestSearchByState(describe.BlockNameForSearch(name), props)
	return rtid, found
}

func BlockStrToRuntimeID(blockNameWithOrWithoutState string) (runtimeID uint32, found bool) {
	Init()
	blockNameWithOrWithoutState = strings.TrimSpace(blockNameWithOrWithoutState)
	ss := strings.Split(blockNameWithOrWithoutState, " ")
	if len(ss) > 1 {
		cleanedSS := []string{}
		for _, s := range ss {
			if s == "" {
				continue
			}
			cleanedSS = append(cleanedSS, s)
		}
		if len(cleanedSS) == 2 {
			val, err := strconv.Atoi(cleanedSS[1])
			if err == nil {
				rtid, found := DefaultAnyToNemcConvertor.TryBestSearchByLegacyValue(describe.BlockNameForSearch(cleanedSS[0]), uint16(val))
				return rtid, found
			}
		}
	}
	blockName, blockProps := ConvertStringToBlockNameAndPropsForSearch(blockNameWithOrWithoutState)
	rtid, _, found := DefaultAnyToNemcConvertor.TryBestSearchByState(blockName, blockProps)
	return rtid, found
}

func SchemBlockStrToRuntimeID(blockNameWithOrWithoutState string) (runtimeID uint32, found bool) {
	Init()
	blockName, blockProps := ConvertStringToBlockNameAndPropsForSearch(blockNameWithOrWithoutState)
	rtid, _, found := SchemToNemcConvertor.TryBestSearchByState(blockName, blockProps)
	if found {
		return rtid, found
	} else {
		return BlockStrToRuntimeID(blockNameWithOrWithoutState)
	}
}

// func SchematicToRuntimeID(block uint8, value uint8) uint32 {
// 	value = value & 0xF
// 	return quickSchematicMapping[block][value]
// }

func ConvertStringToBlockNameAndPropsForSearch(blockString string) (blockNameForSearch describe.BaseWithNameSpace, propsForSearch *describe.PropsForSearch) {
	Init()
	blockString = strings.ReplaceAll(blockString, "{", "[")
	inFrags := strings.Split(blockString, "[")
	inBlockName, inBlockState := inFrags[0], ""
	if len(inFrags) > 1 {
		inBlockState = inFrags[1]
	}
	if len(inBlockState) > 0 {
		if inBlockState[len(inBlockState)-1] == ']' || inBlockState[len(inBlockState)-1] == '}' {
			inBlockState = inBlockState[:len(inBlockState)-1]
		}
	}
	inBlockStateForSearch, err := describe.PropsForSearchFromStr(inBlockState)
	if err != nil {
		// legacy capability
		fmt.Println(err)
	}
	return describe.BlockNameForSearch(inBlockName), inBlockStateForSearch
}
