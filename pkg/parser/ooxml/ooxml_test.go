package ooxml

import (
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
	// TODO: fixture with center-aligned paragraph
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Equal(t, ir.AlignCenter, p.Para.Align)
}

// C4.16 — Paragraph spacing (before/after)
func TestOOXMLParagraphSpacing(t *testing.T) {
	// TODO: fixture with paragraph spacing
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Greater(t, p.Para.Spacing.Before, float64(0))
	assert.Greater(t, p.Para.Spacing.After, float64(0))
}

// C4.17 — Paragraph indent (left, first-line)
func TestOOXMLParagraphIndent(t *testing.T) {
	// TODO: fixture with paragraph indent
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.Greater(t, p.Para.Indent.Left, float64(0))
	assert.Greater(t, p.Para.Indent.FirstLine, float64(0))
}

// C4.18 — Tab stops
func TestOOXMLTabStops(t *testing.T) {
	// TODO: fixture with custom tab stops
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.GreaterOrEqual(t, len(p.Para.TabStops), 1)
	assert.Greater(t, p.Para.TabStops[0].Position, float64(0))
}

// C4.19 — KeepTogether / KeepWithNext
func TestOOXMLKeepTogetherKeepWithNext(t *testing.T) {
	// TODO: fixture with keep-together and keep-with-next
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.True(t, p.Para.KeepTogether)
	assert.True(t, p.Para.KeepWithNext)
}

// C4.20 — PageBreakBefore
func TestOOXMLPageBreakBefore(t *testing.T) {
	// TODO: fixture with page-break-before paragraph
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.True(t, p.Para.PageBreakBefore)
}

// C4.21 — Named styles parsed into StyleSheet
func TestOOXMLNamedStyles(t *testing.T) {
	// TODO: fixture with named styles (e.g. Heading1)
	doc, err := Parse(nil)
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.NotEmpty(t, doc.Styles.Named, "should have at least one named style")
}

// C4.22 — Paragraph references a named style
func TestOOXMLParagraphStyleRef(t *testing.T) {
	// TODO: fixture with a paragraph referencing a style
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	assert.NotEmpty(t, p.Style, "paragraph should reference a named style")
}

// C4.23 — Hyperlink with URL
func TestOOXMLHyperlinkURL(t *testing.T) {
	// TODO: fixture with an external hyperlink
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	var found bool
	for _, r := range p.Runs {
		if r.Hyperlink != nil && r.Hyperlink.URL != "" {
			found = true
			break
		}
	}
	assert.True(t, found, "should find a run with a hyperlink URL")
}

// C4.24 — Internal bookmark hyperlink
func TestOOXMLHyperlinkBookmark(t *testing.T) {
	// TODO: fixture with an internal bookmark hyperlink
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	var found bool
	for _, r := range p.Runs {
		if r.Hyperlink != nil && r.Hyperlink.Bookmark != "" {
			found = true
			break
		}
	}
	assert.True(t, found, "should find a run with a bookmark hyperlink")
}

// C4.25 — Line break
func TestOOXMLLineBreak(t *testing.T) {
	// TODO: fixture with a line break inside a paragraph
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	var found bool
	for _, r := range p.Runs {
		if r.Break == ir.BreakLine {
			found = true
			break
		}
	}
	assert.True(t, found, "should find a run with BreakLine")
}

// C4.26 — Page break (inline)
func TestOOXMLPageBreakInline(t *testing.T) {
	// TODO: fixture with an inline page break
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	var found bool
	for _, r := range p.Runs {
		if r.Break == ir.BreakPage {
			found = true
			break
		}
	}
	assert.True(t, found, "should find a run with BreakPage")
}

// C4.27 — Column break
func TestOOXMLColumnBreak(t *testing.T) {
	// TODO: fixture with a column break
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	var found bool
	for _, r := range p.Runs {
		if r.Break == ir.BreakColumn {
			found = true
			break
		}
	}
	assert.True(t, found, "should find a run with BreakColumn")
}

// C4.28 — Bullet list (numbering reference)
func TestOOXMLBulletList(t *testing.T) {
	// TODO: fixture with a bullet list
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.NotNil(t, p.Numbering)
	// Verify the numbering def is bullet format
	var found bool
	for _, nd := range doc.Styles.Numbering {
		if nd.ID == p.Numbering.ID {
			require.GreaterOrEqual(t, len(nd.Levels), 1)
			assert.Equal(t, ir.NumFormatBullet, nd.Levels[0].Format)
			found = true
			break
		}
	}
	assert.True(t, found, "numbering def should exist for bullet list")
}

// C4.29 — Numbered list (decimal)
func TestOOXMLNumberedList(t *testing.T) {
	// TODO: fixture with a decimal numbered list
	doc, err := Parse(nil)
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
	// TODO: fixture with multi-level list
	doc, err := Parse(nil)
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
	// TODO: fixture with a footnote
	doc, err := Parse(nil)
	require.NoError(t, err)
	p := firstParagraph(t, doc)
	require.GreaterOrEqual(t, len(p.Footnotes), 1)
	assert.NotEmpty(t, p.Footnotes[0].ID)
	assert.GreaterOrEqual(t, len(p.Footnotes[0].Blocks), 1)
}

// C4.37 — Header content
func TestOOXMLHeaderContent(t *testing.T) {
	// TODO: fixture with a page header
	doc, err := Parse(nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.NotNil(t, doc.Sections[0].Header)
	assert.GreaterOrEqual(t, len(doc.Sections[0].Header.Blocks), 1)
}

// C4.38 — Footer content
func TestOOXMLFooterContent(t *testing.T) {
	// TODO: fixture with a page footer
	doc, err := Parse(nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.NotNil(t, doc.Sections[0].Footer)
	assert.GreaterOrEqual(t, len(doc.Sections[0].Footer.Blocks), 1)
}

// C4.39 — Page properties (dimensions, margins, orientation)
func TestOOXMLPageProperties(t *testing.T) {
	// TODO: fixture with explicit page size and margins
	doc, err := Parse(nil)
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
	// TODO: fixture with landscape page
	doc, err := Parse(nil)
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	assert.Equal(t, ir.Landscape, doc.Sections[0].Properties.Orientation)
}

// C4.41 — Document metadata (title, author)
func TestOOXMLDocumentMetadata(t *testing.T) {
	// TODO: fixture with core properties (title, author)
	doc, err := Parse(nil)
	require.NoError(t, err)
	require.NotNil(t, doc)
	assert.NotEmpty(t, doc.Meta.Title)
	assert.NotEmpty(t, doc.Meta.Author)
}

// C4.42 — Bookmark block
func TestOOXMLBookmark(t *testing.T) {
	// TODO: fixture with a bookmark
	doc, err := Parse(nil)
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
	// TODO: fixture — .docx with an unknown XML element in w:body
	doc, err := Parse(nil)
	require.NoError(t, err)
	require.NotNil(t, doc)
	// The unknown element should be silently skipped;
	// the rest of the document should parse correctly.
	assert.GreaterOrEqual(t, len(doc.Sections), 1)
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
