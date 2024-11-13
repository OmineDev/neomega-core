package lang

import (
	"strings"
	"text/template"
)

func execTemplate(templates map[string]*template.Template, key string, data []any) (string, bool) {
	// define a function to format the string
	formatFunc := func(str string) (string, bool) {
		if tmpl, ok := templates[str]; ok {
			var formatedArg strings.Builder
			err := tmpl.Execute(&formatedArg, data)
			return formatedArg.String(), err == nil
		}
		return str, false
	}
	// format by key
	parts := strings.SplitN(key, "%", 2)
	switch len(parts) {
	case 1:
		return formatFunc(parts[0])
	case 2:
		formatted, ok := formatFunc(parts[1])
		return parts[0] + formatted, ok
	default:
		return key, false
	}
}

func LangFormat(lang uint, key string, args ...any) (string, bool) {
	// get the value from the map
	templates, ok := langTemplates[lang]
	if !ok {
		return key, false
	}
	// trim and format args
	for i, arg := range args {
		if str, ok := arg.(string); ok {
			if formatted, ok := execTemplate(templates, str, nil); ok {
				args[i] = formatted
			}
		}
	}
	// trim and format key
	return execTemplate(templates, key, args)
}
