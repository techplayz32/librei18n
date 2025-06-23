package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/techplayz32/librei18n/i18n"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: librei18n <extract|merge> [options]")
		os.Exit(1)
	}
	switch os.Args[1] {
	case "extract":
		var src, out string
		flagSet := flag.NewFlagSet("extract", flag.ExitOnError)
		flagSet.StringVar(&src, "src", "", "Go source file")
		flagSet.StringVar(&out, "out", "", "Output TOML file")
		flagSet.Parse(os.Args[2:])
		if src == "" || out == "" {
			log.Fatal("-src and -out are required")
		}
		messages, err := i18n.ExtractMessagesFromGoFile(src)
		if err != nil {
			log.Fatal(err)
		}
		if err := i18n.WriteMessagesToTOML(messages, out); err != nil {
			log.Fatal(err)
		}
		fmt.Println("Extracted messages written to", out)
	case "merge":
		var src, dst, out string
		flagSet := flag.NewFlagSet("merge", flag.ExitOnError)
		flagSet.StringVar(&src, "src", "", "Source TOML file (e.g. active.en.toml)")
		flagSet.StringVar(&dst, "dst", "", "Destination TOML file (e.g. translate.es.toml)")
		flagSet.StringVar(&out, "out", "", "Output TOML file (optional, default: dst)")
		flagSet.Parse(os.Args[2:])
		if src == "" || dst == "" {
			log.Fatal("-src and -dst are required")
		}
		if out == "" {
			out = dst
		}
		srcMap, err := i18n.LoadMessages(src)
		if err != nil {
			log.Fatalf("Failed to load src: %v", err)
		}
		dstMap, err := i18n.LoadMessages(dst)
		if err != nil {
			log.Fatalf("Failed to load dst: %v", err)
		}
		// merge: add missing keys from srcMap to dstMap
		for k, v := range srcMap {
			if _, ok := dstMap[k]; !ok {
				dstMap[k] = v
			}
		}
		// write merged mapp to out
		ext := filepath.Ext(out)
		var writeErr error
		switch ext {
		case ".toml":
			// convert to []*Message for TOML writerttt
			var messages []*i18n.Message
			for k, v := range dstMap {
				msg := &i18n.Message{ID: k}
				if m, ok := v.(map[string]interface{}); ok {
					if one, ok := m["one"].(string); ok {
						msg.One = one
					}
					if other, ok := m["other"].(string); ok {
						msg.Other = other
					}
				}
				messages = append(messages, msg)
			}
			writeErr = i18n.WriteMessagesToTOML(messages, out)
		case ".yaml", ".yml":
			writeErr = i18n.PrettyPrintAndSortYAML(dstMap, out)
		case ".json":
			data, err := json.MarshalIndent(dstMap, "", "  ")
			if err != nil {
				writeErr = err
			} else {
				writeErr = os.WriteFile(out, data, 0644)
			}
		default:
			writeErr = fmt.Errorf("unsupported output format: %s", ext)
		}
		if writeErr != nil {
			log.Fatalf("Failed to write merged file: %v", writeErr)
		}
		fmt.Println("Merged messages written to", out)
	default:
		fmt.Println("Unknown command:", os.Args[1])
		fmt.Println("Usage: librei18n <extract|merge> [options]")
		os.Exit(1)
	}
}
