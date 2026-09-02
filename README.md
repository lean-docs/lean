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

The first alpha opens and saves a conservative DOCX profile through a typed
document representation. The editable profile covers document metadata, page
and section layout, text and paragraph formatting, tables, links, numbering,
named styles, and embedded images. A fidelity report keeps unsupported
documents read only instead of silently changing them.

Lean also imports and exports CommonMark and HTML. It exports Typst for PDF,
PNG, and SVG rendering. Tracked changes, fields, comments, headers, footers,
footnotes, and mutation APIs remain outside the editable alpha profile.

## Quick Start

```bash
go get github.com/lean-docs/lean@v0.1.0-alpha.1
```

## Building from Source

```bash
git clone https://github.com/lean-docs/lean.git
cd lean
go build ./...

# Test the released DOCX surface
go test -race -count=1 . ./pkg/ir ./pkg/export/ooxml
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
Release notes are kept in [docs/releases](docs/releases).
Maintainers can publish a release by following [RELEASING.md](RELEASING.md).

## License

Lean is MIT licensed. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for
the licenses that cover imported conformance fixtures.
