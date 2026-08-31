# Lean

A lightweight, open-source document engine written in Go.

Lean is building a shared document representation for parsing, transforming,
and exporting standard document formats without binding an application to one editor.

## Direction

Lean is working toward this architecture:

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

The first alpha includes the document representation and an early .docx parser.
Markdown, HTML, export, editing, and rendering APIs are still under development.

## Quick Start

```bash
go get github.com/lean-docs/lean@v0.1.0-alpha.1
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

Alpha software. The public API and document fidelity guarantees can change before v1.0.
See [CHANGELOG.md](CHANGELOG.md) for the implemented scope.

## License

MIT
