package lang

import (
	"strings"
)

func LangFormat(lang uint, key string, args ...any) string {
	// get the value from the map
	templates, ok := langTemplates[lang]
	if !ok {
		return key
	}
	tmpl, ok := templates[key]
	if !ok {
		return key
	}
	// execute the template
	var result strings.Builder
	if err := tmpl.Execute(&result, args); err != nil {
		return key
	}
	return result.String()
}
