# librei18n

> [!WARNING]
> This is still a work-in-progress library and CLI!

A simple and small Go library and CLI tool for managing and validating i18n message files (YAML, TOML, JSON) for Go projects.

## Features
- Extract translation keys from Go source files and templates
- Check for missing/unused keys in each language
- Auto-fill missing keys with English or a placeholder
- Pretty-print and sort YAML files
- Supports YAML, TOML, and JSON formats
- CLI tool for extraction and merging

## Installation

```bash
go get github.com/techplayz32/librei18n/i18n
```

Or clone and build the CLI:

```bash
git clone https://github.com/techplayz32/librei18n.git
cd librei18n
# Build CLI
cd cmd/librei18n
go build -o librei18n
```

## Library Usage

```go
import "github.com/techplayz32/librei18n/i18n"

// Load messages
messages, err := i18n.LoadMessages("active.en.toml")

// Localize a message
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
fmt.Println(msg) // Nick has 2 cats.
```

## CLI Usage

### Extract messages from Go source

```bash
librei18n extract -src path/to/source.go -out active.en.toml
```

### Merge new messages into a translation file

```bash
librei18n merge -src active.en.toml -dst translate.es.toml -out merged.es.toml
```

If `-out` is omitted, it overwrites the destination file.

## Example TOML message file

```toml
[PersonCats]
description = "The number of cats a person has"
one = "{{.Name}} has {{.Count}} cat."
other = "{{.Name}} has {{.Count}} cats."
```

## License

This repository is licensed under MIT license.
