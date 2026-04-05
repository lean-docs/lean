# Architecture

This document explains the design decisions behind Lean and the reasoning
that informed them.

---

## Document IR

The IR (intermediate representation) is the in-memory data structure that
represents a document. Every parser produces it; every exporter consumes it.
It is the single source of truth.

### Model

Lean uses a **flat-run model**: a document is a tree of sections containing
blocks (paragraphs, tables, images), where each paragraph contains a flat
list of runs. A run is a contiguous span of text sharing the same formatting
properties.

```
Document
 └─ Section
     ├─ Paragraph
     │   ├─ Run { text: "Hello ", bold: true }
     │   └─ Run { text: "world" }
     ├─ Table
     │   └─ Row → Cell → [Blocks...]
     └─ Image
```

This mirrors how OOXML represents documents internally — paragraphs contain
runs (`w:r`), and runs cannot nest. It also matches the Google Docs API model
(Body → Paragraph → TextRun). The flat structure makes OOXML round-tripping
lossless and keeps the parser/exporter logic straightforward.

We evaluated ProseMirror's node+mark model, where inline formatting is
represented as overlapping annotations rather than properties on runs. That
model is better suited to real-time editing (overlapping bold+italic is more
natural with marks), but Lean is a document engine first. The editing layer
can project a marks-based view on top of the flat IR when needed.

### Block IDs

Every block carries a stable, unique ID. This is uncommon — OOXML, ProseMirror,
Slate, and Typst do not assign stable IDs to blocks. Lean does because:

- Dirty tracking for incremental layout needs to know which blocks changed
- Collaboration requires stable references to document positions
- The operation model (insert, delete, format) addresses blocks by ID

### Style Resolution

OOXML defines a five-layer inheritance chain for formatting:

1. Document defaults (`w:docDefaults`)
2. Named style (`basedOn` chain)
3. Paragraph style
4. Character style
5. Direct formatting (inline on the run)

Each layer can override properties from the layer below it. Lean models this
chain explicitly. The `StyleSheet.ResolveStyle` function walks the `basedOn`
chain and merges properties.

Most document libraries skip this — `python-docx`, `docx-js`, and `go-docx`
all resolve styles partially or not at all. Getting this right is necessary
for faithful document reproduction.

### Theme Colors

OOXML theme colors (`accent1`, `dk1`, `lt1`, etc.) are defined in
`theme1.xml` and referenced by name throughout the document. Resolving them
to RGB requires loading the theme, mapping the color name, and applying
tint/shade transforms.

No widely-used JavaScript or Python library handles this correctly. Lean
resolves theme colors at parse time and stores the final RGB value in the IR.

---

## Format Strategy

### Native Formats

Markdown (CommonMark) and HTML are treated as first-class. The test suite is
written primarily against Markdown fixtures. These formats parse cleanly into
the IR and export cleanly out of it.

### OOXML (.docx)

.docx is an important import/export target but not the definition of
correctness. A .docx file is a ZIP archive containing XML files — the main
content lives in `word/document.xml`, with styles, numbering, themes,
headers, footers, and media as separate parts linked by relationships.

The OOXML parser maps directly from XML elements to IR types, following the
approach used by docx4j (the most complete open-source OOXML library). This
means the Go types mirror the spec rather than abstracting it away. A
higher-level builder API sits on top for common operations.

A minimal valid .docx requires only three files:
- `[Content_Types].xml`
- `_rels/.rels`
- `word/document.xml`

### PDF, PNG, SVG via Typst

Lean does not include a layout engine or PDF renderer. Instead, it exports
the IR to [Typst](https://typst.app) markup and delegates typesetting to the
Typst compiler.

Typst is a modern typesetting system written in Rust (Apache 2.0). It
compiles documents to PDF, PNG, and SVG in milliseconds, supports incremental
compilation, and handles tables, images, footnotes, headers/footers,
multi-column layout, and mathematical notation.

Integration options, in order of preference:

| Method | External dependency | Latency |
|--------|-------------------|---------|
| CLI (`typst compile`) | `typst` binary on PATH | ~ms |
| WASM via [wazero](https://github.com/tetratelabs/wazero) | None (embedded) | ~ms |

The CLI approach is used initially. WASM embedding is planned for
zero-dependency distribution.

Typst cannot produce .docx files — that path is handled entirely by Lean's
native OOXML exporter.

---

## Collaboration

The operation model (insert, delete, format, split, merge) is designed to be
compatible with the [Yjs](https://github.com/yjs/yjs) CRDT protocol at the
boundary. Yjs has the largest ecosystem for real-time collaboration, with
bindings for ProseMirror, TipTap, and Slate.

Lean itself does not implement a CRDT. It provides a set of deterministic
operations over the IR that a collaboration layer can drive.

---

## Dependencies

All dependencies must be MIT, BSD, or Apache 2.0.

| Package | Purpose |
|---------|---------|
| `github.com/yuin/goldmark` | CommonMark Markdown parser |
| `golang.org/x/net/html` | HTML parser |
| `github.com/stretchr/testify` | Test assertions |

OOXML parsing and generation is implemented from scratch — existing Go
libraries are either commercially licensed or too limited.

PDF/PNG rendering is handled by Typst (external tool, not a Go dependency).

---

## Test Strategy

Tests are organized into clusters of ascending complexity. A cluster cannot
begin until every test in the prior cluster is green. This is enforced by CI.

Fixtures come from established open-source test corpora:
- CommonMark spec (652 examples) for Markdown
- html5lib-tests for HTML
- LibreOffice and python-docx test fixtures for OOXML

Layout and rendering tests use Markdown fixtures generated inline in the test
code, ensuring full reproducibility without external files.

See [SRS.md](SRS.md) for the full test specification.
