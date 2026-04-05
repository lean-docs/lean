# Test Corpus Import Tracking

Each external corpus is imported at a pinned commit. Fixtures are not code —
document files (.docx, .md, .html) from GPL/AGPL projects are usable as test
data. Test logic copied from source code must be MIT/BSD/Apache 2.0 only.

---

## Priority 1 — Blocking Cluster 2 (Markdown Parser)

### CommonMark Spec
- **Source:** https://github.com/commonmark/commonmark-spec
- **License:** CC BY-SA 4.0
- **Pinned commit:** TBD
- **Examples:** 652 numbered input/output pairs
- **Import method:** Script `scripts/import-commonmark-spec.go` parses spec, writes individual .md/.html files
- **Target:** `testdata/fixtures/commonmark/` and `testdata/golden/commonmark/`
- **Test file:** `pkg/parser/md/commonmark_spec_test.go` — table-driven, iterates all 652
- **Status:** [ ] Not started

### cmark Regression Tests
- **Source:** https://github.com/commonmark/cmark `test/` directory
- **License:** BSD-2
- **Import method:** Copy relevant .txt fixtures
- **Target:** `testdata/fixtures/cmark/`
- **Status:** [ ] Not started

### goldmark Test Helpers
- **Source:** https://github.com/yuin/goldmark
- **License:** MIT
- **What to copy:** `testutil/` helpers — `parseMarkdown`, `normalizeHTML`, table-driven patterns
- **Target:** Adapt into `pkg/parser/md/testutil_test.go`
- **Status:** [ ] Not started

### blackfriday Regression Tests
- **Source:** https://github.com/russross/blackfriday `testdata/`
- **License:** BSD-2
- **What to copy:** Edge-case .md files accumulated from real-world breakage
- **Target:** `testdata/fixtures/blackfriday/`
- **Status:** [ ] Not started

---

## Priority 2 — Blocking Cluster 3 (HTML Parser)

### html5lib-tests
- **Source:** https://github.com/html5lib/html5lib-tests
- **License:** MIT
- **What to copy:** `tree-construction/` subdirectory (.dat files)
- **Import method:** Script `scripts/import-html5lib-tests.go` converts .dat → individual .html + .tree golden files
- **Scope:** Filter to block elements, inline formatting, tables, lists. Skip tokenization-only and script/template tests.
- **Target:** `testdata/fixtures/html5lib/` and `testdata/golden/html5lib/`
- **Test file:** `pkg/parser/html/html5lib_spec_test.go`
- **Status:** [ ] Not started

### golang.org/x/net/html parse_test.go
- **Source:** https://github.com/golang/net/tree/master/html
- **License:** BSD-3
- **What to copy:** Test harness pattern for html5lib .dat format, `parseFragment` helper
- **Target:** Adapt into test helpers
- **Status:** [ ] Not started

---

## Priority 3 — Blocking Clusters 4-7 (OOXML Parser)

### python-docx Fixtures
- **Source:** https://github.com/python-openxml/python-docx `tests/`
- **License:** MIT
- **What to copy:** .docx fixture files for specific features
- **Fixtures needed:**
  - [ ] Basic text/runs
  - [ ] Bold, italic, underline formatting
  - [ ] Tables (simple, merged cells)
  - [ ] Images (inline, floating)
  - [ ] Lists (bullet, numbered, nested)
  - [ ] Styles
- **Target:** `testdata/fixtures/ooxml/python-docx/`
- **Status:** [ ] Not started

### LibreOffice Test Documents
- **Source:** https://gerrit.libreoffice.org — `sw/qa/extras/ooxmlexport/data/` and `sw/qa/extras/ww8import/data/`
- **License:** MPL 2.0 (fixtures are documents, not code)
- **What to copy:** Selective — one fixture per test ID, documented in SOURCE.md
- **Fixtures needed:**
  - [ ] Theme colors (tdf* files with theme references)
  - [ ] Tracked changes (insert, delete)
  - [ ] Complex tables (nested, merged, styled)
  - [ ] Numbering edge cases (restart, override, multi-level)
  - [ ] Headers/footers (per-section, first-page, even/odd)
  - [ ] Tab stops (all alignment types, leaders)
  - [ ] Section breaks (continuous, next page)
  - [ ] Footnotes/endnotes
  - [ ] BiDi text
  - [ ] Multiple sections with different page sizes
- **Target:** `testdata/fixtures/ooxml/libreoffice/`
- **Status:** [ ] Not started

### docx4j Test Corpus
- **Source:** https://github.com/plutext/docx4j `docx4j-core/src/test/resources/`
- **License:** Apache 2.0
- **What to copy:** Complex OOXML constructs not covered by python-docx or LibreOffice
- **Fixtures needed:**
  - [ ] Numbering edge cases
  - [ ] Section breaks
  - [ ] Content controls
  - [ ] Complex style inheritance chains
- **Target:** `testdata/fixtures/ooxml/docx4j/`
- **Status:** [ ] Not started

### python-docx Test Logic
- **Source:** https://github.com/python-openxml/python-docx `tests/unit/`
- **License:** MIT
- **What to adapt:** Assertion patterns from `test_table.py`, `test_text.py`, `test_styles.py`
- **Method:** Read Python tests, rewrite assertions in Go. Attribute as "Derived from python-docx"
- **Status:** [ ] Not started

### gooxml/baliance Test Helpers
- **Source:** https://github.com/baliance/gooxml `testhelper/`
- **License:** MIT
- **What to copy:** `DocumentFromFile`, XML comparison utilities
- **Target:** Adapt into `pkg/parser/ooxml/testutil_test.go`
- **Status:** [ ] Not started

---

## Priority 4 — Blocking Cluster 12 (Mutable IR)

### ProseMirror Transform Tests
- **Source:** https://github.com/prosemirror/prosemirror-transform `test/test-trans.js`
- **License:** MIT
- **What to adapt:** Split, merge, replace, mark operation tests (hundreds of cases)
- **Method:** Translate JS test cases to Go. Attribute as "Translated from ProseMirror"
- **Status:** [ ] Not started

### Yjs Tests
- **Source:** https://github.com/yjs/yjs `test/`
- **License:** MIT
- **What to adapt:** Concurrent text insert semantics, delete-insert ordering
- **Status:** [ ] Not started

### Automerge Tests
- **Source:** https://github.com/automerge/automerge
- **License:** MIT
- **What to adapt:** Operation serialization round-trips, undo/redo correctness
- **Status:** [ ] Not started

---

## Priority 5 — Blocking Cluster 13 (Font Management)

### Google Fonts
- **Source:** https://github.com/google/fonts
- **License:** SIL Open Font License (OFL)
- **What to copy:** Small subset of font files
- **Fonts needed:**
  - [ ] Roboto-Regular.ttf — Latin, well-known metrics
  - [ ] NotoSans-Regular.ttf — CJK coverage for Unicode tests
  - [ ] Amiri-Regular.ttf — Arabic/RTL groundwork
- **Target:** `testdata/fixtures/fonts/`
- **Status:** [ ] Not started

### go-text/typesetting Test Patterns
- **Source:** https://github.com/go-text/typesetting `fontscan/`, `opentype/`
- **License:** BSD-3
- **What to adapt:** Font loading, glyph metric assertions, kerning validation patterns
- **Status:** [ ] Not started

---

## Import Rules

1. **Pin everything** — specific commit hash, recorded in SOURCE.md per directory
2. **Never bulk import** — only fixtures that serve a named test ID
3. **Document provenance** — every fixture has a SOURCE.md entry with origin, commit, license, test ID
4. **Re-import is a PR** — updating corpus = PR with description of what changed
5. **Fixtures are immutable** — never modify after commit; version suffix for updates
6. **License audit per PR** — confirm license and that fixtures (not code) are being copied
