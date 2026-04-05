package ir_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Cluster 0 ---

// C0.4
func TestEmptyDocumentIR(t *testing.T) {
	doc := ir.NewDocument()
	require.NotNil(t, doc)
	assert.NotEmpty(t, doc.Meta.IRVersion)
	assert.False(t, doc.Meta.CreatedAt.IsZero())
	assert.False(t, doc.Meta.ModifiedAt.IsZero())
	require.Len(t, doc.Sections, 1, "must have one default section")
	assert.Greater(t, doc.Sections[0].Properties.Width, 0.0)
	assert.Greater(t, doc.Sections[0].Properties.Height, 0.0)
}

// C0.5
func TestIRJSONRoundTrip(t *testing.T) {
	doc := ir.NewDocument()
	data, err := json.Marshal(doc)
	require.NoError(t, err)

	var restored ir.Document
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err)

	assert.Equal(t, doc.Meta.IRVersion, restored.Meta.IRVersion)
	assert.Equal(t, doc.Meta.CreatedAt.Unix(), restored.Meta.CreatedAt.Unix())
	require.Len(t, restored.Sections, 1)
	assert.Equal(t, doc.Sections[0].Properties, restored.Sections[0].Properties)
}

// --- Cluster 1 ---

// C1.1
func TestIRSingleParagraph(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID:   "p1",
			Runs: []ir.Run{{Text: "Hello, world"}},
		},
	}
	restored := roundTrip(t, doc)
	require.Len(t, restored.Sections[0].Blocks, 1)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "p1", p.ID)
	require.Len(t, p.Runs, 1)
	assert.Equal(t, "Hello, world", p.Runs[0].Text)
}

// C1.2
func TestIRRunAttrs(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Runs: []ir.Run{
				{Text: "styled", Attrs: ir.RunAttrs{
					Bold:      true,
					Italic:    true,
					Underline: ir.UnderlineSingle,
					Strike:    true,
					SmallCaps: true,
					AllCaps:   true,
				}},
			},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	attrs := p.Runs[0].Attrs
	assert.True(t, attrs.Bold)
	assert.True(t, attrs.Italic)
	assert.Equal(t, ir.UnderlineSingle, attrs.Underline)
	assert.True(t, attrs.Strike)
	assert.True(t, attrs.SmallCaps)
	assert.True(t, attrs.AllCaps)
}

// C1.3
func TestIRFontProperties(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Runs: []ir.Run{
				{Text: "font", Attrs: ir.RunAttrs{
					FontName: "Roboto",
					FontSize: 14.5,
					Tracking: 1.2,
				}},
			},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "Roboto", p.Runs[0].Attrs.FontName)
	assert.Equal(t, 14.5, p.Runs[0].Attrs.FontSize)
	assert.Equal(t, 1.2, p.Runs[0].Attrs.Tracking)
}

// C1.4
func TestIRColorRGB(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Runs: []ir.Run{
				{Text: "red", Attrs: ir.RunAttrs{
					Color: ir.Color{R: 255, G: 0, B: 0},
				}},
			},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	c := p.Runs[0].Attrs.Color
	assert.Equal(t, uint8(255), c.R)
	assert.Equal(t, uint8(0), c.G)
	assert.Equal(t, uint8(0), c.B)
	assert.False(t, c.None)
}

// C1.5
func TestIRColorNone(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Runs: []ir.Run{
				{Text: "auto", Attrs: ir.RunAttrs{
					Color: ir.Color{None: true},
				}},
			},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.True(t, p.Runs[0].Attrs.Color.None)
}

// C1.6
func TestIRThemeColors(t *testing.T) {
	doc := ir.NewDocument()
	doc.Styles.Theme.Colors = map[string]ir.Color{
		"accent1": {R: 79, G: 129, B: 189},
		"dk1":     {R: 0, G: 0, B: 0},
		"lt1":     {R: 255, G: 255, B: 255},
	}
	restored := roundTrip(t, doc)
	require.Len(t, restored.Styles.Theme.Colors, 3)
	assert.Equal(t, uint8(79), restored.Styles.Theme.Colors["accent1"].R)
}

// C1.7
func TestIRMultipleParagraphs(t *testing.T) {
	doc := ir.NewDocument()
	blocks := make([]ir.Block, 10)
	for i := range blocks {
		blocks[i] = &ir.Paragraph{
			ID:   fmt.Sprintf("p%d", i),
			Runs: []ir.Run{{Text: fmt.Sprintf("Paragraph %d", i)}},
		}
	}
	doc.Sections[0].Blocks = blocks

	restored := roundTrip(t, doc)
	require.Len(t, restored.Sections[0].Blocks, 10)
	for i, b := range restored.Sections[0].Blocks {
		p := b.(*ir.Paragraph)
		assert.Equal(t, fmt.Sprintf("p%d", i), p.ID)
	}
}

// C1.8
func TestIRParagraphAlignment(t *testing.T) {
	alignments := []ir.Alignment{ir.AlignLeft, ir.AlignCenter, ir.AlignRight, ir.AlignJustify}
	doc := ir.NewDocument()
	for i, a := range alignments {
		doc.Sections[0].Blocks = append(doc.Sections[0].Blocks, &ir.Paragraph{
			ID:   fmt.Sprintf("p%d", i),
			Para: ir.ParaAttrs{Align: a},
			Runs: []ir.Run{{Text: "text"}},
		})
	}
	restored := roundTrip(t, doc)
	for i, a := range alignments {
		p := restored.Sections[0].Blocks[i].(*ir.Paragraph)
		assert.Equal(t, a, p.Para.Align, "alignment %d", i)
	}
}

// C1.9
func TestIRSpacing(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Para: ir.ParaAttrs{
				Spacing: ir.Spacing{Before: 12, After: 6, Line: 1.5, LineRule: ir.LineRuleExact},
			},
			Runs: []ir.Run{{Text: "spaced"}},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, 12.0, p.Para.Spacing.Before)
	assert.Equal(t, 6.0, p.Para.Spacing.After)
	assert.Equal(t, 1.5, p.Para.Spacing.Line)
	assert.Equal(t, ir.LineRuleExact, p.Para.Spacing.LineRule)
}

// C1.10
func TestIRIndent(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Para: ir.ParaAttrs{
				Indent: ir.Indent{Left: 36, Right: 18, FirstLine: 24, Hanging: 12},
			},
			Runs: []ir.Run{{Text: "indented"}},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, 36.0, p.Para.Indent.Left)
	assert.Equal(t, 18.0, p.Para.Indent.Right)
	assert.Equal(t, 24.0, p.Para.Indent.FirstLine)
	assert.Equal(t, 12.0, p.Para.Indent.Hanging)
}

// C1.11
func TestIRTabStops(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Para: ir.ParaAttrs{
				TabStops: []ir.TabStop{
					{Position: 72, Alignment: ir.TabAlignLeft, Leader: ir.TabLeaderNone},
					{Position: 288, Alignment: ir.TabAlignRight, Leader: ir.TabLeaderDot},
					{Position: 360, Alignment: ir.TabAlignCenter, Leader: ir.TabLeaderDash},
				},
			},
			Runs: []ir.Run{{Text: "tabs"}},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Len(t, p.Para.TabStops, 3)
	assert.Equal(t, 72.0, p.Para.TabStops[0].Position)
	assert.Equal(t, ir.TabAlignRight, p.Para.TabStops[1].Alignment)
	assert.Equal(t, ir.TabLeaderDash, p.Para.TabStops[2].Leader)
}

// C1.12
func TestIRStyleReference(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID:    "p1",
			Style: "Heading1",
			Runs:  []ir.Run{{Text: "Title"}},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "Heading1", p.Style)
}

// C1.13
func TestIRCharacterStyle(t *testing.T) {
	doc := ir.NewDocument()
	doc.Styles.Named["Strong"] = ir.Style{
		ID:     "strong",
		Name:   "Strong",
		IsChar: true,
		RunAttrs: ir.RunAttrs{Bold: true},
	}
	restored := roundTrip(t, doc)
	s, ok := restored.Styles.Named["Strong"]
	require.True(t, ok)
	assert.True(t, s.IsChar)
	assert.True(t, s.RunAttrs.Bold)
}

// C1.14
func TestIRStyleInheritance(t *testing.T) {
	doc := ir.NewDocument()
	doc.Styles.Named["Normal"] = ir.Style{
		ID:       "normal",
		Name:     "Normal",
		RunAttrs: ir.RunAttrs{FontSize: 12, FontName: "Arial"},
	}
	doc.Styles.Named["Heading1"] = ir.Style{
		ID:       "heading1",
		Name:     "Heading1",
		BasedOn:  "Normal",
		RunAttrs: ir.RunAttrs{FontSize: 24, Bold: true},
	}

	resolved, ok := doc.Styles.ResolveStyle("Heading1")
	require.True(t, ok)
	assert.Equal(t, 24.0, resolved.RunAttrs.FontSize)
	assert.True(t, resolved.RunAttrs.Bold)
	assert.Equal(t, "Arial", resolved.RunAttrs.FontName) // inherited
}

// C1.15
func TestIRStyleSheet(t *testing.T) {
	doc := ir.NewDocument()
	doc.Styles.Defaults = ir.RunAttrs{FontSize: 11, FontName: "Calibri"}
	doc.Styles.Named["Normal"] = ir.Style{
		ID:   "normal",
		Name: "Normal",
	}
	doc.Styles.Named["Heading1"] = ir.Style{
		ID:      "heading1",
		Name:    "Heading1",
		BasedOn: "Normal",
	}
	restored := roundTrip(t, doc)
	assert.Equal(t, 11.0, restored.Styles.Defaults.FontSize)
	assert.Equal(t, "Calibri", restored.Styles.Defaults.FontName)
	require.Len(t, restored.Styles.Named, 2)
}

// C1.16
func TestIRNumberingDef(t *testing.T) {
	doc := ir.NewDocument()
	doc.Styles.Numbering = []ir.NumberingDef{
		{
			ID: "num1",
			Levels: []ir.NumberingLevel{
				{Level: 0, Format: ir.NumFormatBullet, Text: "•", Start: 1},
				{Level: 1, Format: ir.NumFormatDecimal, Text: "%1.", Start: 1},
				{Level: 2, Format: ir.NumFormatLowerAlpha, Text: "%1)", Start: 1},
			},
		},
	}
	restored := roundTrip(t, doc)
	require.Len(t, restored.Styles.Numbering, 1)
	require.Len(t, restored.Styles.Numbering[0].Levels, 3)
	assert.Equal(t, ir.NumFormatBullet, restored.Styles.Numbering[0].Levels[0].Format)
	assert.Equal(t, ir.NumFormatDecimal, restored.Styles.Numbering[0].Levels[1].Format)
}

// C1.17
func TestIRNumberingRef(t *testing.T) {
	doc := ir.NewDocument()
	restart := 1
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID:        "p1",
			Runs:      []ir.Run{{Text: "item 1"}},
			Numbering: &ir.NumberingRef{ID: "num1", Level: 0, Restart: &restart},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	require.NotNil(t, p.Numbering)
	assert.Equal(t, "num1", p.Numbering.ID)
	assert.Equal(t, 0, p.Numbering.Level)
	require.NotNil(t, p.Numbering.Restart)
	assert.Equal(t, 1, *p.Numbering.Restart)
}

// C1.18
func TestIRTable(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Table{
			ID:           "t1",
			ColumnWidths: []float64{200, 200},
			Borders: ir.TableBorders{
				Top: ir.Border{Style: "single", Width: 1, Color: ir.Color{R: 0, G: 0, B: 0}},
			},
			Shading: ir.Color{R: 240, G: 240, B: 240},
			Rows: []ir.TableRow{
				{
					Cells: []ir.TableCell{
						{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
							&ir.Paragraph{ID: "c1", Runs: []ir.Run{{Text: "A"}}},
						}},
						{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
							&ir.Paragraph{ID: "c2", Runs: []ir.Run{{Text: "B"}}},
						}},
					},
				},
				{
					Cells: []ir.TableCell{
						{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
							&ir.Paragraph{ID: "c3", Runs: []ir.Run{{Text: "C"}}},
						}},
						{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
							&ir.Paragraph{ID: "c4", Runs: []ir.Run{{Text: "D"}}},
						}},
					},
				},
			},
		},
	}
	restored := roundTrip(t, doc)
	tbl := restored.Sections[0].Blocks[0].(*ir.Table)
	assert.Equal(t, "t1", tbl.ID)
	require.Len(t, tbl.Rows, 2)
	require.Len(t, tbl.Rows[0].Cells, 2)
	p := tbl.Rows[0].Cells[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "A", p.Runs[0].Text)
	assert.Equal(t, uint8(240), tbl.Shading.R)
}

// C1.19
func TestIRTableSpan(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Table{
			ID: "t1",
			Rows: []ir.TableRow{
				{
					Cells: []ir.TableCell{
						{ColSpan: 2, RowSpan: 1, Blocks: []ir.Block{
							&ir.Paragraph{ID: "c1", Runs: []ir.Run{{Text: "span"}}},
						}},
					},
				},
				{
					Cells: []ir.TableCell{
						{ColSpan: 1, RowSpan: 2, Blocks: []ir.Block{
							&ir.Paragraph{ID: "c2", Runs: []ir.Run{{Text: "vspan"}}},
						}},
						{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
							&ir.Paragraph{ID: "c3", Runs: []ir.Run{{Text: "normal"}}},
						}},
					},
				},
			},
		},
	}
	restored := roundTrip(t, doc)
	tbl := restored.Sections[0].Blocks[0].(*ir.Table)
	assert.Equal(t, 2, tbl.Rows[0].Cells[0].ColSpan)
	assert.Equal(t, 2, tbl.Rows[1].Cells[0].RowSpan)
}

// C1.20
func TestIRImage(t *testing.T) {
	doc := ir.NewDocument()
	imgData := []byte{0x89, 0x50, 0x4E, 0x47} // PNG magic bytes
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Image{
			ID:     "img1",
			Data:   imgData,
			Format: ir.ImagePNG,
			Width:  100,
			Height: 50,
			Alt:    "test image",
		},
	}
	restored := roundTrip(t, doc)
	img := restored.Sections[0].Blocks[0].(*ir.Image)
	assert.Equal(t, "img1", img.ID)
	assert.Equal(t, imgData, img.Data)
	assert.Equal(t, ir.ImagePNG, img.Format)
	assert.Equal(t, 100.0, img.Width)
	assert.Equal(t, 50.0, img.Height)
	assert.Equal(t, "test image", img.Alt)
}

// C1.21
func TestIRHyperlink(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Runs: []ir.Run{
				{
					Text:      "click here",
					Hyperlink: &ir.Hyperlink{URL: "https://example.com"},
				},
				{
					Text:      "see section",
					Hyperlink: &ir.Hyperlink{Bookmark: "section-1"},
				},
			},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	require.NotNil(t, p.Runs[0].Hyperlink)
	assert.Equal(t, "https://example.com", p.Runs[0].Hyperlink.URL)
	require.NotNil(t, p.Runs[1].Hyperlink)
	assert.Equal(t, "section-1", p.Runs[1].Hyperlink.Bookmark)
}

// C1.22
func TestIRFootnote(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID:   "p1",
			Runs: []ir.Run{{Text: "text with footnote"}},
			Footnotes: []ir.Footnote{
				{
					ID: "fn1",
					Blocks: []ir.Block{
						&ir.Paragraph{ID: "fnp1", Runs: []ir.Run{{Text: "footnote content"}}},
					},
				},
			},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Len(t, p.Footnotes, 1)
	assert.Equal(t, "fn1", p.Footnotes[0].ID)
	fnp := p.Footnotes[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "footnote content", fnp.Runs[0].Text)
}

// C1.23
func TestIRHeaderFooter(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Header = &ir.HeaderFooter{
		Kind: ir.HeaderFooterDefault,
		Blocks: []ir.Block{
			&ir.Paragraph{ID: "h1", Runs: []ir.Run{{Text: "Header text"}}},
		},
	}
	doc.Sections[0].Footer = &ir.HeaderFooter{
		Kind: ir.HeaderFooterFirst,
		Blocks: []ir.Block{
			&ir.Paragraph{ID: "f1", Runs: []ir.Run{{Text: "Footer text"}}},
		},
	}
	restored := roundTrip(t, doc)
	require.NotNil(t, restored.Sections[0].Header)
	assert.Equal(t, ir.HeaderFooterDefault, restored.Sections[0].Header.Kind)
	hp := restored.Sections[0].Header.Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "Header text", hp.Runs[0].Text)

	require.NotNil(t, restored.Sections[0].Footer)
	assert.Equal(t, ir.HeaderFooterFirst, restored.Sections[0].Footer.Kind)
	fp := restored.Sections[0].Footer.Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "Footer text", fp.Runs[0].Text)
}

// C1.24
func TestIRSection(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections = []ir.Section{
		{
			Properties: ir.PageProperties{
				Width: 595, Height: 842, // A4
				MarginTop: 72, MarginBottom: 72,
				MarginLeft: 72, MarginRight: 72,
				Orientation: ir.Portrait,
			},
			Columns: []ir.Column{
				{Width: 225, Spacing: 18},
				{Width: 225, Spacing: 0},
			},
		},
	}
	restored := roundTrip(t, doc)
	require.Len(t, restored.Sections, 1)
	s := restored.Sections[0]
	assert.Equal(t, 595.0, s.Properties.Width)
	assert.Equal(t, 842.0, s.Properties.Height)
	assert.Equal(t, ir.Portrait, s.Properties.Orientation)
	require.Len(t, s.Columns, 2)
	assert.Equal(t, 225.0, s.Columns[0].Width)
	assert.Equal(t, 18.0, s.Columns[0].Spacing)
}

// C1.25
func TestIRBookmark(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Bookmark{ID: "bk1", Name: "introduction"},
	}
	restored := roundTrip(t, doc)
	bk := restored.Sections[0].Blocks[0].(*ir.Bookmark)
	assert.Equal(t, "bk1", bk.ID)
	assert.Equal(t, "introduction", bk.Name)
}

// C1.26
func TestIRBlockIDs(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "a"}}},
		&ir.Table{ID: "t1", Rows: []ir.TableRow{}},
		&ir.Image{ID: "img1", Data: []byte{1}},
		&ir.Bookmark{ID: "bk1", Name: "x"},
	}
	for _, b := range doc.Sections[0].Blocks {
		assert.NotEmpty(t, b.BlockID(), "block %T must have non-empty ID", b)
	}
}

// C1.27
func TestIRKeepTogether(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Para: ir.ParaAttrs{
				KeepTogether: true,
				KeepWithNext: true,
			},
			Runs: []ir.Run{{Text: "keep"}},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.True(t, p.Para.KeepTogether)
	assert.True(t, p.Para.KeepWithNext)
}

// C1.28
func TestIRBiDi(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID:   "p1",
			Para: ir.ParaAttrs{BiDi: true},
			Runs: []ir.Run{{Text: "مرحبا"}},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.True(t, p.Para.BiDi)
}

// C1.29
func TestIRBaselineShift(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Runs: []ir.Run{
				{Text: "sup", Attrs: ir.RunAttrs{Baseline: ir.BaselineSuperscript}},
				{Text: "sub", Attrs: ir.RunAttrs{Baseline: ir.BaselineSubscript}},
			},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, ir.BaselineSuperscript, p.Runs[0].Attrs.Baseline)
	assert.Equal(t, ir.BaselineSubscript, p.Runs[1].Attrs.Baseline)
}

// C1.30
func TestIRBreakTypes(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{
		&ir.Paragraph{
			ID: "p1",
			Runs: []ir.Run{
				{Text: "", Break: ir.BreakLine},
				{Text: "", Break: ir.BreakPage},
				{Text: "", Break: ir.BreakColumn},
			},
		},
	}
	restored := roundTrip(t, doc)
	p := restored.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, ir.BreakLine, p.Runs[0].Break)
	assert.Equal(t, ir.BreakPage, p.Runs[1].Break)
	assert.Equal(t, ir.BreakColumn, p.Runs[2].Break)
}

// --- helpers ---

func roundTrip(t *testing.T, doc *ir.Document) *ir.Document {
	t.Helper()
	data, err := json.Marshal(doc)
	require.NoError(t, err, "marshal failed")

	var restored ir.Document
	err = json.Unmarshal(data, &restored)
	require.NoError(t, err, "unmarshal failed")

	return &restored
}

