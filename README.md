<p align="center">
  <picture>
    <source media="(prefers-color-scheme: dark)" srcset="docs/assets/logo-dark.svg">
    <img src="docs/assets/logo-light.svg" alt="Lean" height="72">
  </picture>
</p>

<p align="center">
  An open document engine in Go. Parse, edit, and export standard document formats without binding your application to one editor.
</p>

<p align="center">
  <a href="https://github.com/lean-docs/lean/actions/workflows/ci.yml"><img src="https://github.com/lean-docs/lean/actions/workflows/ci.yml/badge.svg" alt="CI"></a>
  <a href="https://pkg.go.dev/github.com/lean-docs/lean"><img src="https://pkg.go.dev/badge/github.com/lean-docs/lean.svg" alt="Go Reference"></a>
  <a href="https://github.com/lean-docs/lean/releases"><img src="https://img.shields.io/github/v/release/lean-docs/lean?include_prereleases&label=release" alt="Release"></a>
  <a href="LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue.svg" alt="MIT license"></a>
</p>

Lean reads documents into one typed model, the **Document IR**, and writes them back out. Every DOCX it opens comes with a **fidelity report**: if the file contains something Lean cannot save without loss, the document stays read only instead of being silently changed.

## What it does

```mermaid
flowchart LR
    subgraph Read
        direction TB
        DOCX_IN[DOCX]
        MD_IN[Markdown]
        HTML_IN[HTML]
    end

    IR[(Document IR)]
    CMD[[Command batches<br/>edit · suggest · resolve]]

    subgraph Write
        direction TB
        DOCX_OUT[DOCX]
        MD_OUT[Markdown]
        HTML_OUT[HTML]
        TYP[Typst]
    end

    RENDER([typst compile<br/>PDF · PNG · SVG])

    DOCX_IN --> IR
    MD_IN --> IR
    HTML_IN --> IR
    CMD <--> IR
    IR --> DOCX_OUT
    IR --> MD_OUT
    IR --> HTML_OUT
    IR --> TYP --> RENDER

    classDef core fill:#0D9488,stroke:#0F766E,color:#FFFFFF
    classDef accent fill:#CCFBF1,stroke:#0D9488,color:#134E4A
    class IR core
    class CMD accent
```

| Format | Read | Write | Notes |
| --- | :---: | :---: | --- |
| DOCX | ✓ | ✓ | Editable profile with fidelity reports and atomic command batches |
| Markdown | ✓ | ✓ | CommonMark 0.31.2 and GFM; native source files keep their exact bytes |
| HTML | ✓ | ✓ | Markdown previews are sanitized |
| Typst | | ✓ | Compile with the Typst CLI for PDF, PNG, and SVG |
| ODT, EPUB, plain text | | | Planned |

## Install

```bash
go get github.com/lean-docs/lean@v0.1.0-alpha.2
```

Document editing and the `lean docx` command line arrive in `v0.1.0-alpha.3`. Until that release is tagged, use `@main` to try them.

## Edit a DOCX

Open the file, check that it is editable, apply a batch of commands, and save:

```go
input, err := os.ReadFile("report.docx")
if err != nil {
	log.Fatal(err)
}

document, fidelity, err := lean.OpenDOCX(input)
if err != nil {
	log.Fatal(err)
}
if !fidelity.Editable {
	log.Fatalf("read only: %v", fidelity.Unsupported)
}

first := document.Sections[0].Blocks[0].BlockID()
edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{
	{ID: "paragraph.style", Block: first, Style: "Title"},
	{ID: "paragraph.insert", At: 1, Text: "First action"},
})
if err != nil {
	log.Fatal(err) // the input document is unchanged
}

output, _, err := lean.SaveDOCX(edited)
```

A batch runs on a copy of the document. If any command is rejected, nothing is applied:

```mermaid
flowchart TD
    A[OpenDOCX] --> B{Editable?}
    B -- no --> R[Read only<br/>fidelity.Unsupported says why]
    B -- yes --> C[ApplyDOCXCommands<br/>on a copy]
    C -- a command fails --> U[Error<br/>input unchanged]
    C -- all pass --> D[SaveDOCX]
    D -- unsupported result --> X[Save refused]
    D -- ok --> E[DOCX bytes<br/>+ fidelity report]

    classDef ok fill:#CCFBF1,stroke:#0D9488,color:#134E4A
    classDef stop fill:#FEE2E2,stroke:#DC2626,color:#7F1D1D
    class E ok
    class R,U,X stop
```

Commands cover text and formatting, paragraphs, named styles, lists, tables, page layout, sections, headers and footers, footnotes, bookmarks, drawings, watermarks, and tracked changes. `lean.DOCXCommands()` lists every command with the capabilities it needs, and `lean.DOCXCommandSchema(id)` returns its parameter schema.

## Command line

```sh
go install github.com/lean-docs/lean/cmd/lean@main

lean docx commands                                  # list commands as JSON
lean docx inspect input.docx                        # document and fidelity report
lean docx apply input.docx commands.json out.docx   # apply a batch, write a new file
```

`commands.json` holds one JSON array of commands. The output file must not already exist.

## Native Markdown

`OpenMarkdownSource` opens a UTF-8 Markdown file without converting it. `Bytes` returns the original file exactly, including byte order mark, line endings, frontmatter, comments, reference links, and syntax Lean does not model.

| Method | Output | Safe to display? |
| --- | --- | --- |
| `RenderHTML()` | CommonMark with tables, task lists, strikethrough, and autolinks | Yes. Scripts, inline styles, and dangerous links are removed |
| `RenderSyntaxHTML(CommonMark)` | Spec output for conversion and conformance | No. Sanitize before display |
| `RenderSyntaxHTML(GitHubFlavoredMarkdown)` | CommonMark plus the five GFM 0.29 extensions | No. Sanitize before display |

Applications still need to isolate previews and control external resources and navigation. The Markdown parser and exporter are conversion APIs; their normalized output is not a lossless save of the original file.

## Guides

- [Document editing](docs/document-editing.md): commands, positions, and the CLI
- [Tracked changes](docs/tracked-changes.md): suggestions, accept and reject, table styles, and watermarks
- [Explicit run formatting](docs/formatting-overrides.md): on, off, and inherited formatting
- [Architecture](ARCHITECTURE.md): the document model and design decisions
- [Requirements](SRS.md): requirements and test specification
- [Benchmarks](BENCHMARKS.md)

## Project layout

```
lean/
├── cmd/lean/          CLI
├── pkg/
│   ├── ir/            Document IR and JSON serialization
│   ├── parser/        DOCX (OOXML), Markdown, and HTML readers
│   ├── export/        DOCX (OOXML), Markdown, HTML, and Typst writers
│   ├── numbering/     List numbering and markers
│   ├── units/         Measurement conversion
│   ├── layout/        Page layout (not implemented yet)
│   └── font/ hittest/ render/ selection/ wasm/   Reserved for layout and rendering
├── docs/              Guides, release notes, and assets
└── testdata/          Fixtures, golden files, and reference documents
```

## Build and test

```bash
git clone https://github.com/lean-docs/lean.git
cd lean
go build ./...
go test -race -count=1 ./...
```

The Typst export tests compile their output, so they need the `typst` binary on your `PATH`.

## Status

Lean is alpha software. The public API and fidelity guarantees can change before v1.0. Documents outside the editable profile stay read only, and the fidelity report names what each one contains. Each guide lists its own limitations.

See the [changelog](CHANGELOG.md) for what each version adds, [release notes](docs/releases) for each release, and [RELEASING.md](RELEASING.md) for the release process.

## License

MIT. See [THIRD_PARTY_NOTICES.md](THIRD_PARTY_NOTICES.md) for the licenses covering imported conformance fixtures.
