package ooxml

import (
	"testing"

	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Cluster 6 — OOXML Parser (Tables)
// ---------------------------------------------------------------------------

// TODO: All tests in this file will need proper .docx fixture files once the
// parser is implemented. For now they call Parse(nil) and assert no error,
// which will fail against ErrNotImplemented — serving as a skeleton for TDD.

// C6.1 TestOOXMLSimpleTable - 2×2 table: correct row/cell count
func TestOOXMLSimpleTable(t *testing.T) {
	// TODO: Replace nil with bytes of a .docx containing a simple 2×2 table.
	doc, err := Parse(nil)
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.NotEmpty(t, doc.Sections)

	var table *ir.Table
	for _, b := range doc.Sections[0].Blocks {
		if tbl, ok := b.(*ir.Table); ok {
			table = tbl
			break
		}
	}
	require.NotNil(t, table, "expected a Table block")
	assert.Len(t, table.Rows, 2)
	for _, row := range table.Rows {
		assert.Len(t, row.Cells, 2)
	}
}

// C6.2 TestOOXMLTableCellText - cell text in paragraph blocks
func TestOOXMLTableCellText(t *testing.T) {
	// TODO: Fixture with known cell text "Hello" in first cell.
	doc, err := Parse(nil)
	require.NoError(t, err)
	require.NotNil(t, doc)

	table := firstTable(t, doc)
	require.NotNil(t, table)
	require.NotEmpty(t, table.Rows)
	require.NotEmpty(t, table.Rows[0].Cells)

	cell := table.Rows[0].Cells[0]
	require.NotEmpty(t, cell.Blocks, "cell should contain at least one block")

	para, ok := cell.Blocks[0].(*ir.Paragraph)
	require.True(t, ok, "first block in cell should be a Paragraph")
	require.NotEmpty(t, para.Runs)
	assert.Equal(t, "Hello", para.Runs[0].Text)
}

// C6.3 TestOOXMLColumnWidths - gridCol widths from twips
func TestOOXMLColumnWidths(t *testing.T) {
	// TODO: Fixture with two gridCol entries (e.g. 4320 twips = 216pt, 5760 twips = 288pt).
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	require.Len(t, table.ColumnWidths, 2)
	// 4320 twips = 4320/20 = 216 points
	assert.InDelta(t, 216.0, table.ColumnWidths[0], 0.1)
	// 5760 twips = 5760/20 = 288 points
	assert.InDelta(t, 288.0, table.ColumnWidths[1], 0.1)
}

// C6.4 TestOOXMLCellBorders - border style, size, color per cell
func TestOOXMLCellBorders(t *testing.T) {
	// TODO: Fixture with explicit cell borders (single, sz=4, color=FF0000).
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	cell := table.Rows[0].Cells[0]

	assert.Equal(t, "single", cell.Borders.Top.Style)
	assert.InDelta(t, 0.5, cell.Borders.Top.Width, 0.01) // sz=4 eighths of a point → 0.5pt
	assert.Equal(t, ir.Color{R: 0xFF, G: 0x00, B: 0x00}, cell.Borders.Top.Color)

	assert.Equal(t, "single", cell.Borders.Bottom.Style)
	assert.Equal(t, "single", cell.Borders.Left.Style)
	assert.Equal(t, "single", cell.Borders.Right.Style)
}

// C6.5 TestOOXMLTableBorders - table-level borders when cell absent
func TestOOXMLTableBorders(t *testing.T) {
	// TODO: Fixture with table-level borders but no cell-level borders.
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	assert.Equal(t, "single", table.Borders.Top.Style)
	assert.Equal(t, "single", table.Borders.Bottom.Style)
	assert.Equal(t, "single", table.Borders.Left.Style)
	assert.Equal(t, "single", table.Borders.Right.Style)
	assert.Equal(t, "single", table.Borders.InsideH.Style)
	assert.Equal(t, "single", table.Borders.InsideV.Style)
}

// C6.6 TestOOXMLCellShading - cell background color
func TestOOXMLCellShading(t *testing.T) {
	// TODO: Fixture with cell shading fill="FFFF00" (yellow).
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	cell := table.Rows[0].Cells[0]
	assert.Equal(t, ir.Color{R: 0xFF, G: 0xFF, B: 0x00}, cell.Shading)
}

// C6.7 TestOOXMLCellPadding - cell margin/padding
func TestOOXMLCellPadding(t *testing.T) {
	// TODO: Fixture with explicit tcMar (top=72 twips, bottom=72, left=108, right=108).
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	cell := table.Rows[0].Cells[0]

	// 72 twips = 72/20 = 3.6 points
	assert.InDelta(t, 3.6, cell.Padding.Top, 0.1)
	assert.InDelta(t, 3.6, cell.Padding.Bottom, 0.1)
	// 108 twips = 108/20 = 5.4 points
	assert.InDelta(t, 5.4, cell.Padding.Left, 0.1)
	assert.InDelta(t, 5.4, cell.Padding.Right, 0.1)
}

// C6.8 TestOOXMLColSpan - gridSpan sets ColSpan
func TestOOXMLColSpan(t *testing.T) {
	// TODO: Fixture with a cell having gridSpan val="3".
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	cell := table.Rows[0].Cells[0]
	assert.Equal(t, 3, cell.ColSpan)
}

// C6.9 TestOOXMLRowSpan - vMerge sets RowSpan
func TestOOXMLRowSpan(t *testing.T) {
	// TODO: Fixture with vMerge restart on row 0 and vMerge continue on row 1
	// (2-row span).
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	require.True(t, len(table.Rows) >= 2)
	// The origin cell should carry the full span count.
	assert.Equal(t, 2, table.Rows[0].Cells[0].RowSpan)
}

// C6.10 TestOOXMLHeaderRow - tblHeader marks IsHeader
func TestOOXMLHeaderRow(t *testing.T) {
	// TODO: Fixture with first row marked as tblHeader.
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	require.NotEmpty(t, table.Rows)
	assert.True(t, table.Rows[0].IsHeader, "first row should be a header row")
	if len(table.Rows) > 1 {
		assert.False(t, table.Rows[1].IsHeader, "second row should not be a header row")
	}
}

// C6.11 TestOOXMLTableAlignment - table alignment
func TestOOXMLTableAlignment(t *testing.T) {
	// TODO: Fixture with jc val="center" on tblPr.
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	assert.Equal(t, ir.AlignCenter, table.Align)
}

// C6.12 TestOOXMLTableStyle - tblStyle reference
func TestOOXMLTableStyle(t *testing.T) {
	// TODO: Fixture with tblStyle val="TableGrid".
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	assert.Equal(t, "TableGrid", table.Style)
}

// C6.13 TestOOXMLNestedTable - table inside cell
func TestOOXMLNestedTable(t *testing.T) {
	// TODO: Fixture with a table inside a cell of the outer table.
	doc, err := Parse(nil)
	require.NoError(t, err)

	outer := firstTable(t, doc)
	require.NotEmpty(t, outer.Rows)
	require.NotEmpty(t, outer.Rows[0].Cells)

	cell := outer.Rows[0].Cells[0]
	var nested *ir.Table
	for _, b := range cell.Blocks {
		if tbl, ok := b.(*ir.Table); ok {
			nested = tbl
			break
		}
	}
	require.NotNil(t, nested, "expected a nested table inside the first cell")
	assert.NotEmpty(t, nested.Rows, "nested table should have rows")
}

// C6.14 TestOOXMLVerticalAlign - vAlign maps to VAlignment
func TestOOXMLVerticalAlign(t *testing.T) {
	// TODO: Fixture with three cells having vAlign top, center, bottom.
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	require.NotEmpty(t, table.Rows)
	require.True(t, len(table.Rows[0].Cells) >= 3)

	assert.Equal(t, ir.VAlignTop, table.Rows[0].Cells[0].VAlign)
	assert.Equal(t, ir.VAlignCenter, table.Rows[0].Cells[1].VAlign)
	assert.Equal(t, ir.VAlignBottom, table.Rows[0].Cells[2].VAlign)
}

// C6.15 TestOOXMLRowHeight - trHeight in points
func TestOOXMLRowHeight(t *testing.T) {
	// TODO: Fixture with trHeight val="480" (480 twips = 24 points).
	doc, err := Parse(nil)
	require.NoError(t, err)

	table := firstTable(t, doc)
	require.NotEmpty(t, table.Rows)
	require.NotNil(t, table.Rows[0].Height, "row height should be set")
	assert.InDelta(t, 24.0, *table.Rows[0].Height, 0.1) // 480 twips / 20 = 24pt
}

// ---------------------------------------------------------------------------
// Cluster 7 — OOXML Parser (Images)
// ---------------------------------------------------------------------------

// C7.1 TestOOXMLInlineImagePNG - inline PNG bytes and format
func TestOOXMLInlineImagePNG(t *testing.T) {
	// TODO: Fixture with an inline drawing referencing a PNG in media/.
	doc, err := Parse(nil)
	require.NoError(t, err)
	require.NotNil(t, doc)

	img := firstImage(t, doc)
	assert.Equal(t, ir.ImagePNG, img.Format)
	assert.NotEmpty(t, img.Data, "image data should not be empty")
	// PNG magic bytes
	assert.True(t, len(img.Data) >= 4, "data too short for PNG header")
	assert.Equal(t, byte(0x89), img.Data[0])
	assert.Equal(t, byte('P'), img.Data[1])
	assert.Equal(t, byte('N'), img.Data[2])
	assert.Equal(t, byte('G'), img.Data[3])
}

// C7.2 TestOOXMLInlineImageJPEG - inline JPEG
func TestOOXMLInlineImageJPEG(t *testing.T) {
	// TODO: Fixture with an inline drawing referencing a JPEG in media/.
	doc, err := Parse(nil)
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.Equal(t, ir.ImageJPEG, img.Format)
	assert.NotEmpty(t, img.Data)
	// JPEG magic bytes (SOI marker)
	assert.True(t, len(img.Data) >= 2, "data too short for JPEG header")
	assert.Equal(t, byte(0xFF), img.Data[0])
	assert.Equal(t, byte(0xD8), img.Data[1])
}

// C7.3 TestOOXMLImageDimensions - EMUs to points
func TestOOXMLImageDimensions(t *testing.T) {
	// TODO: Fixture with cx="914400" cy="457200" (EMUs).
	// 914400 EMU = 1 inch = 72 points; 457200 EMU = 0.5 inch = 36 points.
	doc, err := Parse(nil)
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.InDelta(t, 72.0, img.Width, 0.1)
	assert.InDelta(t, 36.0, img.Height, 0.1)
}

// C7.4 TestOOXMLImageRelationship - rId resolves to media file
func TestOOXMLImageRelationship(t *testing.T) {
	// TODO: Fixture where blip r:embed="rId5" resolves via
	// word/_rels/document.xml.rels to word/media/image1.png.
	doc, err := Parse(nil)
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.NotEmpty(t, img.Data, "relationship should resolve to actual image bytes")
	assert.NotEmpty(t, img.ID)
}

// C7.5 TestOOXMLImageAltText - descr attribute
func TestOOXMLImageAltText(t *testing.T) {
	// TODO: Fixture with docPr descr="Company Logo".
	doc, err := Parse(nil)
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.Equal(t, "Company Logo", img.Alt)
}

// C7.6 TestOOXMLFloatLeft - anchor with left wrap
func TestOOXMLFloatLeft(t *testing.T) {
	// TODO: Fixture with wp:anchor and wrapSquare wrapText="left".
	doc, err := Parse(nil)
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.Equal(t, ir.FloatLeft, img.Float)
}

// C7.7 TestOOXMLFloatRight - right wrap
func TestOOXMLFloatRight(t *testing.T) {
	// TODO: Fixture with wp:anchor and wrapSquare wrapText="right".
	doc, err := Parse(nil)
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.Equal(t, ir.FloatRight, img.Float)
}

// C7.8 TestOOXMLMissingImageGraceful - broken rId → placeholder, no panic
func TestOOXMLMissingImageGraceful(t *testing.T) {
	// TODO: Fixture with blip r:embed="rIdBROKEN" that does not exist in rels.
	// The parser should not panic and should either skip the image or return a
	// placeholder with empty data.
	doc, err := Parse(nil)
	require.NoError(t, err)

	// If the parser emits an image block it should be safe to inspect.
	for _, sec := range doc.Sections {
		for _, b := range sec.Blocks {
			if img, ok := b.(*ir.Image); ok {
				// Placeholder image: may have empty data but must not be nil struct.
				assert.NotNil(t, img)
				// Accept either empty data (placeholder) or non-empty data.
				_ = img.Data
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// firstTable returns the first *ir.Table block found in the document.
func firstTable(t *testing.T, doc *ir.Document) *ir.Table {
	t.Helper()
	require.NotNil(t, doc)
	require.NotEmpty(t, doc.Sections)
	for _, sec := range doc.Sections {
		for _, b := range sec.Blocks {
			if tbl, ok := b.(*ir.Table); ok {
				return tbl
			}
		}
	}
	t.Fatal("no Table block found in document")
	return nil
}

// firstImage returns the first *ir.Image block found in the document.
func firstImage(t *testing.T, doc *ir.Document) *ir.Image {
	t.Helper()
	require.NotNil(t, doc)
	require.NotEmpty(t, doc.Sections)
	for _, sec := range doc.Sections {
		for _, b := range sec.Blocks {
			if img, ok := b.(*ir.Image); ok {
				return img
			}
		}
	}
	t.Fatal("no Image block found in document")
	return nil
}
