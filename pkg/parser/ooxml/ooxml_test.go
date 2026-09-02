package ooxml

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Cluster 4 — OOXML (.docx) Parsing
// ---------------------------------------------------------------------------

// C4.1 — Minimal .docx round-trip
func TestOOXMLMinimalDocx(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML("", "hello"))))
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.Equal(t, ir.IRVersion, doc.Meta.IRVersion)
	assert.GreaterOrEqual(t, len(doc.Sections), 1)
}

// C4.2 — Single paragraph with plain text
func TestOOXMLSingleParagraphPlainText(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML("", "hello world"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.GreaterOrEqual(t, len(p.Runs), 1)
	assert.Equal(t, "hello world", p.Runs[0].Text)
}

// C4.3 — Bold run
func TestOOXMLBoldRun(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:b/>`, "bold"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.GreaterOrEqual(t, len(p.Runs), 1)
	assert.True(t, p.Runs[0].Attrs.Bold)
}

// C4.4 — Italic run
func TestOOXMLItalicRun(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:i/>`, "it"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.True(t, p.Runs[0].Attrs.Italic)
}

// C4.5 — Underline run (single)
func TestOOXMLUnderlineRun(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:u w:val="single"/>`, "u"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, ir.UnderlineSingle, p.Runs[0].Attrs.Underline)
}

// C4.6 — Strikethrough run
func TestOOXMLStrikethroughRun(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:strike/>`, "s"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.True(t, p.Runs[0].Attrs.Strike)
}

// C4.7 — Superscript run
func TestOOXMLSuperscriptRun(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:vertAlign w:val="superscript"/>`, "x"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, ir.BaselineSuperscript, p.Runs[0].Attrs.Baseline)
}

// C4.8 — Subscript run
func TestOOXMLSubscriptRun(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:vertAlign w:val="subscript"/>`, "x"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, ir.BaselineSubscript, p.Runs[0].Attrs.Baseline)
}

// C4.9 — SmallCaps run
func TestOOXMLSmallCapsRun(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:smallCaps/>`, "s"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.True(t, p.Runs[0].Attrs.SmallCaps)
}

// C4.10 — AllCaps run
func TestOOXMLAllCapsRun(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:caps/>`, "c"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.True(t, p.Runs[0].Attrs.AllCaps)
}

// C4.11 — Font size
func TestOOXMLFontSize(t *testing.T) {
	// sz is in half-points: 24 → 12pt
	doc, err := Parse(buildDocx(t, para(runXML(`<w:sz w:val="24"/>`, "x"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, float64(12), p.Runs[0].Attrs.FontSize)
}

// C4.12 — Font name
func TestOOXMLFontName(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:rFonts w:ascii="Arial"/>`, "x"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, "Arial", p.Runs[0].Attrs.FontName)
}

// C4.13 — Run color
func TestOOXMLRunColor(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:color w:val="FF0000"/>`, "x"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, ir.Color{R: 255, G: 0, B: 0}, p.Runs[0].Attrs.Color)
}

// C4.14 — Highlight color
func TestOOXMLHighlightColor(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(runXML(`<w:highlight w:val="yellow"/>`, "x"))))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, ir.Color{R: 255, G: 255, B: 0}, p.Runs[0].Attrs.Highlight)
}

// C4.15 — Paragraph alignment (center)
func TestOOXMLParagraphAlignCenter(t *testing.T) {
	body := `<w:p><w:pPr><w:jc w:val="center"/></w:pPr>` + runXML("", "x") + `</w:p>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, ir.AlignCenter, p.Para.Align)
}

// C4.16 — Paragraph spacing (before/after)
func TestOOXMLParagraphSpacing(t *testing.T) {
	body := `<w:p><w:pPr><w:spacing w:before="240" w:after="120" w:line="240" w:lineRule="auto"/></w:pPr>` + runXML("", "x") + `</w:p>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, float64(12), p.Para.Spacing.Before) // 240 twips = 12 pt
	assert.Equal(t, float64(6), p.Para.Spacing.After)
	assert.Equal(t, ir.LineRuleAuto, p.Para.Spacing.LineRule)
}

// C4.17 — Paragraph indent (left, first-line)
func TestOOXMLParagraphIndent(t *testing.T) {
	body := `<w:p><w:pPr><w:ind w:left="720" w:firstLine="240"/></w:pPr>` + runXML("", "x") + `</w:p>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, float64(36), p.Para.Indent.Left) // 720 twips = 36 pt
	assert.Equal(t, float64(12), p.Para.Indent.FirstLine)
}

// C4.18 — Tab stops
func TestOOXMLTabStops(t *testing.T) {
	body := `<w:p><w:pPr><w:tabs><w:tab w:val="left" w:pos="2160" w:leader="dot"/></w:tabs></w:pPr>` + runXML("", "x") + `</w:p>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.Len(t, p.Para.TabStops, 1)
	assert.Equal(t, float64(108), p.Para.TabStops[0].Position) // 2160 twips = 108 pt
	assert.Equal(t, ir.TabLeaderDot, p.Para.TabStops[0].Leader)
}

// C4.19 — KeepTogether / KeepWithNext
func TestOOXMLKeepTogetherKeepWithNext(t *testing.T) {
	body := `<w:p><w:pPr><w:keepLines/><w:keepNext/></w:pPr>` + runXML("", "x") + `</w:p>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.True(t, p.Para.KeepTogether)
	assert.True(t, p.Para.KeepWithNext)
}

// C4.20 — PageBreakBefore
func TestOOXMLPageBreakBefore(t *testing.T) {
	body := `<w:p><w:pPr><w:pageBreakBefore/></w:pPr>` + runXML("", "x") + `</w:p>`
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.True(t, p.Para.PageBreakBefore)
}

// C4.21 — Named styles parsed into StyleSheet
func TestOOXMLNamedStyles(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "par-known-styles.docx"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.NotEmpty(t, doc.Styles.Named, "should have at least one named style")
}

// C4.22 — Paragraph references a named style
func TestOOXMLParagraphStyleRef(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "par-known-styles.docx"))
	require.NoError(t, err)
	var found bool
	for _, block := range doc.Sections[0].Blocks {
		if paragraph, ok := block.(*ir.Paragraph); ok && paragraph.Style == "Heading1" {
			found = true
		}
	}
	assert.True(t, found, "paragraph should reference Heading1")
}

// C4.23 — Hyperlink with URL
func TestOOXMLHyperlinkURL(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "par-hyperlinks.docx"))
	require.NoError(t, err)
	var found bool
	for _, block := range doc.Sections[0].Blocks {
		if paragraph, ok := block.(*ir.Paragraph); ok {
			for _, run := range paragraph.Runs {
				if run.Hyperlink != nil && run.Hyperlink.URL != "" {
					found = true
				}
			}
		}
	}
	assert.True(t, found, "should find a run with a hyperlink URL")
}

// C4.24 — Internal bookmark hyperlink
func TestOOXMLHyperlinkBookmark(t *testing.T) {
	doc, err := Parse(libreOfficeFixture(t, "hyperlink.docx"))
	require.NoError(t, err)
	var found bool
	for _, section := range doc.Sections {
		for _, block := range section.Blocks {
			paragraph, ok := block.(*ir.Paragraph)
			if !ok {
				continue
			}
			for _, run := range paragraph.Runs {
				if run.Hyperlink != nil && run.Hyperlink.Bookmark != "" {
					found = true
				}
			}
		}
	}
	assert.True(t, found, "should find a run with a bookmark hyperlink")
}

// C4.25 — Line break
func TestOOXMLLineBreak(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(`<w:r><w:br/></w:r>`)))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.GreaterOrEqual(t, len(p.Runs), 1)
	assert.Equal(t, ir.BreakLine, p.Runs[0].Break)
}

// C4.26 — Page break (inline)
func TestOOXMLPageBreakInline(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(`<w:r><w:br w:type="page"/></w:r>`)))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.GreaterOrEqual(t, len(p.Runs), 1)
	assert.Equal(t, ir.BreakPage, p.Runs[0].Break)
}

// C4.27 — Column break
func TestOOXMLColumnBreak(t *testing.T) {
	doc, err := Parse(buildDocx(t, para(`<w:r><w:br w:type="column"/></w:r>`)))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.GreaterOrEqual(t, len(p.Runs), 1)
	assert.Equal(t, ir.BreakColumn, p.Runs[0].Break)
}

// C4.28 — Bullet list (numbering reference)
func TestOOXMLBulletList(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "num-having-numbering-part.docx"))
	require.NoError(t, err)
	var found bool
	for _, nd := range doc.Styles.Numbering {
		if len(nd.Levels) > 0 && nd.Levels[0].Format == ir.NumFormatBullet {
			found = true
			break
		}
	}
	assert.True(t, found, "numbering def should exist for bullet list")
}

// C4.29 — Numbered list (decimal)
func TestOOXMLNumberedList(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "num-having-numbering-part.docx"))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.NotNil(t, p.Numbering)
	var found bool
	for _, nd := range doc.Styles.Numbering {
		if nd.ID == p.Numbering.ID {
			require.GreaterOrEqual(t, len(nd.Levels), 1)
			assert.Equal(t, ir.NumFormatDecimal, nd.Levels[0].Format)
			found = true
			break
		}
	}
	assert.True(t, found, "numbering def should exist for numbered list")
}

// C4.30 — Multi-level numbering
func TestOOXMLMultiLevelNumbering(t *testing.T) {
	doc, err := Parse(libreOfficeFixture(t, "mixednumberings.docx"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	// Expect at least one numbering def with multiple levels
	var found bool
	for _, nd := range doc.Styles.Numbering {
		if len(nd.Levels) >= 2 {
			found = true
			break
		}
	}
	assert.True(t, found, "should have a numbering def with at least 2 levels")
}

// Note: Table tests (C6.x) are in ooxml_tables_images_test.go
// Note: Image tests (C7.x) are in ooxml_tables_images_test.go

// C4.36 — Footnote
func TestOOXMLFootnote(t *testing.T) {
	doc, err := Parse(libreOfficeFixture(t, "footnote.docx"))
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.GreaterOrEqual(t, len(p.Footnotes), 1)
	assert.NotEmpty(t, p.Footnotes[0].ID)
	assert.GreaterOrEqual(t, len(p.Footnotes[0].Blocks), 1)
}

// C4.37 — Header content
func TestOOXMLHeaderContent(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "hdr-header-footer.docx"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.NotNil(t, doc.Sections[0].Header)
	assert.GreaterOrEqual(t, len(doc.Sections[0].Header.Blocks), 1)
}

// C4.38 — Footer content
func TestOOXMLFooterContent(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "hdr-header-footer.docx"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.NotNil(t, doc.Sections[0].Footer)
	assert.GreaterOrEqual(t, len(doc.Sections[0].Footer.Blocks), 1)
}

// C4.39 — Page properties (dimensions, margins, orientation)
func TestOOXMLPageProperties(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "sct-section-props.docx"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	pp := doc.Sections[0].Properties
	assert.Greater(t, pp.Width, float64(0))
	assert.Greater(t, pp.Height, float64(0))
	assert.Greater(t, pp.MarginTop, float64(0))
	assert.Greater(t, pp.MarginBottom, float64(0))
	assert.Greater(t, pp.MarginLeft, float64(0))
	assert.Greater(t, pp.MarginRight, float64(0))
}

// C4.40 — Landscape orientation
func TestOOXMLLandscapeOrientation(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "sct-section-props.docx"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	assert.Equal(t, ir.Landscape, doc.Sections[0].Properties.Orientation)
}

// C4.41 — Document metadata (title, author)
func TestOOXMLDocumentMetadata(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "doc-coreprops.docx"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.NotEmpty(t, doc.Meta.Title)
	assert.NotEmpty(t, doc.Meta.Author)
}

// C4.42 — Bookmark block
func TestOOXMLBookmark(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "num-having-numbering-part.docx"))
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	var bm *ir.Bookmark
	for _, b := range doc.Sections[0].Blocks {
		if bk, ok := b.(*ir.Bookmark); ok {
			bm = bk
			break
		}
	}
	require.NotNil(t, bm, "should find a bookmark block")
	assert.NotEmpty(t, bm.Name)
}

// ---------------------------------------------------------------------------
// Cluster 5 — OOXML Edge Cases & Error Handling
// ---------------------------------------------------------------------------

// C5.1 — Nil input returns error
func TestOOXMLNilInput(t *testing.T) {
	doc, err := Parse(nil)
	assert.ErrorIs(t, err, ErrEmptyInput)
	assert.Nil(t, doc)
}

// C5.2 — Empty byte slice returns error
func TestOOXMLEmptyInput(t *testing.T) {
	doc, err := Parse([]byte{})
	assert.Error(t, err)
	assert.Nil(t, doc)
}

// C5.3 — Invalid ZIP data returns error
func TestOOXMLInvalidZIP(t *testing.T) {
	doc, err := Parse([]byte("this is not a zip file"))
	assert.Error(t, err)
	assert.Nil(t, doc)
}

// C5.4 — Valid ZIP but missing [Content_Types].xml
func TestOOXMLMissingContentTypes(t *testing.T) {
	// TODO: fixture — valid ZIP without [Content_Types].xml
	doc, err := Parse([]byte{0x50, 0x4b, 0x05, 0x06, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.Error(t, err)
	assert.Nil(t, doc)
}

// C5.5 — Valid ZIP but missing word/document.xml
func TestOOXMLMissingDocumentXML(t *testing.T) {
	// TODO: fixture — valid OOXML ZIP without word/document.xml
	doc, err := Parse([]byte{0x50, 0x4b, 0x05, 0x06, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0})
	assert.Error(t, err)
	assert.Nil(t, doc)
}

// C5.6 — Corrupt XML inside otherwise valid ZIP
func TestOOXMLCorruptXML(t *testing.T) {
	// TODO: fixture — valid ZIP with corrupt XML in word/document.xml
	doc, err := Parse([]byte("corrupt-xml-fixture"))
	assert.Error(t, err)
	assert.Nil(t, doc)
}

// C5.7 — Unknown element in body is gracefully ignored
func TestOOXMLUnknownElementIgnored(t *testing.T) {
	doc, err := Parse(pythonDocxFixture(t, "blk-paras-and-tables.docx"))
	require.NoError(t, err)
	require.NotNil(t, doc)
	// The unknown element should be silently skipped;
	// the rest of the document should parse correctly.
	assert.GreaterOrEqual(t, len(doc.Sections), 1)
}

func pythonDocxFixture(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "fixtures", "ooxml", "python-docx", name))
	require.NoError(t, err)
	return content
}

func libreOfficeFixture(t *testing.T, name string) []byte {
	t.Helper()
	content, err := os.ReadFile(filepath.Join("../../../testdata/fixtures/ooxml/libreoffice", name))
	require.NoError(t, err)
	return content
}

// C5.8 — Large document does not panic
func TestOOXMLLargeDocumentNoPanic(t *testing.T) {
	var runs []string
	for i := 0; i < 5000; i++ {
		runs = append(runs, runXML("", "x"))
	}
	large := buildDocx(t, para(runs...))
	assert.NotPanics(t, func() {
		_, _ = Parse(large)
	})
}

// C5.9 — Paragraph ID uniqueness
func TestOOXMLParagraphIDUniqueness(t *testing.T) {
	body := para(runXML("", "a")) + para(runXML("", "b")) + para(runXML("", "c"))
	doc, err := Parse(buildDocx(t, body))
	require.NoError(t, err)
	require.NotNil(t, doc)
	ids := make(map[string]bool)
	for _, sec := range doc.Sections {
		for _, b := range sec.Blocks {
			if p, ok := b.(*ir.Paragraph); ok {
				assert.False(t, ids[p.ID], "duplicate paragraph ID: %s", p.ID)
				ids[p.ID] = true
			}
		}
	}
}

// ---------------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------------

// firstParagraph returns the first Paragraph block from the first section,
// failing the test if none is found.
func firstParagraph(t *testing.T, doc *ir.Document) *ir.Paragraph {
	t.Helper()
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	for _, b := range doc.Sections[0].Blocks {
		if p, ok := b.(*ir.Paragraph); ok {
			return p
		}
	}
	t.Fatal("no paragraph found in first section")
	return nil
}

// Note: firstTable helper is in ooxml_tables_images_test.go
