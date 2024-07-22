package i18n

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strings"
)

//go:embed vanilla.json
var vanilla []byte

var Vanilla map[string]map[string]string

func init() {
	if err := json.Unmarshal(vanilla, &Vanilla); err != nil {
		panic(err)
	}
}

func T_MC(in string) (string, bool) {
	out, ok := T_MCItem(in)
	if ok {
		return out, true
	}
	for _, group := range Vanilla {
		if out, found := group[in]; found {
			return out, true
		}
	}
	out, ok = T_MCBlock(in)
	if ok {
		return out, true
	}

	return in, false
}

func T_MC_(in string) string {
	in, _ = T_MC(in)
	return in
}

func T_MCBlock(in string) (string, bool) {
	orig := in
	ss := strings.SplitN(in, "[", 2)
	in = ss[0]
	in = strings.TrimSpace(in)
	if out, found := Vanilla["block"][in]; found {
		if len(ss) > 1 {
			out += "[" + ss[1]
		}
		return out, true
	}
	in = strings.TrimPrefix(in, "minecraft:")
	if out, found := Vanilla["block"][in]; found {
		if len(ss) > 1 {
			out += "[" + ss[1]
		}
		return out, true
	}
	in = in + "_block"
	if out, found := Vanilla["block"][in]; found {
		if len(ss) > 1 {
			out += "[" + ss[1]
		}
		return out, true
	}
	return orig, false
}

func T_MCItem(in string) (string, bool) {
	if out, found := Vanilla["item"][in]; found {
		return out, true
	}
	in = strings.TrimPrefix(in, "minecraft:")
	in = strings.TrimSpace(in)
	if out, found := Vanilla["item"][in]; found {
		return out, true
	}
	return in, false
}

func T_MCEnchantStr(e int32) string {
	switch e {
	case 0:
		return "保护"
	case 1:
		return "火焰保护"
	case 2:
		return "摔落缓冲"
	case 3:
		return "爆炸保护"
	case 4:
		return "弹射物保护"
	case 5:
		return "荆棘"
	case 6:
		return "水下呼吸"
	case 7:
		return "深海探索者"
	case 8:
		return "水下速掘"
	case 9:
		return "锋利"
	case 10:
		return "亡灵杀手"
	case 11:
		return "节肢杀手"
	case 12:
		return "击退"
	case 13:
		return "火焰附加"
	case 14:
		return "抢夺"
	case 15:
		return "效率"
	case 16:
		return "精准采集"
	case 17:
		return "耐久"
	case 18:
		return "时运"
	case 19:
		return "力量"
	case 20:
		return "冲击"
	case 21:
		return "火矢"
	case 22:
		return "无限"
	case 23:
		return "海之眷顾"
	case 24:
		return "饵钓"
	case 25:
		return "冰霜行者"
	case 26:
		return "经验修补"
	case 27:
		return "绑定诅咒"
	case 28:
		return "消失诅咒"
	case 29:
		return "穿刺"
	case 30:
		return "激流"
	case 31:
		return "忠诚"
	case 32:
		return "引雷"
	case 33:
		return "多重射击"
	case 34:
		return "穿透"
	case 35:
		return "快速装填"
	case 36:
		return "灵魂疾行"
	case 37:
		return "迅捷潜行"
	default:
		return fmt.Sprintf("魔咒: %v", e)
	}
}
