package lang

import (
	"encoding/json"
	"fmt"
)

type RawTextItem struct {
	Text      *string  `json:"text,omitempty"`
	Translate *string  `json:"translate,omitempty"`
	With      *RawText `json:"with,omitempty"`
}

type RawText struct {
	RawTextItemList []RawTextItem `json:"rawtext"`
}

func ParseGameRawText(str string) (result string) {
	rawText := &RawText{}
	if err := json.Unmarshal([]byte(str), rawText); err != nil {
		return str
	}
	for _, item := range parseGameRawText(rawText) {
		result += fmt.Sprintf("%v", item)
	}
	return
}

func parseGameRawText(rawText *RawText) (result []any) {
	for _, item := range rawText.RawTextItemList {
		if item.Text != nil {
			// text
			result = append(result, *item.Text)
			continue
		}
		if item.Translate != nil {
			// with
			args := []any{}
			if item.With != nil {
				args = parseGameRawText(item.With)
			}
			// translate
			result = append(result, LangFormat(LANG_ZH_CN, *item.Translate, args...))
		}
	}
	return
}
