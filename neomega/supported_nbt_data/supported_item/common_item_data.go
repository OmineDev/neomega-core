package supported_item

import (
	"encoding/json"
	"fmt"

	"github.com/OmineDev/neomega-core/i18n"
)

type ItemLockPlace string

const (
	ItemLockPlaceInventory = ItemLockPlace("lock_in_inventory") //2
	ItemLockPlaceSlot      = ItemLockPlace("lock_in_slot")      //1
)

type ItemPropsInGiveOrReplace struct {
	CanPlaceOn  []string      `json:"can_place_on,omitempty"`
	CanDestroy  []string      `json:"can_destroy,omitempty"`
	ItemLock    ItemLockPlace `json:"item_lock,omitempty"`
	KeepOnDeath bool          `json:"keep_on_death,omitempty"`
}

func (c *ItemPropsInGiveOrReplace) IsEmpty() bool {
	return (len(c.CanPlaceOn) == 0) && (len(c.CanDestroy) == 0) && (c.ItemLock == "") && (!c.KeepOnDeath)
}

func (c *ItemPropsInGiveOrReplace) String() string {
	if c == nil {
		return ""
	}
	out := ""
	if c != nil {
		if len(c.CanPlaceOn) > 1 {
			out += "可被放置于: "
			for _, name := range c.CanPlaceOn {
				out += " " + i18n.T_MC_(name)
			}
		}
		if len(c.CanDestroy) > 1 {
			out += "\n可破坏: "
			for _, name := range c.CanDestroy {
				out += " " + name
			}
		}
		if c.ItemLock != "" || c.KeepOnDeath {
			if out != "" {
				out += "\n"
			}
			if c.ItemLock != "" {
				if c.ItemLock == ItemLockPlaceSlot {
					out += "锁定在物品栏"
				} else if c.ItemLock == ItemLockPlaceInventory {
					out += "锁定在背包"
				} else {
					out += fmt.Sprintf("锁定在: %v", c.ItemLock)
				}
			}
			if c.KeepOnDeath {
				if c.ItemLock != "" {
					out += "&"
				}
				out += "死亡时保留"
			}
		}
	}
	return out
}

func (c *ItemPropsInGiveOrReplace) CmdString() string {
	if c == nil {
		return ""
	}
	components := map[string]any{}

	if len(c.CanPlaceOn) > 0 {
		components["minecraft:can_place_on"] = map[string][]string{"blocks": c.CanPlaceOn}
	}
	if len(c.CanDestroy) > 0 {
		components["minecraft:can_destroy"] = map[string][]string{"blocks": c.CanDestroy}
	}
	if c.ItemLock != "" {
		components["item_lock"] = map[string]string{"mode": string(c.ItemLock)}
	}
	if c.KeepOnDeath {
		components["keep_on_death"] = map[string]string{}
	}
	if len(components) == 0 {
		return ""
	}
	// will not fail
	bs, _ := json.Marshal(components)
	return string(bs)
}
