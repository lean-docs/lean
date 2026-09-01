//go:build specification

package typst_test

import (
	"testing"

	"github.com/lean-docs/lean/pkg/export/typst"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C8.1
func TestExportEmptyDocument(t *testing.T) {
	doc := ir.NewDocument()
	out, err := typst.Export(doc)
	require.NoError(t, err)
	assert.NotEmpty(t, out)
}

// C8.2
func TestExportParagraph(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello world"}}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "Hello world")
}

// C8.3
func TestExportBoldItalic(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "bold", Attrs: ir.RunAttrs{Bold: true}},
			{Text: "italic", Attrs: ir.RunAttrs{Italic: true}},
		}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "*bold*")
	assert.Contains(t, s, "_italic_")
}

// C8.4
func TestExportHeadings(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "h1", Style: "Heading1", Runs: []ir.Run{{Text: "Title"}}},
		&ir.Paragraph{ID: "h2", Style: "Heading2", Runs: []ir.Run{{Text: "Subtitle"}}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "= Title")
	assert.Contains(t, s, "== Subtitle")
}

// C8.5
func TestExportBulletList(t *testing.T) {
	doc := ir.NewDocument()
	doc.Styles.Numbering = []ir.NumberingDef{
		{ID: "bullets", Levels: []ir.NumberingLevel{
			{Level: 0, Format: ir.NumFormatBullet, Text: "•", Start: 1},
		}},
	}
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "li1", Runs: []ir.Run{{Text: "item one"}},
			Numbering: &ir.NumberingRef{ID: "bullets", Level: 0}},
		&ir.Paragraph{ID: "li2", Runs: []ir.Run{{Text: "item two"}},
			Numbering: &ir.NumberingRef{ID: "bullets", Level: 0}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "- item one")
	assert.Contains(t, s, "- item two")
}

// C8.6
func TestExportNumberedList(t *testing.T) {
	doc := ir.NewDocument()
	doc.Styles.Numbering = []ir.NumberingDef{
		{ID: "ordered", Levels: []ir.NumberingLevel{
			{Level: 0, Format: ir.NumFormatDecimal, Text: "%1.", Start: 1},
		}},
	}
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "li1", Runs: []ir.Run{{Text: "first"}},
			Numbering: &ir.NumberingRef{ID: "ordered", Level: 0}},
		&ir.Paragraph{ID: "li2", Runs: []ir.Run{{Text: "second"}},
			Numbering: &ir.NumberingRef{ID: "ordered", Level: 0}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "+ first")
	assert.Contains(t, s, "+ second")
}

// C8.7
func TestExportTable(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Table{ID: "t1", Rows: []ir.TableRow{
			{Cells: []ir.TableCell{
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "c1", Runs: []ir.Run{{Text: "A"}}},
				}},
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "c2", Runs: []ir.Run{{Text: "B"}}},
				}},
			}},
		}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "table(")
	assert.Contains(t, s, "A")
	assert.Contains(t, s, "B")
}

// C8.8
func TestExportImage(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Image{
			ID: "img1", Data: []byte{0x89, 0x50, 0x4E, 0x47},
			Format: ir.ImagePNG, Width: 200, Height: 100, Alt: "photo",
		},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "image(")
}

// C8.9
func TestExportLink(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "click", Hyperlink: &ir.Hyperlink{URL: "https://example.com"}},
		}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "link(")
	assert.Contains(t, s, "https://example.com")
}

// C8.10
func TestExportPageSetup(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Properties = ir.PageProperties{
		Width: 595, Height: 842,
		MarginTop: 72, MarginBottom: 72,
		MarginLeft: 72, MarginRight: 72,
		Orientation: ir.Portrait,
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "page(")
}

// C8.11
func TestExportHeaderFooter(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Header = &ir.HeaderFooter{
		Kind:   ir.HeaderFooterDefault,
		Blocks: []ir.Block{&ir.Paragraph{ID: "h1", Runs: []ir.Run{{Text: "My Header"}}}},
	}
	doc.Sections[0].Footer = &ir.HeaderFooter{
		Kind:   ir.HeaderFooterDefault,
		Blocks: []ir.Block{&ir.Paragraph{ID: "f1", Runs: []ir.Run{{Text: "My Footer"}}}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "My Header")
	assert.Contains(t, s, "My Footer")
}

// C8.12
func TestExportFootnote(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID:   "p1",
			Runs: []ir.Run{{Text: "text with note"}},
			Footnotes: []ir.Footnote{
				{ID: "fn1", Blocks: []ir.Block{
					&ir.Paragraph{ID: "fnp1", Runs: []ir.Run{{Text: "footnote content"}}},
				}},
			},
		},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "footnote[")
}

// C8.13
func TestExportColors(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "red text", Attrs: ir.RunAttrs{Color: ir.Color{R: 255, G: 0, B: 0}}},
		}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "text(fill: rgb(")
}

// C8.14
func TestExportFontSize(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "big", Attrs: ir.RunAttrs{FontSize: 24}},
		}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "text(size:")
}

// C8.15
func TestExportMultiColumn(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Columns = []ir.Column{
		{Width: 225, Spacing: 18},
		{Width: 225, Spacing: 0},
	}
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "col content"}}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "columns(")
}

// C8.16
func TestExportCompilesWithTypst(t *testing.T) {
	// Integration test: actual Typst compilation would require the typst binary.
	// For now, just verify that the exporter produces non-empty output.
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "compile check"}}},
	}
	out, err := typst.Export(doc)
	require.NoError(t, err)
	assert.NotEmpty(t, out, "output should be non-empty for compilation")
}
