package supported_item

import (
	"fmt"
	"strings"

	"github.com/OmineDev/neomega-core/neomega/blocks"
	"github.com/mitchellh/mapstructure"
)

func GenContainerItemsInfoFromItemsNbt(itemsNbt []any) (container map[uint8]*ContainerSlotItemStack, err error) {
	container = map[uint8]*ContainerSlotItemStack{}
	for _, itemNbt := range itemsNbt {
		item, ok := itemNbt.(map[string]any)
		if !ok {
			err = fmt.Errorf("fail to decode item nbt: %v", itemNbt)
			continue
		}
		itemNbt := &containerSlotItemNBT{}
		if err = mapstructure.Decode(item, &itemNbt); err != nil {
			err = fmt.Errorf("fail to decode item nbt: %v, %v", itemNbt, err)
			continue
		}
		slot, itemStack := itemNbt.toContainerSlotItemStack()
		if itemNbt.Tag.Damage != 0 {
			itemStack.Item.Value = itemNbt.Tag.Damage
		}
		container[slot] = itemStack
	}
	return container, err
}

// GlowItemFrame, ItemFrame ["item"]
func GenItemInfoFromItemFrameNBT(itemNbt any) (*Item, error) {
	item, ok := itemNbt.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("fail to decode item nbt: %v", itemNbt)
	}
	ir := &containerSlotItemNBT{}
	if err := mapstructure.Decode(item, &ir); err != nil {
		return nil, fmt.Errorf("fail to decode item nbt: %v, %v", itemNbt, err)
	}
	_, itemStack := ir.toContainerSlotItemStack()
	if ir.Tag.Damage != 0 {
		itemStack.Item.Value = ir.Tag.Damage
	}

	return itemStack.Item, nil
}

type containerSlotItemNBT struct {
	Slot       uint8
	Count      uint8
	Name       string
	Damage     int16 // if is a non-block item, use this as value
	CanDestroy []string
	CanPlaceOn []string
	Tag        struct {
		Display struct {
			Name string
		} `mapstructure:"display"`
		Enchant []struct {
			ID    int32 `mapstructure:"id"`
			Level int32 `mapstructure:"lvl"`
		} `mapstructure:"ench"`
		KeepOnDeath uint8 `mapstructure:"minecraft:keep_on_death"`
		ItemLock    uint8 `mapstructure:"minecraft:item_lock"`
		// if is chest/container
		Items []any
		// if is a book
		Pages []struct {
			Text string `mapstructure:"text"`
		} `mapstructure:"pages"`
		// if is written book
		Title  string `mapstructure:"title"`
		Author string `mapstructure:"author"`
		Damage int32  `mapstructure:"Damage"`
	} `mapstructure:"tag"`
	Block struct {
		Name  string         `mapstructure:"name"`
		State map[string]any `mapstructure:"states"`
		Value int16          `mapstructure:"val"` // if is a block item, use this as value
	} `mapstructure:"Block"`
}

func (i *containerSlotItemNBT) toContainerSlotItemStack() (slot uint8, stack *ContainerSlotItemStack) {
	slot = i.Slot
	stack = &ContainerSlotItemStack{
		Count: i.Count,
		Item: &Item{
			Name:                           i.Name,
			Value:                          int32(i.Damage),
			IsBlock:                        false,
			RelatedBlockBedrockStateString: "",
			BaseProps: &ItemPropsInGiveOrReplace{
				CanPlaceOn: i.CanPlaceOn,
				CanDestroy: i.CanDestroy,
				ItemLock: map[uint8]ItemLockPlace{
					1: ItemLockPlaceSlot,
					2: ItemLockPlaceInventory,
				}[i.Tag.ItemLock],
				KeepOnDeath: i.Tag.KeepOnDeath == 1,
			},
			Enchants:    make(Enchants),
			DisplayName: i.Tag.Display.Name,
		},
	}
	if stack.Item.BaseProps.IsEmpty() {
		stack.Item.BaseProps = nil
	}
	// if strings.Contains(stack.Item.Name, "box") {
	// 	fmt.Println("box")
	// }
	if i.Block.Name != "" {
		stack.Item.IsBlock = true
		stack.Item.Value = int32(i.Block.Value)
		states := map[string]any{}
		if len(i.Block.State) > 0 {
			states = i.Block.State
		}
		rtid, found := blocks.BlockNameAndStateToRuntimeID(i.Block.Name, states)
		if !found {
			fmt.Printf("unknown nested block: %v %v", i.Block.Name, states)
			stack.Item.IsBlock = false
		} else {
			blockNameNoStates, statesStr, _ := blocks.RuntimeIDToBlockNameAndStateStr(rtid)
			i.Block.Name = blockNameNoStates
			stack.Item.RelatedBlockBedrockStateString = statesStr
			stack.Item.Name = blockNameNoStates
		}
		// if len(i.Block.State) > 0 {
		// 	stack.Item.RelatedBlockBedrockStateString = "["
		// 	props := make([]string, 0)
		// 	for k, v := range i.Block.State {
		// 		if bv, ok := v.(uint8); ok {
		// 			if bv == 0 {
		// 				v = "false"
		// 			} else {
		// 				v = "true"
		// 			}
		// 		}
		// 		if sv, ok := v.(string); ok {
		// 			v = fmt.Sprintf("\"%v\"", sv)
		// 		}
		// 		props = append(props, fmt.Sprintf("\"%v\": %v", k, v))
		// 	}
		// 	stateStr := strings.Join(props, ", ")
		// 	stack.Item.RelatedBlockBedrockStateString = fmt.Sprintf("[%v]", stateStr)
		// }
	}
	for _, enchant := range i.Tag.Enchant {
		stack.Item.Enchants[Enchant(enchant.ID)] = enchant.Level
	}
	// a chest
	if len(i.Tag.Items) > 0 && stack.Item.IsBlock {
		if stack.Item.RelateComplexBlockData == nil {
			stack.Item.RelateComplexBlockData = &ComplexBlockData{}
		}
		stack.Item.RelateComplexBlockData.Container, _ = GenContainerItemsInfoFromItemsNbt(i.Tag.Items)
	}
	shortName := strings.TrimPrefix(i.Name, "minecraft:")
	// a book/written_book
	if shortName == "writable_book" || shortName == "written_book" {
		if stack.Item.SpecificKnownNonBlockItemData == nil {
			stack.Item.SpecificKnownNonBlockItemData = &SpecificKnownNonBlockItemData{}
		}
		stack.Item.SpecificKnownNonBlockItemData.Pages = make([]string, len(i.Tag.Pages))
		for i, data := range i.Tag.Pages {
			stack.Item.SpecificKnownNonBlockItemData.Pages[i] = data.Text
		}
	}

	if shortName == "written_book" {
		if stack.Item.SpecificKnownNonBlockItemData == nil {
			stack.Item.SpecificKnownNonBlockItemData = &SpecificKnownNonBlockItemData{}
		}
		stack.Item.SpecificKnownNonBlockItemData.BookAuthor = i.Tag.Author
		stack.Item.SpecificKnownNonBlockItemData.BookName = i.Tag.Title
	}
	return
}
