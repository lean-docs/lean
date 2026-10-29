# Changelog

## [Unreleased]

## [0.1.0-alpha.2] - 2026-10-29

### Added
- Native Markdown source preservation with sanitized HTML previews and explicit CommonMark and GFM syntax rendering
- Explicit on, off, and inherited run formatting through JSON and DOCX
- All eighteen standard DOCX underline values
- Explicit underline, baseline, and character-spacing resets

### Fixed
- Named paragraph styles retain supported inherited run formatting and reject cyclic parent references
- Unknown underline styles and unsupported underline properties prevent canonical saves

## [0.1.0-alpha.1] - 2026-09-02

### Added
- Document representation with JSON serialization
- Style resolution with inherited named styles
- DOCX parsing for text, paragraphs, tables, links, numbering, named styles, images, notes, bookmarks, headers, and footers
- DOCX export for text, page and section layout, formatting, tables, links, numbering, named styles, and images
- CommonMark and HTML import and export
- Typst export for PDF, PNG, and SVG rendering
- Fidelity reports that keep unsupported documents read only
- Real LibreOffice and python-docx conformance fixtures
- Reproducible CLI archives, checksums, release notes, and automated GitHub releases
- Architecture and requirements documentation
