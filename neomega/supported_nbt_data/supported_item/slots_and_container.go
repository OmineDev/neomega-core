package supported_item

import (
	"fmt"
	"sort"
	"strings"
)

// to avoid using interface or "any" we enum all complex block supported here
type ComplexBlockData struct {
	// is a container
	Container map[uint8]*ContainerSlotItemStack `json:"container,omitempty"`
	// unknown (describe as a nbt)
	Unknown map[string]any `json:"unknown_nbt,omitempty"`
}

func (d *ComplexBlockData) String() string {
	if d == nil {
		return ""
	}
	if d.Container != nil {
		out := "容器内容:"
		origOrder := []int{}
		for slotID, _ := range d.Container {
			origOrder = append(origOrder, int(slotID))
		}
		sort.Ints(origOrder)
		for _, slotID := range origOrder {
			out += "\n  " + strings.ReplaceAll(fmt.Sprintf("槽%v: ", slotID)+d.Container[uint8(slotID)].String(), "\n", "\n  | ")
		}
		return out
	}
	return ""
}

type ContainerSlotItemStack struct {
	Item  *Item `json:"item"`
	Count uint8 `json:"count"`
}

func (s *ContainerSlotItemStack) String() string {
	if s == nil {
		return ""
	}
	out := fmt.Sprintf("%v个 %v", s.Count, s.Item.String())
	return out
}
