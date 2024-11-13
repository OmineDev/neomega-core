package rawtext_wrapper

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

type RawTextItem struct {
	Text      string   `json:"text,omitempty"`
	Translate string   `json:"translate,omitempty"`
	With      *RawText `json:"with,omitempty"`
}

type RawText struct {
	RawTextItemList []RawTextItem `json:"rawtext"`
}

func ParseGameRawText(rawtext string) string {
	var v RawText
	// 解析失败, 返回原消息
	if json.Unmarshal([]byte(rawtext), &v) != nil {
		return rawtext
	}
	// 解析并返回
	return strings.Join(RawTextStruct2StringSlice(&v), "")
}

func RawTextStruct2StringSlice(rawtext *RawText) (result []string) {
	for _, rawTextItem := range rawtext.RawTextItemList {
		k := reflect.TypeOf(rawTextItem)
		v := reflect.ValueOf(rawTextItem)
		for i := 0; i < k.NumField(); i++ {
			str := v.Field(i).String()
			if str == "" {
				continue
			}
			switch k.Field(i).Name {
			case "Text":
				result = append(result, str)
			case "Translate":
				result = append(result, str)
			case "With":
				if rawtext, ok := v.Field(i).Interface().(*RawText); ok && rawtext != nil {
					result = append(
						result,
						fmt.Sprintf(" (%v)", strings.Join(RawTextStruct2StringSlice(rawtext), ", ")),
					)
				}
			default:
				result = append(result, str)
			}
		}
	}
	return result
}
