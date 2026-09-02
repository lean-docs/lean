package md_test

import (
	"testing"

	"github.com/lean-docs/lean/pkg/export/md"
	"github.com/lean-docs/lean/pkg/ir"
	mdparser "github.com/lean-docs/lean/pkg/parser/md"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C9.1
func TestExportPlainText(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello world"}}},
	}
	out, err := md.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "Hello world")
}

// C9.2
func TestExportBold(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "strong", Attrs: ir.RunAttrs{Bold: true}},
		}},
	}
	out, err := md.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "**strong**")
}

// C9.3
func TestExportItalic(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "emphasis", Attrs: ir.RunAttrs{Italic: true}},
		}},
	}
	out, err := md.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "*emphasis*")
}

// C9.4
func TestExportHeadings(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "h1", Style: "Heading1", Runs: []ir.Run{{Text: "H1"}}},
		&ir.Paragraph{ID: "h2", Style: "Heading2", Runs: []ir.Run{{Text: "H2"}}},
		&ir.Paragraph{ID: "h3", Style: "Heading3", Runs: []ir.Run{{Text: "H3"}}},
		&ir.Paragraph{ID: "h4", Style: "Heading4", Runs: []ir.Run{{Text: "H4"}}},
		&ir.Paragraph{ID: "h5", Style: "Heading5", Runs: []ir.Run{{Text: "H5"}}},
		&ir.Paragraph{ID: "h6", Style: "Heading6", Runs: []ir.Run{{Text: "H6"}}},
	}
	out, err := md.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "# H1")
	assert.Contains(t, s, "## H2")
	assert.Contains(t, s, "### H3")
	assert.Contains(t, s, "#### H4")
	assert.Contains(t, s, "##### H5")
	assert.Contains(t, s, "###### H6")
}

// C9.5
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
	out, err := md.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "- item one")
	assert.Contains(t, s, "- item two")
}

// C9.6
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
	out, err := md.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "1. first")
	assert.Contains(t, s, "2. second")
}

// C9.7
func TestExportTable(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Table{ID: "t1", Rows: []ir.TableRow{
			{IsHeader: true, Cells: []ir.TableCell{
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "h1", Runs: []ir.Run{{Text: "Name"}}},
				}},
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "h2", Runs: []ir.Run{{Text: "Value"}}},
				}},
			}},
			{Cells: []ir.TableCell{
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "c1", Runs: []ir.Run{{Text: "A"}}},
				}},
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "c2", Runs: []ir.Run{{Text: "1"}}},
				}},
			}},
		}},
	}
	out, err := md.Export(doc)
	require.NoError(t, err)
	s := string(out)
	// GFM table syntax
	assert.Contains(t, s, "| Name")
	assert.Contains(t, s, "| Value")
	assert.Contains(t, s, "---")
	assert.Contains(t, s, "| A")
}

// C9.8
func TestExportLink(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "click here", Hyperlink: &ir.Hyperlink{URL: "https://example.com"}},
		}},
	}
	out, err := md.Export(doc)
	require.NoError(t, err)
	assert.Contains(t, string(out), "[click here](https://example.com)")
}

// C9.9
func TestExportImage(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Image{
			ID: "img1", Data: []byte{0x89, 0x50, 0x4E, 0x47},
			Format: ir.ImagePNG, Width: 200, Height: 100, Alt: "photo",
		},
	}
	out, err := md.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "![photo]")
}

// C9.10
func TestExportCodeBlock(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Style: "Code", Runs: []ir.Run{{Text: "fmt.Println()"}}},
	}
	out, err := md.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "```")
	assert.Contains(t, s, "fmt.Println()")
}

// C9.11
func TestExportRoundTrip(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "h1", Style: "Heading1", Runs: []ir.Run{{Text: "Title"}}},
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "Normal "},
			{Text: "bold", Attrs: ir.RunAttrs{Bold: true}},
			{Text: " text"},
		}},
	}
	out, err := md.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "# Title")
	assert.Contains(t, s, "**bold**")
	reopened, err := mdparser.Parse(out)
	require.NoError(t, err)
	require.Len(t, reopened.Sections[0].Blocks, 2)
	heading := reopened.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "Heading1", heading.Style)
	assert.Equal(t, "Title", heading.Runs[0].Text)
	paragraph := reopened.Sections[0].Blocks[1].(*ir.Paragraph)
	require.Len(t, paragraph.Runs, 3)
	assert.Equal(t, "Normal ", paragraph.Runs[0].Text)
	assert.Equal(t, "bold", paragraph.Runs[1].Text)
	assert.True(t, paragraph.Runs[1].Attrs.Bold)
	assert.Equal(t, " text", paragraph.Runs[2].Text)
}
