package lang

import (
	"bufio"
	_ "embed"
	"fmt"
	"regexp"
	"strings"
	"text/template"
)

//go:embed zh_CN.lang
var langFile string

var langMap map[string]string

func convertconvertToTemplateFormat(input string) string {
	// find all format specifiers in the input string
	re := regexp.MustCompile(`%(\d*\$?\d*(\.\d+)?[a-zA-Z]|[1-9]\d*)|%%`)
	// replace all format specifiers with template format
	indexCounter := 1
	output := re.ReplaceAllStringFunc(input, func(match string) string {
		// if the match is "%%", return "%"
		if match == "%%" {
			return "%"
		}
		// exclude float
		if !regexp.MustCompile(`%\.\d*f`).MatchString(match) {
			// try to find the index in the format specifier
			indexMatches := regexp.MustCompile(`\d+`).FindStringSubmatch(match)
			// if there is an index, use it
			if len(indexMatches) > 0 {
				index := indexMatches[0]
				return fmt.Sprintf("{{.Arg%s}}", index)
			}
		}
		// otherwise, use the index from counter
		result := fmt.Sprintf("{{.Arg%d}}", indexCounter)
		indexCounter++
		return result
	})
	return output
}

func readLangFile() (map[string]string, error) {
	result := make(map[string]string)
	// read the lang file line by line
	scanner := bufio.NewScanner(strings.NewReader(langFile))
	for scanner.Scan() {
		line := scanner.Text()
		// remove comments
		parts := strings.SplitN(line, "#", 2)
		if len(parts) == 0 {
			continue
		}
		// split the line into key and value
		content := strings.Split(parts[0], "=")
		if len(content) != 2 || (content[0] == "" || content[1] == "") {
			continue
		}
		content[0] = strings.TrimSpace(content[0])
		content[1] = strings.TrimSpace(content[1])
		// add the key and value to the map
		result[content[0]] = convertconvertToTemplateFormat(content[1])
	}
	// check for errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	var err error
	langMap, err = readLangFile()
	if err != nil {
		panic(err)
	}
	// test the LangFormat function
	// fmt.Println(LangFormat("commands.event.error.failed", "dataA", "dataB"))
	// fmt.Println(LangFormat("commands.event.error.empty"))
	// fmt.Println(LangFormat("commands.execute.outRangedDetectPosition", 100, 200, 300, 400))
	// fmt.Println(LangFormat("commands.execute.trueCondition"))
	// fmt.Println(LangFormat("commands.fill.tooManyBlocks", 32129, 32768))
}

func LangFormat(key string, args ...any) string {
	// get the value from the map
	templateStr, ok := langMap[key]
	if !ok {
		return key
	}
	// create a map with the arguments
	data := make(map[string]any)
	for i, arg := range args {
		argName := fmt.Sprintf("Arg%d", i+1)
		data[argName] = arg
	}
	// parse the template
	tmpl, err := template.New("lang").Parse(templateStr)
	if err != nil {
		return key
	}
	// execute the template
	var result strings.Builder
	err = tmpl.Execute(&result, data)
	if err != nil {
		return key
	}
	return result.String()
}
