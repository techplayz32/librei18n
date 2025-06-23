package i18n

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"text/template"

	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// TranslationMap represents a map of translation keys to values.
type TranslationMap map[string]interface{}
type QuotedString string

type Message struct {
	ID    string
	One   string
	Other string
}

type LocalizeConfig struct {
	DefaultMessage *Message
	TemplateData   map[string]interface{}
	PluralCount    int
}

func (qs QuotedString) MarshalYAML() (interface{}, error) {
	return fmt.Sprintf("%q", string(qs)), nil
}

func toQuotedMap(m TranslationMap) map[string]QuotedString {
	quoted := make(map[string]QuotedString, len(m))
	for k, v := range m {
		quoted[k] = QuotedString(fmt.Sprintf("%v", v))
	}
	return quoted
}

// LoadYAML loads a YAML translation file into a TranslationMap.
func LoadYAML(path string) (TranslationMap, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}
	var m TranslationMap
	err = yaml.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

// ExtractKeysFromTemplate extracts all translation keys from a template file.
// TODO: Implement template parsing and key extraction. [X]
func ExtractKeysFromTemplate(templatePath string) ([]string, error) {
	data, err := os.ReadFile(templatePath)
	if err != nil {
		return nil, err
	}

	re := regexp.MustCompile(`{{\\s*(t|i18n)\\s+"([^"]+)"}}`)
	matches := re.FindAllStringSubmatch(string(data), -1)

	keysMap := make(map[string]struct{})
	for _, match := range matches {
		if len(match) > 2 {
			keysMap[match[2]] = struct{}{}
		}
	}

	keys := make([]string, 0, len(keysMap))
	for k := range keysMap {
		keys = append(keys, k)
	}

	return keys, nil
}

// CheckMissingAndUnusedKeys checks for missing and unused keys in each language.
// Returns missing and unused keys.
func CheckMissingAndUnusedKeys(base TranslationMap, other TranslationMap) (missing, unused []string) {
	for k := range base {
		if _, ok := other[k]; !ok {
			missing = append(missing, k)
		}
	}

	for k := range other {
		if _, ok := base[k]; !ok {
			unused = append(unused, k)
		}
	}
	return
}

// AutoFillMissingKeys fills missing keys with English or a placeholder.
func AutoFillMissingKeys(base TranslationMap, target TranslationMap, placeholder string) TranslationMap {
	filled := make(TranslationMap)
	for k, v := range target {
		filled[k] = v
	}

	for k, v := range base {
		if _, ok := filled[k]; !ok {
			if placeholder != "" {
				filled[k] = placeholder
			} else {
				filled[k] = v
			}
		}
	}
	return filled
}

// PrettyPrintAndSortYAML pretty-prints and sorts a TranslationMap, then writes it to a file.
func PrettyPrintAndSortYAML(m TranslationMap, path string) error {
	// Sort keys
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Create a new sorted map with quoted values
	sorted := make(TranslationMap)
	for _, k := range keys {
		sorted[k] = m[k]
	}

	quoted := toQuotedMap(sorted)

	// Marshal to YAML
	data, err := yaml.Marshal(quoted)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func Localize(cfg *LocalizeConfig) (string, error) {
	var tmplStr string
	if cfg.PluralCount == 1 && cfg.DefaultMessage.One != "" {
		tmplStr = cfg.DefaultMessage.One
	} else {
		tmplStr = cfg.DefaultMessage.Other
	}

	tmpl, err := template.New(cfg.DefaultMessage.ID).Parse(tmplStr)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, cfg.TemplateData)
	if err != nil {
		return "", err
	}

	return buf.String(), nil
}

// LoadMessages loads a translation file (YAML, JSON, or TOML) into a TranslationMap.
func LoadMessages(path string) (TranslationMap, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	ext := strings.ToLower(filepath.Ext(path))
	var m TranslationMap

	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(data, &m)
	case ".json":
		err = json.Unmarshal(data, &m)
	case ".toml":
		err = toml.Unmarshal(data, &m)
	default:
		return nil, errors.New("unsupported file format: " + ext)
	}
	if err != nil {
		return nil, err
	}
	return m, nil
}

func ExtractMessagesFromGoFile(path string) ([]*Message, error) {
	fset := token.NewFileSet()
	node, err := parser.ParseFile(fset, path, nil, parser.AllErrors)
	if err != nil {
		return nil, err
	}

	var messages []*Message
	ast.Inspect(node, func(n ast.Node) bool {
		cl, ok := n.(*ast.CompositeLit)
		if !ok {
			return true
		}
		// Check if it's an i18n.Message literal
		if se, ok := cl.Type.(*ast.SelectorExpr); ok {
			if se.Sel.Name == "Message" {
				msg := &Message{}
				for _, elt := range cl.Elts {
					if kv, ok := elt.(*ast.KeyValueExpr); ok {
						key := kv.Key.(*ast.Ident).Name
						val := kv.Value.(*ast.BasicLit).Value
						val = val[1 : len(val)-1] // remove quotesss
						switch key {
						case "ID":
							msg.ID = val
						case "One":
							msg.One = val
						case "Other":
							msg.Other = val
						}
					}
				}
				messages = append(messages, msg)
			}
		}
		return true
	})
	return messages, nil
}

func WriteMessagesToTOML(messages []*Message, path string) error {
	type tomlMsg struct {
		Description string `toml:"description,omitempty"`
		One         string `toml:"one,omitempty"`
		Other       string `toml:"other,omitempty"`
	}
	tomlMap := make(map[string]tomlMsg)
	for _, m := range messages {
		tomlMap[m.ID] = tomlMsg{
			One:   m.One,
			Other: m.Other,
		}
	}
	data, err := toml.Marshal(tomlMap)
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
