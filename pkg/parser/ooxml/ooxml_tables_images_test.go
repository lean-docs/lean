package ooxml

import (
	"archive/zip"
	"bytes"
	"io"
	"testing"

	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Cluster 6 — OOXML Parser (Tables)
// ---------------------------------------------------------------------------

// C6.1 TestOOXMLSimpleTable - 2×2 table: correct row/cell count
func TestOOXMLSimpleTable(t *testing.T) {
	body := `<w:tbl>
		<w:tr><w:tc>` + runPara("a") + `</w:tc><w:tc>` + runPara("b") + `</w:tc></w:tr>
		<w:tr><w:tc>` + runPara("c") + `</w:tc><w:tc>` + runPara("d") + `</w:tc></w:tr>
	</w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	table := firstTable(t, doc)
	assert.Len(t, table.Rows, 2)
	for _, row := range table.Rows {
		assert.Len(t, row.Cells, 2)
	}
}

// C6.2 TestOOXMLTableCellText - cell text in paragraph blocks
func TestOOXMLTableCellText(t *testing.T) {
	body := `<w:tbl><w:tr><w:tc>` + runPara("Hello") + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	table := firstTable(t, doc)
	cell := table.Rows[0].Cells[0]
	require.NotEmpty(t, cell.Blocks)
	para, ok := cell.Blocks[0].(*ir.Paragraph)
	require.True(t, ok)
	require.NotEmpty(t, para.Runs)
	assert.Equal(t, "Hello", para.Runs[0].Text)
}

// C6.3 TestOOXMLColumnWidths - gridCol widths from twips
func TestOOXMLColumnWidths(t *testing.T) {
	body := `<w:tbl>
		<w:tblGrid><w:gridCol w:w="4320"/><w:gridCol w:w="5760"/></w:tblGrid>
		<w:tr><w:tc>` + runPara("a") + `</w:tc><w:tc>` + runPara("b") + `</w:tc></w:tr>
	</w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	table := firstTable(t, doc)
	require.Len(t, table.ColumnWidths, 2)
	assert.InDelta(t, 216.0, table.ColumnWidths[0], 0.1)
	assert.InDelta(t, 288.0, table.ColumnWidths[1], 0.1)
}

// C6.4 TestOOXMLCellBorders - border style, size, color per cell
func TestOOXMLCellBorders(t *testing.T) {
	borders := `<w:tcBorders>
		<w:top w:val="single" w:sz="4" w:color="FF0000"/>
		<w:bottom w:val="single" w:sz="4" w:color="FF0000"/>
		<w:left w:val="single" w:sz="4" w:color="FF0000"/>
		<w:right w:val="single" w:sz="4" w:color="FF0000"/>
	</w:tcBorders>`
	body := `<w:tbl><w:tr><w:tc><w:tcPr>` + borders + `</w:tcPr>` + runPara("x") + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	cell := firstTable(t, doc).Rows[0].Cells[0]
	assert.Equal(t, "single", cell.Borders.Top.Style)
	assert.InDelta(t, 0.5, cell.Borders.Top.Width, 0.01)
	assert.Equal(t, ir.Color{R: 0xFF}, cell.Borders.Top.Color)
	assert.Equal(t, "single", cell.Borders.Bottom.Style)
	assert.Equal(t, "single", cell.Borders.Left.Style)
	assert.Equal(t, "single", cell.Borders.Right.Style)
}

// C6.5 TestOOXMLTableBorders - table-level borders when cell absent
func TestOOXMLTableBorders(t *testing.T) {
	tblBorders := `<w:tblBorders>
		<w:top w:val="single"/><w:bottom w:val="single"/>
		<w:left w:val="single"/><w:right w:val="single"/>
		<w:insideH w:val="single"/><w:insideV w:val="single"/>
	</w:tblBorders>`
	body := `<w:tbl><w:tblPr>` + tblBorders + `</w:tblPr><w:tr><w:tc>` + runPara("x") + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
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
	body := `<w:tbl><w:tr><w:tc><w:tcPr><w:shd w:fill="FFFF00"/></w:tcPr>` + runPara("x") + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	cell := firstTable(t, doc).Rows[0].Cells[0]
	assert.Equal(t, ir.Color{R: 0xFF, G: 0xFF}, cell.Shading)
}

// C6.7 TestOOXMLCellPadding - cell margin/padding
func TestOOXMLCellPadding(t *testing.T) {
	mar := `<w:tcMar><w:top w:w="72"/><w:bottom w:w="72"/><w:left w:w="108"/><w:right w:w="108"/></w:tcMar>`
	body := `<w:tbl><w:tr><w:tc><w:tcPr>` + mar + `</w:tcPr>` + runPara("x") + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	cell := firstTable(t, doc).Rows[0].Cells[0]
	assert.InDelta(t, 3.6, cell.Padding.Top, 0.1)
	assert.InDelta(t, 3.6, cell.Padding.Bottom, 0.1)
	assert.InDelta(t, 5.4, cell.Padding.Left, 0.1)
	assert.InDelta(t, 5.4, cell.Padding.Right, 0.1)
}

// C6.8 TestOOXMLColSpan - gridSpan sets ColSpan
func TestOOXMLColSpan(t *testing.T) {
	body := `<w:tbl><w:tr><w:tc><w:tcPr><w:gridSpan w:val="3"/></w:tcPr>` + runPara("x") + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	assert.Equal(t, 3, firstTable(t, doc).Rows[0].Cells[0].ColSpan)
}

// C6.9 TestOOXMLRowSpan - vMerge sets RowSpan
func TestOOXMLRowSpan(t *testing.T) {
	body := `<w:tbl>
		<w:tr><w:tc><w:tcPr><w:vMerge w:val="restart"/></w:tcPr>` + runPara("x") + `</w:tc></w:tr>
		<w:tr><w:tc><w:tcPr><w:vMerge/></w:tcPr>` + runPara("") + `</w:tc></w:tr>
	</w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	table := firstTable(t, doc)
	require.GreaterOrEqual(t, len(table.Rows), 2)
	assert.Equal(t, 2, table.Rows[0].Cells[0].RowSpan)
}

// C6.10 TestOOXMLHeaderRow - tblHeader marks IsHeader
func TestOOXMLHeaderRow(t *testing.T) {
	body := `<w:tbl>
		<w:tr><w:trPr><w:tblHeader/></w:trPr><w:tc>` + runPara("h") + `</w:tc></w:tr>
		<w:tr><w:tc>` + runPara("b") + `</w:tc></w:tr>
	</w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	table := firstTable(t, doc)
	assert.True(t, table.Rows[0].IsHeader)
	assert.False(t, table.Rows[1].IsHeader)
}

// C6.11 TestOOXMLTableAlignment - table alignment
func TestOOXMLTableAlignment(t *testing.T) {
	body := `<w:tbl><w:tblPr><w:jc w:val="center"/></w:tblPr><w:tr><w:tc>` + runPara("x") + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	assert.Equal(t, ir.AlignCenter, firstTable(t, doc).Align)
}

// C6.12 TestOOXMLTableStyle - tblStyle reference
func TestOOXMLTableStyle(t *testing.T) {
	body := `<w:tbl><w:tblPr><w:tblStyle w:val="TableGrid"/></w:tblPr><w:tr><w:tc>` + runPara("x") + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	assert.Equal(t, "TableGrid", firstTable(t, doc).Style)
}

// C6.13 TestOOXMLNestedTable - table inside cell
func TestOOXMLNestedTable(t *testing.T) {
	inner := `<w:tbl><w:tr><w:tc>` + runPara("inner") + `</w:tc></w:tr></w:tbl>`
	body := `<w:tbl><w:tr><w:tc>` + runPara("outer") + inner + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	outer := firstTable(t, doc)
	cell := outer.Rows[0].Cells[0]
	var nested *ir.Table
	for _, b := range cell.Blocks {
		if tbl, ok := b.(*ir.Table); ok {
			nested = tbl
			break
		}
	}
	require.NotNil(t, nested)
	assert.NotEmpty(t, nested.Rows)
}

// C6.14 TestOOXMLVerticalAlign - vAlign maps to VAlignment
func TestOOXMLVerticalAlign(t *testing.T) {
	body := `<w:tbl><w:tr>
		<w:tc><w:tcPr><w:vAlign w:val="top"/></w:tcPr>` + runPara("a") + `</w:tc>
		<w:tc><w:tcPr><w:vAlign w:val="center"/></w:tcPr>` + runPara("b") + `</w:tc>
		<w:tc><w:tcPr><w:vAlign w:val="bottom"/></w:tcPr>` + runPara("c") + `</w:tc>
	</w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	cells := firstTable(t, doc).Rows[0].Cells
	assert.Equal(t, ir.VAlignTop, cells[0].VAlign)
	assert.Equal(t, ir.VAlignCenter, cells[1].VAlign)
	assert.Equal(t, ir.VAlignBottom, cells[2].VAlign)
}

// C6.15 TestOOXMLRowHeight - trHeight in points
func TestOOXMLRowHeight(t *testing.T) {
	body := `<w:tbl><w:tr><w:trPr><w:trHeight w:val="480"/></w:trPr><w:tc>` + runPara("x") + `</w:tc></w:tr></w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	row := firstTable(t, doc).Rows[0]
	require.NotNil(t, row.Height)
	assert.InDelta(t, 24.0, *row.Height, 0.1)
}

// C6.9b TestOOXMLRowSpanMixedColspan - vMerge with mixed gridSpan
func TestOOXMLRowSpanMixedColspan(t *testing.T) {
	// Row 0: [colspan=2 cell] [vMerge=restart cell]  → grid cols 0-1, 2
	// Row 1: [cell] [cell] [vMerge=continue cell]    → grid cols 0, 1, 2
	body := `<w:tbl>
		<w:tr>
			<w:tc><w:tcPr><w:gridSpan w:val="2"/></w:tcPr>` + runPara("wide") + `</w:tc>
			<w:tc><w:tcPr><w:vMerge w:val="restart"/></w:tcPr>` + runPara("origin") + `</w:tc>
		</w:tr>
		<w:tr>
			<w:tc>` + runPara("a") + `</w:tc>
			<w:tc>` + runPara("b") + `</w:tc>
			<w:tc><w:tcPr><w:vMerge/></w:tcPr>` + runPara("") + `</w:tc>
		</w:tr>
	</w:tbl>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	table := firstTable(t, doc)
	require.GreaterOrEqual(t, len(table.Rows), 2)
	// Origin cell is row 0 cell 1 (grid column 2) — should span 2 rows.
	assert.Equal(t, 2, table.Rows[0].Cells[1].RowSpan)
}

// runPara wraps plain text in a <w:p><w:r><w:t>... paragraph for fixtures.
func runPara(text string) string {
	return "<w:p><w:r><w:t>" + text + "</w:t></w:r></w:p>"
}

// ---------------------------------------------------------------------------
// Cluster 7 — OOXML Parser (Images)
// ---------------------------------------------------------------------------

// C7.1 TestOOXMLInlineImagePNG - inline PNG bytes and format
func TestOOXMLInlineImagePNG(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "shp-inline-shape-access.docx"))
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
	doc, err := Parse(pythonDocxFixture(t, "shp-inline-shape-access.docx"))
	require.NoError(t, err)

	img := firstImageWithFormat(t, doc, ir.ImageJPEG)
	assert.Equal(t, ir.ImageJPEG, img.Format)
	assert.NotEmpty(t, img.Data)
	// JPEG magic bytes (SOI marker)
	assert.True(t, len(img.Data) >= 2, "data too short for JPEG header")
	assert.Equal(t, byte(0xFF), img.Data[0])
	assert.Equal(t, byte(0xD8), img.Data[1])
}

// C7.3 TestOOXMLImageDimensions - EMUs to points
func TestOOXMLImageDimensions(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "shp-inline-shape-access.docx"))
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.InDelta(t, 140.0, img.Width, 0.1)
	assert.InDelta(t, 56.0, img.Height, 0.1)
}

// C7.4 TestOOXMLImageRelationship - rId resolves to media file
func TestOOXMLImageRelationship(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "shp-inline-shape-access.docx"))
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.NotEmpty(t, img.Data, "relationship should resolve to actual image bytes")
	assert.NotEmpty(t, img.ID)
}

// C7.5 TestOOXMLImageAltText - descr attribute
func TestOOXMLImageAltText(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "shp-inline-shape-access.docx"))
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.Equal(t, "Picture 1", img.Alt)
}

// C7.6 TestOOXMLFloatLeft - anchor with left wrap
func TestOOXMLFloatLeft(t *testing.T) {
	doc, err := Parse(libreOfficeFixture(t, "float-left.docx"))
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.Equal(t, ir.FloatLeft, img.Float)
}

// C7.7 TestOOXMLFloatRight - right wrap
func TestOOXMLFloatRight(t *testing.T) {
	doc, err := Parse(libreOfficeFixture(t, "float-right.docx"))
	require.NoError(t, err)

	img := firstImage(t, doc)
	assert.Equal(t, ir.FloatRight, img.Float)
}

// C7.8 TestOOXMLMissingImageGraceful - broken rId → placeholder, no panic
func TestOOXMLMissingImageGraceful(t *testing.T) {
	doc, err := Parse(withMissingImage(t, libreOfficeFixture(t, "float-left.docx")))
	require.NoError(t, err)

	imageCount := 0
	for _, sec := range doc.Sections {
		for _, b := range sec.Blocks {
			if _, ok := b.(*ir.Image); ok {
				imageCount++
			}
		}
	}
	assert.Zero(t, imageCount)
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

func withMissingImage(t *testing.T, source []byte) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(source), int64(len(source)))
	require.NoError(t, err)
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	for _, file := range reader.File {
		stream, openErr := file.Open()
		require.NoError(t, openErr)
		content, readErr := io.ReadAll(stream)
		require.NoError(t, readErr)
		require.NoError(t, stream.Close())
		if file.Name == "word/_rels/document.xml.rels" {
			content = bytes.ReplaceAll(content, []byte("media/image1.png"), []byte("media/missing.png"))
		}
		destination, createErr := writer.CreateHeader(&file.FileHeader)
		require.NoError(t, createErr)
		_, writeErr := destination.Write(content)
		require.NoError(t, writeErr)
	}
	require.NoError(t, writer.Close())
	return output.Bytes()
}

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

func firstImageWithFormat(t *testing.T, doc *ir.Document, format ir.ImageFormat) *ir.Image {
	t.Helper()
	for _, section := range doc.Sections {
		for _, block := range section.Blocks {
			if image, ok := block.(*ir.Image); ok && image.Format == format {
				return image
			}
		}
	}
	t.Fatalf("no image found with format %d", format)
	return nil
}
