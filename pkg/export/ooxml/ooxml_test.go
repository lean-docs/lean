package ooxml_test

import (
	"archive/zip"
	"bytes"
	"testing"

	"github.com/lean-docs/lean/pkg/export/ooxml"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C11.1
func TestExportMinimalDocx(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello"}}},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)
	assert.NotEmpty(t, out)

	// Should be a valid ZIP
	r, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	require.NoError(t, err)
	assert.NotEmpty(t, r.File, "ZIP must contain files")
}

// C11.2
func TestExportTextRoundTrip(t *testing.T) {
	// TODO: Full round-trip requires a DOCX parser (not yet implemented).
	// For now, verify that the exported ZIP contains document.xml with the text.
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "round trip text"}}},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	content := readZipEntry(t, out, "word/document.xml")
	assert.Contains(t, content, "round trip text")
}

// C11.3
func TestExportFormatting(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{
			{Text: "bold", Attrs: ir.RunAttrs{Bold: true}},
			{Text: "italic", Attrs: ir.RunAttrs{Italic: true}},
			{Text: "underlined", Attrs: ir.RunAttrs{Underline: ir.UnderlineSingle}},
			{Text: "struck", Attrs: ir.RunAttrs{Strike: true}},
		}},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	content := readZipEntry(t, out, "word/document.xml")
	assert.Contains(t, content, "<w:b/>")
	assert.Contains(t, content, "<w:i/>")
	assert.Contains(t, content, "<w:u ")
	assert.Contains(t, content, "<w:strike/>")
}

// C11.4
func TestExportList(t *testing.T) {
	doc := ir.NewDocument()
	doc.Styles.Numbering = []ir.NumberingDef{
		{ID: "num1", Levels: []ir.NumberingLevel{
			{Level: 0, Format: ir.NumFormatBullet, Text: "•", Start: 1},
		}},
	}
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "li1", Runs: []ir.Run{{Text: "item"}},
			Numbering: &ir.NumberingRef{ID: "num1", Level: 0}},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	content := readZipEntry(t, out, "word/document.xml")
	assert.Contains(t, content, "<w:numId")
	assert.Contains(t, content, "<w:ilvl")
}

// C11.5
func TestExportTable(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Table{ID: "t1", Rows: []ir.TableRow{
			{Cells: []ir.TableCell{
				{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
					&ir.Paragraph{ID: "c1", Runs: []ir.Run{{Text: "cell"}}},
				}},
			}},
		}},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	content := readZipEntry(t, out, "word/document.xml")
	assert.Contains(t, content, "<w:tbl>")
	assert.Contains(t, content, "<w:tr>")
	assert.Contains(t, content, "<w:tc>")
}

// C11.6
func TestExportImage(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Image{
			ID: "img1", Data: []byte{0x89, 0x50, 0x4E, 0x47},
			Format: ir.ImagePNG, Width: 200, Height: 100, Alt: "photo",
		},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	content := readZipEntry(t, out, "word/document.xml")
	assert.Contains(t, content, "<w:drawing>")

	// Image data should be embedded in the ZIP
	r, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	require.NoError(t, err)
	found := false
	for _, f := range r.File {
		if f.Name == "word/media/image1.png" || f.Name == "word/media/img1.png" {
			found = true
			break
		}
	}
	assert.True(t, found, "ZIP should contain the image file in word/media/")
}

// C11.7
func TestExportTabStops(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Para: ir.ParaAttrs{
				TabStops: []ir.TabStop{
					{Position: 72, Alignment: ir.TabAlignLeft, Leader: ir.TabLeaderNone},
					{Position: 288, Alignment: ir.TabAlignRight, Leader: ir.TabLeaderDot},
				},
			},
			Runs: []ir.Run{{Text: "tabbed"}},
		},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	content := readZipEntry(t, out, "word/document.xml")
	assert.Contains(t, content, "<w:tabs>")
	assert.Contains(t, content, "<w:tab ")
}

// C11.8
func TestExportHeaderFooter(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Header = &ir.HeaderFooter{
		Kind:   ir.HeaderFooterDefault,
		Blocks: []ir.Block{&ir.Paragraph{ID: "h1", Runs: []ir.Run{{Text: "Header"}}}},
	}
	doc.Sections[0].Footer = &ir.HeaderFooter{
		Kind:   ir.HeaderFooterDefault,
		Blocks: []ir.Block{&ir.Paragraph{ID: "f1", Runs: []ir.Run{{Text: "Footer"}}}},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	r, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	require.NoError(t, err)

	hasHeader := false
	hasFooter := false
	for _, f := range r.File {
		if f.Name == "word/header1.xml" {
			hasHeader = true
		}
		if f.Name == "word/footer1.xml" {
			hasFooter = true
		}
	}
	assert.True(t, hasHeader, "ZIP should contain header1.xml")
	assert.True(t, hasFooter, "ZIP should contain footer1.xml")
}

// C11.9
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
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	r, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	require.NoError(t, err)
	hasFootnotes := false
	for _, f := range r.File {
		if f.Name == "word/footnotes.xml" {
			hasFootnotes = true
			break
		}
	}
	assert.True(t, hasFootnotes, "ZIP should contain footnotes.xml")
}

// C11.10
func TestExportValidZIP(t *testing.T) {
	doc := ir.NewDocument()
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	r, err := zip.NewReader(bytes.NewReader(out), int64(len(out)))
	require.NoError(t, err)
	assert.NotEmpty(t, r.File, "ZIP must contain files")

	// Verify essential OOXML parts exist
	names := make(map[string]bool)
	for _, f := range r.File {
		names[f.Name] = true
	}
	assert.True(t, names["word/document.xml"], "must contain word/document.xml")
}

// C11.11
func TestExportContentTypes(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello"}}},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	content := readZipEntry(t, out, "[Content_Types].xml")
	assert.Contains(t, content, "ContentType")
	assert.Contains(t, content, "application/vnd.openxmlformats")
}

// C11.12
func TestExportRelationships(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello"}}},
	}
	out, err := ooxml.Export(doc)
	require.NoError(t, err)

	content := readZipEntry(t, out, "_rels/.rels")
	assert.Contains(t, content, "Relationship")
	assert.Contains(t, content, "officeDocument")
}

// readZipEntry reads a named file from a ZIP archive stored in data.
func readZipEntry(t *testing.T, data []byte, name string) string {
	t.Helper()
	r, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	require.NoError(t, err)
	for _, f := range r.File {
		if f.Name == name {
			rc, err := f.Open()
			require.NoError(t, err)
			defer rc.Close()
			var buf bytes.Buffer
			_, err = buf.ReadFrom(rc)
			require.NoError(t, err)
			return buf.String()
		}
	}
	t.Fatalf("ZIP entry %q not found", name)
	return ""
}
