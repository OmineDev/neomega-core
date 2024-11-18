package lang

import (
	"fmt"
	"strings"
)

func LangFormat(lang uint, key string, args ...any) string {
	// get root node by lang code
	node, ok := langNodes[lang]
	if !ok {
		return key
	}
	// trim and format args
	for i, arg := range args {
		if formatted, ok := execTemplateNode(node, fmt.Sprintf("%v", arg), nil); ok {
			args[i] = formatted
		}
	}
	// split key into parts
	parts := strings.Split(key, "%")
	// iterate to format each part
	for i, part := range parts {
		if formatted, ok := execTemplateNode(node, part, args); ok {
			parts[i] = formatted
		} else if i > 0 {
			parts[i] = "%" + part
		}
	}
	return strings.Join(parts, "")
}
