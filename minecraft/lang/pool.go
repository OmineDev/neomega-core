package lang

import (
	"bufio"
	"bytes"
	_ "embed"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"text/template"

	"github.com/andybalholm/brotli"
)

const (
	LANG_ZH_CN = iota
)

//go:embed zh_CN.lang.brotli
var zh_CN []byte

var (
	formatSpecifierRegex = regexp.MustCompile(`%(\d+\$)?[-+#0 ]?(\d+|\*)?(\.\d+|\.\*)?[hlL]?[a-zA-Z%]`)
	excludeFloatRegex    = regexp.MustCompile(`%\.\d*f`)
	indexRegex           = regexp.MustCompile(`\d+`)
	langTemplates        map[uint]map[string]*template.Template // map[langCode][key]template
)

func convertToTemplate(input string) (*template.Template, error) {
	// replace all format specifiers with template format
	indexCounter := 0
	output := formatSpecifierRegex.ReplaceAllStringFunc(input, func(match string) string {
		// if the match is "%%", return "%"
		if match == "%%" {
			return "%"
		}
		// exclude float
		if !excludeFloatRegex.MatchString(match) {
			// try to find the index in the format specifier
			indexMatches := indexRegex.FindStringSubmatch(match)
			// if there is an index, use it
			if len(indexMatches) > 0 {
				index, _ := strconv.Atoi(indexMatches[0])
				return fmt.Sprintf("{{index . %d}}", index-1)
			}
		}
		// otherwise, use the index from counter
		result := fmt.Sprintf("{{index . %d}}", indexCounter)
		indexCounter++
		return result
	})
	// parse and return template
	return template.New("").Parse(output)
}

func readLangFile(content []byte) (result map[string]*template.Template, err error) {
	result = make(map[string]*template.Template)
	// read the lang file line by line
	brotliReader := brotli.NewReader(bytes.NewReader(content))
	scanner := bufio.NewScanner(brotliReader)
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
		result[content[0]], err = convertToTemplate(content[1])
		if err != nil {
			return nil, err
		}
	}
	// check for errors
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func init() {
	// load lang files
	langFiles := map[uint][]byte{
		LANG_ZH_CN: zh_CN,
	}
	// parse lang files
	langTemplates = make(map[uint]map[string]*template.Template)
	for lang, content := range langFiles {
		langMap, err := readLangFile(content)
		if err != nil {
			panic(err)
		}
		langTemplates[lang] = langMap
	}
	// test the LangFormat function
	// fmt.Println(LangFormat(LANG_ZH_CN, "commands.event.error.failed", "dataA", "dataB"))
	// fmt.Println(LangFormat(LANG_ZH_CN, "netease.report.kick.hint"))
	// fmt.Println(LangFormat(LANG_ZH_CN, "commands.execute.outRangedDetectPosition", 100, 200, 300, 400))
	// fmt.Println(LangFormat(LANG_ZH_CN, "commands.generic.syntax", "/summon", "errorStr", "Steve"))
	// fmt.Println(LangFormat(LANG_ZH_CN, "gameMode.changed", LangFormat(LANG_ZH_CN, "createWorldScreen.gameMode.creative")))
	// fmt.Println(LangFormat(LANG_ZH_CN, "commands.generic.double.tooSmall", 2.1, 4.5))
}
