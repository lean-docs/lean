package html_test

import (
	"testing"

	"github.com/lean-docs/lean/pkg/export/html"
	"github.com/lean-docs/lean/pkg/ir"
	htmlparser "github.com/lean-docs/lean/pkg/parser/html"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C10.1
func TestExportValidHTML5(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello"}}},
	}
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "<!DOCTYPE html>")
	assert.Contains(t, s, "<html")
	assert.Contains(t, s, "</html>")
	assert.Contains(t, s, "<head>")
	assert.Contains(t, s, "<body>")
}

// C10.2
func TestExportPlainText(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Simple text"}}},
	}
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "<p>")
	assert.Contains(t, s, "Simple text")
}

// C10.3
func TestExportFormatting(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "bold", Attrs: ir.RunAttrs{Bold: true}},
			{Text: "italic", Attrs: ir.RunAttrs{Italic: true}},
			{Text: "underlined", Attrs: ir.RunAttrs{Underline: ir.UnderlineSingle}},
		}},
	}
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "<strong>bold</strong>")
	assert.Contains(t, s, "<em>italic</em>")
	assert.Contains(t, s, "<u>underlined</u>")
}

// C10.4
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
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "<h1>H1</h1>")
	assert.Contains(t, s, "<h2>H2</h2>")
	assert.Contains(t, s, "<h3>H3</h3>")
	assert.Contains(t, s, "<h4>H4</h4>")
	assert.Contains(t, s, "<h5>H5</h5>")
	assert.Contains(t, s, "<h6>H6</h6>")
}

// C10.5
func TestExportLists(t *testing.T) {
	doc := ir.NewDocument()
	doc.Styles.Numbering = []ir.NumberingDef{
		{ID: "bullets", Levels: []ir.NumberingLevel{
			{Level: 0, Format: ir.NumFormatBullet, Text: "•", Start: 1},
		}},
		{ID: "ordered", Levels: []ir.NumberingLevel{
			{Level: 0, Format: ir.NumFormatDecimal, Text: "%1.", Start: 1},
		}},
	}
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "li1", Runs: []ir.Run{{Text: "bullet item"}},
			Numbering: &ir.NumberingRef{ID: "bullets", Level: 0}},
		&ir.Paragraph{ID: "li2", Runs: []ir.Run{{Text: "numbered item"}},
			Numbering: &ir.NumberingRef{ID: "ordered", Level: 0}},
	}
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "<ul>")
	assert.Contains(t, s, "</ul>")
	assert.Contains(t, s, "<ol>")
	assert.Contains(t, s, "</ol>")
	assert.Contains(t, s, "<li>")
}

// C10.6
func TestExportTable(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Table{ID: "t1", Rows: []ir.TableRow{
			{IsHeader: true, Cells: []ir.TableCell{
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "th1", Runs: []ir.Run{{Text: "Name"}}},
				}},
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "th2", Runs: []ir.Run{{Text: "Age"}}},
				}},
			}},
			{Cells: []ir.TableCell{
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "c1", Runs: []ir.Run{{Text: "Alice"}}},
				}},
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "c2", Runs: []ir.Run{{Text: "30"}}},
				}},
			}},
		}},
	}
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "<table>")
	assert.Contains(t, s, "<thead>")
	assert.Contains(t, s, "<tbody>")
	assert.Contains(t, s, "<th>")
	assert.Contains(t, s, "<td>")
	assert.Contains(t, s, "Alice")
}

// C10.7
func TestExportImage(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Image{
			ID: "img1", Data: []byte{0x89, 0x50, 0x4E, 0x47},
			Format: ir.ImagePNG, Width: 200, Height: 100, Alt: "photo",
		},
	}
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "<img")
	assert.Contains(t, s, `alt="photo"`)
}

// C10.8
func TestExportHyperlink(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "click here", Hyperlink: &ir.Hyperlink{URL: "https://example.com"}},
		}},
	}
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, `<a href="https://example.com"`)
	assert.Contains(t, s, "click here")
	assert.Contains(t, s, "</a>")
}

// C10.9
func TestExportRoundTrip(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "h1", Style: "Heading1", Runs: []ir.Run{{Text: "Title"}}},
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "Normal "},
			{Text: "bold", Attrs: ir.RunAttrs{Bold: true}},
		}},
	}
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, "<h1>Title</h1>")
	assert.Contains(t, s, "<strong>bold</strong>")
	reopened, err := htmlparser.Parse(out)
	require.NoError(t, err)
	require.Len(t, reopened.Sections[0].Blocks, 2)
	heading := reopened.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "Heading1", heading.Style)
	assert.Equal(t, "Title", heading.Runs[0].Text)
	paragraph := reopened.Sections[0].Blocks[1].(*ir.Paragraph)
	require.Len(t, paragraph.Runs, 2)
	assert.Equal(t, "Normal ", paragraph.Runs[0].Text)
	assert.Equal(t, "bold", paragraph.Runs[1].Text)
	assert.True(t, paragraph.Runs[1].Attrs.Bold)
}

// C10.10
func TestExportAccessibility(t *testing.T) {
	doc := ir.NewDocument()
	doc.Meta.Language = "en"
	doc.Meta.Title = "Accessible Document"
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "content"}}},
	}
	out, err := html.Export(doc)
	require.NoError(t, err)
	s := string(out)
	assert.Contains(t, s, `lang="en"`)
	assert.Contains(t, s, "<title>Accessible Document</title>")
}
