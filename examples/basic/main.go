package main

import (
	"fmt"
	"log"
	"path/filepath"

	"github.com/techplayz32/librei18n/i18n"
)

func main() {
	// paths to your files
	basePath := filepath.Join("examples", "basic", "i18n", "en.yaml")
	targetPath := filepath.Join("examples", "basic", "i18n", "fr.yaml")
	// go html template or other thing, but im sure it works with .tmpl
	templatePath := filepath.Join("examples", "basic", "template.tmpl")

	// load translation files
	base, err := i18n.LoadMessages(basePath)
	if err != nil {
		log.Fatalf("Failed to load base: %v", err)
	}

	target, err := i18n.LoadMessages(targetPath)
	if err != nil {
		log.Fatalf("Failed to load target: %v", err)
	}

	// extract keys from template
	keys, err := i18n.ExtractKeysFromTemplate(templatePath)
	if err != nil {
		log.Fatalf("Failed to extract keys: %v", err)
	}
	fmt.Println("Keys in template:", keys)

	// check for missing and unused keys
	missing, unused := i18n.CheckMissingAndUnusedKeys(base, target)
	fmt.Println("Missing:", missing)
	fmt.Println("Unused:", unused)

	// auto-fill missing ones
	filled := i18n.AutoFillMissingKeys(base, target, "[MISSING TRANSLATION]")
	fmt.Println("Filled target:", filled)

	// pretty-print and sort the filed target YAML
	outPath := filepath.Join("examples", "basic", "i18n", "fr.filled.yaml")
	if err := i18n.PrettyPrintAndSortYAML(filled, outPath); err != nil {
		log.Fatalf("Failed to write filled YAML file: %v", err)
	}
	fmt.Println("Filled and sorted YAML written to", outPath)

	msg, err := i18n.Localize(&i18n.LocalizeConfig{
		DefaultMessage: &i18n.Message{
			ID:    "PersonCats",
			One:   "{{.Name}} has {{.Count}} cat.",
			Other: "{{.Name}} has {{.Count}} cats.",
		},
		TemplateData: map[string]interface{}{
			"Name":  "Nick",
			"Count": 2,
		},
		PluralCount: 2,
	})
	if err != nil {
		log.Fatalf("Failed to extract keys: %v", err)
	}
	fmt.Println(msg)

	sourceGoFile := filepath.Join("examples", "basic", "main.go")
	extractedMessages, err := i18n.ExtractMessagesFromGoFile(sourceGoFile)
	if err != nil {
		log.Fatalf("Failed to extract messages from Go file: %v", err)
	}

	tomlOutPath := filepath.Join("examples", "basic", "i18n", "active.en.toml")
	if err := i18n.WriteMessagesToTOML(extractedMessages, tomlOutPath); err != nil {
		log.Fatalf("Failed to write messages to TOML: %v", err)
	}
	fmt.Println("Extracted messages written to", tomlOutPath)
}
