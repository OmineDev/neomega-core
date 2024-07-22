package supported_item

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/OmineDev/neomega-core/i18n"
)

type Item struct {
	// Basic Item: item can be given by replace/give, etc.
	Name  string `json:"name"`
	Value int16  `json:"val"`
	// can place on, can break, lock, etc. safe to be nil
	BaseProps *ItemPropsInGiveOrReplace `json:"base_props,omitempty"`
	// if item can be put as a block, it could have RelatedBlockStateString, safe to be empty
	IsBlock                        bool   `json:"is_block"`
	RelatedBlockBedrockStateString string `json:"block_bedrock_state_string,omitempty"`
	// END Basic Item

	// KnownItem
	SpecificKnownNonBlockItemData *SpecificKnownNonBlockItemData `json:"specific_known_item_data,omitempty"`
	// End Known Item

	// NeedHotBarItem: item that requires putted in hot bar when generating, usually enchant, safe to be empty
	Enchants Enchants `json:"enchants,omitempty"`
	// End NeedHotBarItem

	// NeedAuxItem: item need aux block when generating, e.g. rename, safe to be empty
	DisplayName string `json:"display_name,omitempty"`
	// End NeedAuxItem

	// Complex Block: when item put as a block, the block contains certain info, safe to be empty
	RelateComplexBlockData *ComplexBlockData `json:"complex_block_data,omitempty"`
	// End Complex Block
}

func (i *Item) String() string {
	if i == nil {
		return ""
	}
	out := fmt.Sprintf("%v[特殊值=%v]", i18n.T_MC_(i.Name), i.Value)
	if i.IsBlock {
		out += " 物品方块属性:" + i.RelatedBlockBedrockStateString
	}
	common := i.BaseProps.String()
	if common != "" {
		out += "\n -基础属性:\n  " + strings.ReplaceAll(common, "\n", "\n  ")
	}
	specific := i.SpecificKnownNonBlockItemData.String()
	if specific != "" {
		out += "\n -信息:" + strings.ReplaceAll(specific, "\n", "\n  ")
	}
	enchants := i.Enchants.TranslatedString()
	if enchants != "" {
		out += "\n -附魔:" + strings.ReplaceAll(enchants, "\n", "\n  ")
	}
	if i.DisplayName != "" {
		out += "\n -被命名为: " + strings.ReplaceAll(i.DisplayName, "\n", "\n  ")
	}
	if i.RelateComplexBlockData != nil && i.RelateComplexBlockData.Container != nil {
		contained := i.RelateComplexBlockData.String()
		out += "\n -包含子容器: \n  " + strings.ReplaceAll(contained, "\n", "\n  ")
	}
	return out
}

func (i *Item) GetJsonString() string {
	bs, _ := json.Marshal(i)
	return string(bs)
}

func (i *Item) GetTypeDescription() ItemTypeDescription {
	if i.IsBlock {
		if i.RelateComplexBlockData != nil {
			if i.RelateComplexBlockData.Container != nil {
				return ComplexBlockItemContainer
			}
			return ComplexBlockItemUnknown
		}
		if i.DisplayName != "" {
			return NeedAuxBlockBlockItem
		} else {
			if len(i.Enchants) > 0 {
				return NeedHotBarBlockItem
			} else {
				return SimpleBlockItem
			}
		}
	} else {
		knownItem := ""
		shortName := strings.TrimPrefix(i.Name, "minecraft:")
		knownItem = map[string]string{
			"written_book":  ": " + KnownItemWrittenBook,
			"writable_book": ": " + KnownItemWritableBook,
		}[shortName]
		if i.DisplayName != "" {
			return NeedAuxBlockNonBlockItem + ItemTypeDescription(knownItem)
		} else {
			if len(i.Enchants) > 0 || knownItem != "" {
				return NeedHotBarNonBlockItem + ItemTypeDescription(knownItem)
			} else {
				return SimpleNonBlockItem + ItemTypeDescription(knownItem)
			}
		}
	}
}
