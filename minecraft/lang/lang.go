package lang

import (
	"fmt"
	"strings"
	"text/template"
)

func execTemplate(templates map[string]*template.Template, key string, data []any) (result string, ok bool) {
	parts := strings.Split(key, "%")
	tmplNames := make([]string, len(parts))
	// iterate template names
	for name := range templates {
		// iterate parts to get the longest match
		for index, part := range parts {
			if strings.HasPrefix(part, name) && len(name) > len(tmplNames[index]) {
				tmplNames[index] = name
			}
		}
	}
	// execute templates
	for index, tmplName := range tmplNames {
		// no match
		if tmplName == "" {
			continue
		}
		// execute the template
		var formatted strings.Builder
		if err := templates[tmplName].Execute(&formatted, data); err != nil {
			continue
		}
		parts[index] = strings.Replace(parts[index], tmplName, formatted.String(), 1)
	}
	return strings.Join(parts, ""), true
}

func LangFormat(lang uint, key string, args ...any) (string, bool) {
	// get the value from the map
	templates, ok := langTemplates[lang]
	if !ok {
		return key, false
	}
	// trim and format args
	for i, arg := range args {
		if formatted, ok := execTemplate(templates, fmt.Sprintf("%v", arg), nil); ok {
			args[i] = formatted
		}
	}
	// trim and format key
	return execTemplate(templates, key, args)
}
