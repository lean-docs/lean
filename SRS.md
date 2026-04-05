# Lean — Requirements and Test Specification

> Every line of implementation is preceded by a failing test.

---

## Scope

Lean is a document engine. It provides:

1. **Parsers** — Markdown, HTML, and .docx into a clean IR
2. **Exporters** — IR to Markdown, HTML, .docx, and Typst
3. **Operation model** — formal mutations over the IR (insert, delete, format, undo/redo)
4. **Typst integration** — IR to Typst markup for PDF/PNG/SVG rendering

The editing layer (UI, toolbar, selection, collaboration) is out of scope for
this repository.

---

## IR (Intermediate Representation)

The IR is the source of truth. It must be:

- Serializable to/from JSON without loss
- Versionable (schema version tracked in metadata)
- Traversable with stable block IDs

The full type definitions are in `pkg/ir/types.go`. The key structures:

- `Document` → `Section` → `Block` (Paragraph, Table, Image, Bookmark)
- `Paragraph` → flat list of `Run` (text + formatting attributes)
- `Table` → `TableRow` → `TableCell` → nested `Block` list
- `StyleSheet` with named styles, `basedOn` inheritance, and theme colors
- `NumberingDef` with multi-level list definitions

---

## Test Clusters

Tests are grouped into clusters of ascending complexity.
**Every cluster must be fully green before the next begins.**

Test function names match the IDs exactly. No test depends on another test's
side effects. No test makes network calls.

---

### Cluster 0 — Project Bootstrap

> Repo compiles, test harness runs, CI is configured.

| ID | Test | Pass Condition |
|----|------|----------------|
| C0.1 | `TestBuildCompiles` | `go build ./...` exits 0 |
| C0.2 | `TestTestHarnessRuns` | `go test ./...` runs without panic |
| C0.3 | `TestVersionString` | `lean.Version()` returns valid semver |
| C0.4 | `TestEmptyDocumentIR` | `NewDocument()` returns valid zero-value IR |
| C0.5 | `TestIRJSONRoundTrip` | Empty document round-trips via JSON |
| C0.6 | `TestCIConfigPresent` | `.github/workflows/ci.yml` exists with test + lint steps |
| C0.7 | `TestGoldenUpdateFlag` | `go test ./... -update` regenerates golden files |

---

### Cluster 1 — IR Fundamentals

> The document model is correct, traversable, and round-trippable.

| ID | Test | Pass Condition |
|----|------|----------------|
| C1.1 | `TestIRSingleParagraph` | One paragraph round-trips via JSON |
| C1.2 | `TestIRRunAttrs` | Bold, italic, underline, strike, smallcaps, allcaps survive |
| C1.3 | `TestIRFontProperties` | FontName, FontSize, Tracking survive |
| C1.4 | `TestIRColorRGB` | RGB color survives |
| C1.5 | `TestIRColorNone` | Color.None=true survives |
| C1.6 | `TestIRThemeColors` | Theme color map survives |
| C1.7 | `TestIRMultipleParagraphs` | 10 paragraphs, order preserved |
| C1.8 | `TestIRParagraphAlignment` | Left, center, right, justify |
| C1.9 | `TestIRSpacing` | Before, after, line, lineRule |
| C1.10 | `TestIRIndent` | Left, right, firstLine, hanging |
| C1.11 | `TestIRTabStops` | Position, alignment, leader |
| C1.12 | `TestIRStyleReference` | Paragraph style name survives |
| C1.13 | `TestIRCharacterStyle` | IsChar=true distinguished |
| C1.14 | `TestIRStyleInheritance` | basedOn chain resolves correctly |
| C1.15 | `TestIRStyleSheet` | Full stylesheet round-trips |
| C1.16 | `TestIRNumberingDef` | 3-level numbering survives |
| C1.17 | `TestIRNumberingRef` | Paragraph NumberingRef survives |
| C1.18 | `TestIRTable` | 2x2 table with borders and shading |
| C1.19 | `TestIRTableSpan` | ColSpan, RowSpan survive |
| C1.20 | `TestIRImage` | Image bytes, format, dimensions |
| C1.21 | `TestIRHyperlink` | URL and bookmark ref |
| C1.22 | `TestIRFootnote` | Footnote with blocks |
| C1.23 | `TestIRHeaderFooter` | Header and footer blocks |
| C1.24 | `TestIRSection` | PageProperties and columns |
| C1.25 | `TestIRBookmark` | Bookmark ID and name |
| C1.26 | `TestIRBlockIDs` | Every block has a non-empty stable ID |
| C1.27 | `TestIRKeepTogether` | KeepTogether, KeepWithNext |
| C1.28 | `TestIRBiDi` | BiDi flag |
| C1.29 | `TestIRBaselineShift` | Superscript, subscript |
| C1.30 | `TestIRBreakTypes` | Line, page, column breaks |

---

### Cluster 2 — Markdown Parser

> CommonMark Markdown parsed into IR.

| ID | Test | Pass Condition |
|----|------|----------------|
| C2.1 | `TestMDEmptyDocument` | Empty string → one empty section |
| C2.2 | `TestMDPlainParagraph` | Plain text → single paragraph, one run |
| C2.3 | `TestMDMultipleParagraphs` | Blank line separates two paragraphs |
| C2.4 | `TestMDBold` | `**text**` → Bold=true |
| C2.5 | `TestMDItalic` | `*text*` → Italic=true |
| C2.6 | `TestMDBoldItalic` | `***text***` → both |
| C2.7 | `TestMDStrike` | `~~text~~` → Strike=true |
| C2.8 | `TestMDInlineCode` | `` `code` `` → monospace font |
| C2.9 | `TestMDHeading1` | `# Title` → Style="Heading1" |
| C2.10 | `TestMDHeading2` | `## Title` → Style="Heading2" |
| C2.11 | `TestMDHeading3to6` | H3–H6 → correct style names |
| C2.12 | `TestMDBulletListHyphen` | `- item` → bullet NumberingRef |
| C2.13 | `TestMDBulletListAsterisk` | `* item` → same |
| C2.14 | `TestMDOrderedList` | `1. item` → decimal NumberingRef |
| C2.15 | `TestMDNestedList` | Indented item → level 1 |
| C2.16 | `TestMDBlockquote` | `> text` → left indent |
| C2.17 | `TestMDCodeBlock` | Fenced block → monospace font |
| C2.18 | `TestMDHorizontalRule` | `---` → bottom border |
| C2.19 | `TestMDLink` | `[text](url)` → Hyperlink.URL |
| C2.20 | `TestMDImage` | `![alt](url)` → Image with Alt |
| C2.21 | `TestMDLineBreak` | Trailing double space → Break=line |
| C2.22 | `TestMDHardBreak` | Trailing backslash → Break=line |
| C2.23 | `TestMDTable` | GFM table → Table with rows/cells |
| C2.24 | `TestMDTableAlignment` | Column alignment → cell align |
| C2.25 | `TestMDFrontmatter` | YAML frontmatter → DocumentMeta |
| C2.26 | `TestMDUnicodeText` | CJK, Arabic, emoji parse correctly |
| C2.27 | `TestMDMixedInlineFormatting` | Bold inside italic inside link |

---

### Cluster 3 — HTML Parser

> HTML parsed into IR.

| ID | Test | Pass Condition |
|----|------|----------------|
| C3.1 | `TestHTMLEmptyBody` | Empty body → empty section |
| C3.2 | `TestHTMLParagraph` | `<p>` → paragraph |
| C3.3 | `TestHTMLHeadings` | `<h1>`–`<h6>` → heading styles |
| C3.4 | `TestHTMLBold` | `<strong>` and `<b>` → Bold=true |
| C3.5 | `TestHTMLItalic` | `<em>` and `<i>` → Italic=true |
| C3.6 | `TestHTMLUnderline` | `<u>` → Underline |
| C3.7 | `TestHTMLStrike` | `<s>` and `<del>` → Strike=true |
| C3.8 | `TestHTMLAnchor` | `<a href>` → Hyperlink.URL |
| C3.9 | `TestHTMLImage` | `<img>` → Image with Alt |
| C3.10 | `TestHTMLUnorderedList` | `<ul><li>` → bullet numbering |
| C3.11 | `TestHTMLOrderedList` | `<ol><li>` → decimal numbering |
| C3.12 | `TestHTMLNestedList` | Nested `<ul>` → level 1 |
| C3.13 | `TestHTMLTable` | `<table>` → Table block |
| C3.14 | `TestHTMLTableHeader` | `<th>` / `<thead>` → IsHeader |
| C3.15 | `TestHTMLColspan` | `colspan="2"` → ColSpan=2 |
| C3.16 | `TestHTMLInlineStyle` | `font-weight:bold` → Bold=true |
| C3.17 | `TestHTMLInlineColor` | `color:#FF0000` → Color RGB |
| C3.18 | `TestHTMLInlineFontSize` | `font-size:14pt` → FontSize=14 |
| C3.19 | `TestHTMLBlockquote` | `<blockquote>` → indented paragraph |
| C3.20 | `TestHTMLPreCode` | `<pre><code>` → monospace block |
| C3.21 | `TestHTMLLineBreak` | `<br>` → Break=line |
| C3.22 | `TestHTMLHorizontalRule` | `<hr>` → bottom border |
| C3.23 | `TestHTMLEntities` | `&amp;`, `&lt;`, `&nbsp;` decoded |
| C3.24 | `TestHTMLMalformedGraceful` | Malformed HTML → best-effort IR, no panic |

---

### Cluster 4 — OOXML Parser (Text)

> Parse .docx text content into IR.

| ID | Test | Pass Condition |
|----|------|----------------|
| C4.1 | `TestOOXMLMinimalDocx` | Smallest valid .docx parses |
| C4.2 | `TestOOXMLPlainText` | Single run text extracted |
| C4.3 | `TestOOXMLMultipleRuns` | Mixed bold/normal runs |
| C4.4 | `TestOOXMLParagraphStyles` | Normal, Heading1–6 extracted |
| C4.5 | `TestOOXMLCharacterStyles` | IsChar=true extracted |
| C4.6 | `TestOOXMLStyleInheritance` | Run inherits paragraph style |
| C4.7 | `TestOOXMLThemeColors` | Theme color map from theme1.xml |
| C4.8 | `TestOOXMLThemeColorResolution` | Theme color → RGB |
| C4.9–C4.14 | Bold, italic, underline, strike, smallcaps, allcaps | Formatting flags extracted |
| C4.15–C4.16 | Superscript, subscript | Baseline shift extracted |
| C4.17–C4.20 | Font size, name, color, tracking | Run properties extracted |
| C4.21–C4.24 | Spacing, indent, alignment, tab stops | Paragraph properties extracted |
| C4.25–C4.27 | KeepTogether, KeepWithNext, PageBreakBefore | Pagination controls extracted |
| C4.28–C4.30 | Line break, page break, column break | Break types extracted |
| C4.31 | `TestOOXMLHyperlink` | URL and run text extracted |
| C4.32 | `TestOOXMLBookmark` | Bookmark extracted |
| C4.33 | `TestOOXMLBiDi` | BiDi flag extracted |
| C4.34–C4.35 | Multiple sections, page properties | Sections and dimensions extracted |
| C4.36–C4.38 | Tracked insert, delete, comments | Revision markup extracted |
| C4.39–C4.41 | Footnote, header, footer | Document parts extracted |
| C4.42 | `TestOOXMLCorruptDocxReturnsError` | Malformed ZIP → error, no panic |

---

### Cluster 5 — OOXML Parser (Lists)

| ID | Test | Pass Condition |
|----|------|----------------|
| C5.1–C5.2 | Bullet list, numbered list | NumberingRef with correct format |
| C5.3 | Nested list | Levels 0 and 1 distinguished |
| C5.4–C5.5 | List continuation, resume | Independent and resumed counters |
| C5.6 | List indent | Indent levels → correct points |
| C5.7 | Restart override | `startOverride` resets counter |
| C5.8–C5.9 | Run formatting, multi-line items | Formatting preserved in list items |

---

### Cluster 6 — OOXML Parser (Tables)

| ID | Test | Pass Condition |
|----|------|----------------|
| C6.1–C6.3 | Simple table, cell text, column widths | Structure and dimensions |
| C6.4–C6.7 | Cell borders, table borders, shading, padding | Formatting extracted |
| C6.8–C6.9 | ColSpan, RowSpan | Merge spans extracted |
| C6.10–C6.12 | Header row, alignment, table style | Table-level properties |
| C6.13–C6.15 | Nested table, vertical align, row height | Complex table features |

---

### Cluster 7 — OOXML Parser (Images)

| ID | Test | Pass Condition |
|----|------|----------------|
| C7.1–C7.2 | Inline PNG, JPEG | Bytes and format extracted |
| C7.3–C7.5 | Dimensions, relationship, alt text | Image properties extracted |
| C7.6–C7.7 | Float left, float right | Float mode extracted |
| C7.8 | Missing image | Placeholder returned, no panic |

---

### Cluster 8 — Typst Exporter

> IR converts to valid Typst markup for rendering.

| ID | Test | Pass Condition |
|----|------|----------------|
| C8.1 | `TestTypstEmptyDocument` | Empty IR → valid .typ |
| C8.2 | `TestTypstParagraph` | Plain text paragraph |
| C8.3 | `TestTypstBoldItalic` | `*bold*` and `_italic_` markup |
| C8.4 | `TestTypstHeadings` | `=`, `==`, etc. heading levels |
| C8.5 | `TestTypstBulletList` | `- item` list markup |
| C8.6 | `TestTypstNumberedList` | `+ item` or `1.` list markup |
| C8.7 | `TestTypstTable` | `table()` function call |
| C8.8 | `TestTypstImage` | `image()` function call |
| C8.9 | `TestTypstLink` | `link()` function call |
| C8.10 | `TestTypstPageSetup` | Page size and margins |
| C8.11 | `TestTypstHeaderFooter` | Header and footer content |
| C8.12 | `TestTypstFootnote` | `footnote[]` markup |
| C8.13 | `TestTypstColors` | `text(fill: rgb(...))` |
| C8.14 | `TestTypstFontSize` | `text(size: ...)` |
| C8.15 | `TestTypstMultiColumn` | `columns()` layout |
| C8.16 | `TestTypstCompiles` | Output compiles with `typst compile` |

---

### Cluster 9 — Markdown Exporter

| ID | Test | Pass Condition |
|----|------|----------------|
| C9.1 | `TestExportMDPlainText` | Plain paragraph → text |
| C9.2 | `TestExportMDBold` | Bold → `**text**` |
| C9.3 | `TestExportMDItalic` | Italic → `*text*` |
| C9.4 | `TestExportMDHeadings` | Heading styles → `#` through `######` |
| C9.5 | `TestExportMDBulletList` | Bullet list → `- item` |
| C9.6 | `TestExportMDNumberedList` | Numbered list → `1. item` |
| C9.7 | `TestExportMDTable` | Table → GFM table syntax |
| C9.8 | `TestExportMDLink` | Hyperlink → `[text](url)` |
| C9.9 | `TestExportMDImage` | Image → `![alt](ref)` |
| C9.10 | `TestExportMDCodeBlock` | Monospace block → fenced block |
| C9.11 | `TestExportMDRoundTrip` | Parse → export → re-parse → identical IR |

---

### Cluster 10 — HTML Exporter

| ID | Test | Pass Condition |
|----|------|----------------|
| C10.1 | `TestExportHTMLValid` | Valid HTML5 |
| C10.2 | `TestExportHTMLText` | Plain text |
| C10.3 | `TestExportHTMLFormatting` | Semantic tags for bold, italic, underline |
| C10.4 | `TestExportHTMLHeadings` | `<h1>`–`<h6>` |
| C10.5 | `TestExportHTMLLists` | `<ul>` / `<ol>` with nesting |
| C10.6 | `TestExportHTMLTable` | `<thead>` and `<tbody>` |
| C10.7 | `TestExportHTMLImage` | `<img>` with alt |
| C10.8 | `TestExportHTMLHyperlink` | `<a href>` |
| C10.9 | `TestExportHTMLRoundTrip` | Parse → export → re-parse → identical IR |
| C10.10 | `TestExportHTMLAccessibility` | ARIA roles and lang attribute |

---

### Cluster 11 — OOXML Exporter

| ID | Test | Pass Condition |
|----|------|----------------|
| C11.1 | `TestExportOOXMLMinimal` | Empty document → openable .docx |
| C11.2 | `TestExportOOXMLTextRoundTrip` | Parse → export → re-parse → identical IR |
| C11.3 | `TestExportOOXMLFormatting` | Bold, italic, underline, color survive |
| C11.4 | `TestExportOOXMLListRoundTrip` | Lists survive |
| C11.5 | `TestExportOOXMLTableRoundTrip` | 2x2 table survives |
| C11.6 | `TestExportOOXMLImageRoundTrip` | Image bytes identical |
| C11.7 | `TestExportOOXMLTabStops` | Tab stops survive |
| C11.8 | `TestExportOOXMLHeaderFooter` | Header and footer survive |
| C11.9 | `TestExportOOXMLFootnote` | Footnote survives |
| C11.10 | `TestExportOOXMLValidZip` | Valid ZIP archive |
| C11.11 | `TestExportOOXMLContentTypes` | Required content type entries |
| C11.12 | `TestExportOOXMLRelationships` | Correct resource references |

---

### Cluster 12 — Mutable IR

> IR supports mutation with dirty tracking for incremental operations.

| ID | Test | Pass Condition |
|----|------|----------------|
| C12.1 | `TestMutableIRInsertText` | Text insertion mutates run |
| C12.2 | `TestMutableIRDeleteText` | Text deletion removes characters |
| C12.3 | `TestMutableIRFormatRun` | Format applies to character range |
| C12.4 | `TestMutableIRInsertParagraph` | Paragraph inserted at path |
| C12.5 | `TestMutableIRDeleteParagraph` | Paragraph removed at path |
| C12.6 | `TestMutableIRSplitParagraph` | Split at offset |
| C12.7 | `TestMutableIRMergeParagraph` | Merge correctly |
| C12.8 | `TestMutableIRInsertTable` | Table inserted at path |
| C12.9 | `TestMutableIRChangeStyle` | Style updated |
| C12.10 | `TestMutableIRDirtyTracking` | Mutated block marked dirty; siblings not |
| C12.11 | `TestMutableIREventEmit` | Change event emitted |
| C12.12 | `TestMutableIRUndoSingleOp` | Undo reverses single op |
| C12.13 | `TestMutableIRRedoSingleOp` | Redo reapplies |
| C12.14 | `TestMutableIRUndoStack` | 10 ops undone, IR matches initial state |
| C12.15 | `TestMutableIROpSerialization` | All ops round-trip via JSON |
| C12.16 | `TestMutableIRConcurrentOps` | Two concurrent ops → valid IR, no panic |

---

### Cluster 13 — Font Management

| ID | Test | Pass Condition |
|----|------|----------------|
| C13.1 | `TestFontLoadTTF` | TTF loaded, metrics available |
| C13.2 | `TestFontLoadWOFF2` | WOFF2 loaded |
| C13.3 | `TestFontMetrics` | Ascender, descender, line gap correct |
| C13.4 | `TestFontGlyphWidth` | Advance width matches expected |
| C13.5 | `TestFontFallback` | Missing font → default, no panic |
| C13.6 | `TestFontKerning` | Kerning pairs applied |
| C13.7 | `TestFontUnicode` | CJK glyphs load correctly |
| C13.8 | `TestFontCache` | Same font loaded twice → cached |

---

### Cluster 14 — Performance Baselines

| ID | Test | Pass Condition |
|----|------|----------------|
| C14.1 | `BenchmarkParseMD10Page` | < 10ms |
| C14.2 | `BenchmarkParseDocx10Page` | < 50ms |
| C14.3 | `BenchmarkExportTypst10Page` | < 20ms |
| C14.4 | `BenchmarkExportDocx10Page` | < 50ms |
| C14.5 | `BenchmarkMemory10Page` | Peak < 50MB |

---

## Version Milestones

| Version | Clusters | Description |
|---------|----------|-------------|
| 0.1.0 | 0–1 | IR defined, test harness running |
| 0.2.0 | 2–3 | Markdown and HTML parsers |
| 0.3.0 | 4–7 | OOXML parser complete |
| 0.4.0 | 8 | Typst exporter, PDF/PNG output |
| 0.5.0 | 9–11 | All exporters |
| 0.6.0 | 12 | Mutable IR and operation model |
| 0.7.0 | 13 | Font management |
| 1.0.0 | 14 | Performance validated, public API stable |

---

## Out of Scope (v1)

- Mail merge and field codes
- Macros / VBA / scripting
- Full bidirectional text layout (BiDi flag parsed; layout deferred)
- Equations / MathML / OMML
- SmartArt / DrawingML complex shapes
- Charts
- Cross-references and TOC generation
- Digital signatures
- Form fields and content controls
- ODT import (planned for v1.1)
- Hyphenation (planned for v1.1)

---

## CI Pipeline

```yaml
- go build ./...
- go vet ./...
- staticcheck ./...
- go test ./... -race -count=1
```

A cluster gate enforces that PRs touching Cluster N+1 fail if Cluster N
has any failures.
