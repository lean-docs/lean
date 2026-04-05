# Lean

A lightweight, open-source document engine written in Go.

Lean parses, transforms, and exports documents across formats through a single
intermediate representation.

## What It Does

```
Markdown ──┐                              ┌──→ Markdown
HTML ──────┤                              ├──→ HTML
.docx ─────┤    ┌────────────────┐        ├──→ .docx
ODT (v1.1) ┤───→│  Document IR   │───────→├──→ ODT (v1.1)
Plain text ─┘    └────────────────┘        ├──→ Typst → PDF / PNG / SVG
                        ↕                  └──→ EPUB (planned)
                  Operation Model
                 (edit, undo, redo)
```

**Parsers** read Markdown, HTML, .docx, and plain text into the IR.
ODT support is planned for v1.1.
**Exporters** write the IR back out to any supported format.
PDF, PNG, and SVG rendering is delegated to [Typst](https://typst.app).

## Quick Start

```bash
go install github.com/lean-docs/lean/cmd/lean@latest
```

## Building from Source

```bash
git clone https://github.com/lean-docs/lean.git
cd lean
go build ./...
go test ./...
```

## Project Layout

```
lean/
├── cmd/lean/              CLI binary
├── pkg/
│   ├── ir/                Document model
│   ├── parser/            Markdown, HTML, OOXML readers
│   ├── export/            Markdown, HTML, OOXML, Typst writers
│   ├── hittest/           Coordinate → document position mapping
│   ├── selection/         Selection geometry
│   └── font/              Font loading and metrics
├── testdata/              Fixtures and golden files
├── ARCHITECTURE.md        Design decisions and rationale
└── SRS.md                 Requirements and test specification
```

## Status

Early development. See [CHANGELOG.md](CHANGELOG.md) for progress.

## License

MIT
