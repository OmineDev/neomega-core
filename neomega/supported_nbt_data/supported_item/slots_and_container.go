package supported_item

import (
	"fmt"
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
		for _, slot := range d.Container {
			out += "\n  " + strings.ReplaceAll(slot.String(), "\n", "\n  | ")
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
