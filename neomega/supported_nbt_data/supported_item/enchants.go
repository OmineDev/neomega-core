package supported_item

import (
	"fmt"

	"github.com/OmineDev/neomega-core/i18n"
)

type Enchant int32

func (e Enchant) TranslatedString() string {
	return i18n.T_MCEnchantStr(int32(e))
}

type Enchants map[Enchant]int32

func (es Enchants) TranslatedString() string {
	out := ""
	if len(es) > 0 {
		for enchant, level := range es {
			out += fmt.Sprintf("[%v:%v级]", enchant.TranslatedString(), level)
		}
	}
	return out
}
