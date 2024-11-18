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
	langNodes            map[uint]*langNode // map[langCode][NodeTree]
)

type langNode struct {
	template *template.Template
	subNodes map[string]*langNode
}

func addTemplateNode(node *langNode, key string, template *template.Template) {
	// split the key into parts
	nodeNames := strings.SplitN(key, ".", 2)
	// create sub node
	subnode, ok := node.subNodes[nodeNames[0]]
	if !ok {
		subnode = &langNode{
			subNodes: make(map[string]*langNode),
		}
		node.subNodes[nodeNames[0]] = subnode
	}
	// add the template to the sub node
	if len(nodeNames) == 1 {
		subnode.template = template
	} else {
		addTemplateNode(subnode, nodeNames[1], template)
	}
}

func execTemplateNode(node *langNode, str string, value []any) (string, bool) {
	strNode := strings.SplitN(str, ".", 2)[0]
	// get the longest match of sub node
	var subNode *langNode
	var subNodeName string
	for name, node := range node.subNodes {
		if strings.HasPrefix(strNode, name) && (len(name) > len(subNodeName)) {
			subNode = node
			subNodeName = name
		}
		// if full match, break
		if len(strNode) == len(subNodeName) {
			break
		}
	}
	// cut off matched string (node name)
	str = strings.TrimPrefix(str, subNodeName)
	// return if no matches sub node or the first char is not "." (no next)
	if subNode == nil {
		if node.template == nil {
			return "", false
		}
		var formatted strings.Builder
		if err := node.template.Execute(&formatted, value); err != nil {
			return "", false
		}
		return formatted.String() + str, true
	}
	// if sub node exists and the first char is "."
	return execTemplateNode(subNode, strings.TrimPrefix(str, "."), value)
}

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

func readLangFile(node *langNode, content []byte) error {
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
		// convert the value to template
		template, err := convertToTemplate(content[1])
		if err != nil {
			continue
		}
		// add the template to the node tree
		addTemplateNode(node, content[0], template)
	}
	// check for errors
	return scanner.Err()
}

func init() {
	// load lang files
	langFiles := map[uint][]byte{
		LANG_ZH_CN: zh_CN,
	}
	// parse lang files
	langNodes = make(map[uint]*langNode)
	for langCode, content := range langFiles {
		node := &langNode{
			subNodes: make(map[string]*langNode),
		}
		if err := readLangFile(node, content); err != nil {
			panic(err)
		}
		langNodes[langCode] = node
	}
}
