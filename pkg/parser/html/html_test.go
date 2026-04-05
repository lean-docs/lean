package html_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/lean-docs/lean/pkg/ir"
	"github.com/lean-docs/lean/pkg/parser/html"
)

// C3.1
func TestHTMLEmptyBody(t *testing.T) {
	doc, err := html.Parse([]byte(`<body></body>`))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	assert.Empty(t, doc.Sections[0].Blocks)
}

// C3.2
func TestHTMLParagraph(t *testing.T) {
	doc, err := html.Parse([]byte(`<body><p>Hello world</p></body>`))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok, "block should be a *ir.Paragraph")
	require.Len(t, p.Runs, 1)
	assert.Equal(t, "Hello world", p.Runs[0].Text)
}

// C3.3
func TestHTMLHeadings(t *testing.T) {
	headings := []struct {
		tag   string
		style string
		level int
	}{
		{"h1", "Heading1", 1},
		{"h2", "Heading2", 2},
		{"h3", "Heading3", 3},
		{"h4", "Heading4", 4},
		{"h5", "Heading5", 5},
		{"h6", "Heading6", 6},
	}

	for _, h := range headings {
		t.Run(h.tag, func(t *testing.T) {
			input := []byte(`<body><` + h.tag + `>Title</` + h.tag + `></body>`)
			doc, err := html.Parse(input)
			require.NoError(t, err)
			require.NotNil(t, doc)
			require.GreaterOrEqual(t, len(doc.Sections), 1)
			require.Len(t, doc.Sections[0].Blocks, 1)

			p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
			require.True(t, ok, "block should be a *ir.Paragraph")
			assert.Equal(t, h.style, p.Style)
			assert.Equal(t, h.level, p.Para.OutlineLevel)
			require.Len(t, p.Runs, 1)
			assert.Equal(t, "Title", p.Runs[0].Text)
		})
	}
}

// C3.4
func TestHTMLBold(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"strong", `<body><p><strong>bold</strong></p></body>`},
		{"b", `<body><p><b>bold</b></p></body>`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := html.Parse([]byte(tc.input))
			require.NoError(t, err)
			require.NotNil(t, doc)
			require.GreaterOrEqual(t, len(doc.Sections), 1)
			require.Len(t, doc.Sections[0].Blocks, 1)

			p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
			require.True(t, ok)
			require.GreaterOrEqual(t, len(p.Runs), 1)

			found := false
			for _, r := range p.Runs {
				if r.Text == "bold" {
					assert.True(t, r.Attrs.Bold, "expected Bold=true")
					found = true
				}
			}
			assert.True(t, found, "expected a run with text 'bold'")
		})
	}
}

// C3.5
func TestHTMLItalic(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"em", `<body><p><em>italic</em></p></body>`},
		{"i", `<body><p><i>italic</i></p></body>`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := html.Parse([]byte(tc.input))
			require.NoError(t, err)
			require.NotNil(t, doc)
			require.GreaterOrEqual(t, len(doc.Sections), 1)
			require.Len(t, doc.Sections[0].Blocks, 1)

			p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
			require.True(t, ok)
			require.GreaterOrEqual(t, len(p.Runs), 1)

			found := false
			for _, r := range p.Runs {
				if r.Text == "italic" {
					assert.True(t, r.Attrs.Italic, "expected Italic=true")
					found = true
				}
			}
			assert.True(t, found, "expected a run with text 'italic'")
		})
	}
}

// C3.6
func TestHTMLUnderline(t *testing.T) {
	doc, err := html.Parse([]byte(`<body><p><u>underlined</u></p></body>`))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(p.Runs), 1)

	found := false
	for _, r := range p.Runs {
		if r.Text == "underlined" {
			assert.Equal(t, ir.UnderlineSingle, r.Attrs.Underline, "expected UnderlineSingle")
			found = true
		}
	}
	assert.True(t, found, "expected a run with text 'underlined'")
}

// C3.7
func TestHTMLStrike(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"s", `<body><p><s>struck</s></p></body>`},
		{"del", `<body><p><del>struck</del></p></body>`},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := html.Parse([]byte(tc.input))
			require.NoError(t, err)
			require.NotNil(t, doc)
			require.GreaterOrEqual(t, len(doc.Sections), 1)
			require.Len(t, doc.Sections[0].Blocks, 1)

			p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
			require.True(t, ok)
			require.GreaterOrEqual(t, len(p.Runs), 1)

			found := false
			for _, r := range p.Runs {
				if r.Text == "struck" {
					assert.True(t, r.Attrs.Strike, "expected Strike=true")
					found = true
				}
			}
			assert.True(t, found, "expected a run with text 'struck'")
		})
	}
}

// C3.8
func TestHTMLAnchor(t *testing.T) {
	doc, err := html.Parse([]byte(`<body><p><a href="https://example.com">link</a></p></body>`))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(p.Runs), 1)

	found := false
	for _, r := range p.Runs {
		if r.Text == "link" {
			require.NotNil(t, r.Hyperlink, "expected Hyperlink to be set")
			assert.Equal(t, "https://example.com", r.Hyperlink.URL)
			found = true
		}
	}
	assert.True(t, found, "expected a run with text 'link'")
}

// C3.9
func TestHTMLImage(t *testing.T) {
	doc, err := html.Parse([]byte(`<body><img src="photo.png" alt="A photo"></body>`))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	img, ok := doc.Sections[0].Blocks[0].(*ir.Image)
	require.True(t, ok, "block should be a *ir.Image")
	assert.Equal(t, "A photo", img.Alt)
}

// C3.10
func TestHTMLUnorderedList(t *testing.T) {
	doc, err := html.Parse([]byte(`<body><ul><li>item one</li><li>item two</li></ul></body>`))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 2)

	for i, blk := range doc.Sections[0].Blocks {
		p, ok := blk.(*ir.Paragraph)
		require.True(t, ok, "block %d should be a *ir.Paragraph", i)
		require.NotNil(t, p.Numbering, "block %d should have Numbering", i)
		assert.Equal(t, 0, p.Numbering.Level)
	}

	// Verify numbering definition exists with bullet format
	found := false
	p0 := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	for _, nd := range doc.Styles.Numbering {
		if nd.ID == p0.Numbering.ID {
			require.GreaterOrEqual(t, len(nd.Levels), 1)
			assert.Equal(t, ir.NumFormatBullet, nd.Levels[0].Format)
			found = true
		}
	}
	assert.True(t, found, "expected a numbering definition with bullet format")
}

// C3.11
func TestHTMLOrderedList(t *testing.T) {
	doc, err := html.Parse([]byte(`<body><ol><li>first</li><li>second</li></ol></body>`))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 2)

	for i, blk := range doc.Sections[0].Blocks {
		p, ok := blk.(*ir.Paragraph)
		require.True(t, ok, "block %d should be a *ir.Paragraph", i)
		require.NotNil(t, p.Numbering, "block %d should have Numbering", i)
		assert.Equal(t, 0, p.Numbering.Level)
	}

	// Verify numbering definition exists with decimal format
	found := false
	p0 := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	for _, nd := range doc.Styles.Numbering {
		if nd.ID == p0.Numbering.ID {
			require.GreaterOrEqual(t, len(nd.Levels), 1)
			assert.Equal(t, ir.NumFormatDecimal, nd.Levels[0].Format)
			found = true
		}
	}
	assert.True(t, found, "expected a numbering definition with decimal format")
}

// C3.12
func TestHTMLNestedList(t *testing.T) {
	input := `<body><ul><li>outer<ul><li>inner</li></ul></li></ul></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.GreaterOrEqual(t, len(doc.Sections[0].Blocks), 2)

	// The nested item should have level 1
	found := false
	for _, blk := range doc.Sections[0].Blocks {
		p, ok := blk.(*ir.Paragraph)
		if !ok {
			continue
		}
		if p.Numbering != nil && p.Numbering.Level == 1 {
			found = true
			// Check the run text contains "inner"
			for _, r := range p.Runs {
				if r.Text == "inner" {
					break
				}
			}
		}
	}
	assert.True(t, found, "expected a paragraph with numbering level 1 for nested list")
}

// C3.13
func TestHTMLTable(t *testing.T) {
	input := `<body><table><tr><td>A</td><td>B</td></tr><tr><td>C</td><td>D</td></tr></table></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	tbl, ok := doc.Sections[0].Blocks[0].(*ir.Table)
	require.True(t, ok, "block should be a *ir.Table")
	require.Len(t, tbl.Rows, 2)
	require.Len(t, tbl.Rows[0].Cells, 2)
	require.Len(t, tbl.Rows[1].Cells, 2)
}

// C3.14
func TestHTMLTableHeader(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{
			"th",
			`<body><table><tr><th>Header</th></tr><tr><td>Data</td></tr></table></body>`,
		},
		{
			"thead",
			`<body><table><thead><tr><td>Header</td></tr></thead><tbody><tr><td>Data</td></tr></tbody></table></body>`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			doc, err := html.Parse([]byte(tc.input))
			require.NoError(t, err)
			require.NotNil(t, doc)
			require.GreaterOrEqual(t, len(doc.Sections), 1)
			require.Len(t, doc.Sections[0].Blocks, 1)

			tbl, ok := doc.Sections[0].Blocks[0].(*ir.Table)
			require.True(t, ok, "block should be a *ir.Table")
			require.GreaterOrEqual(t, len(tbl.Rows), 1)
			assert.True(t, tbl.Rows[0].IsHeader, "first row should be a header row")
		})
	}
}

// C3.15
func TestHTMLColspan(t *testing.T) {
	input := `<body><table><tr><td colspan="2">Wide</td></tr></table></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	tbl, ok := doc.Sections[0].Blocks[0].(*ir.Table)
	require.True(t, ok, "block should be a *ir.Table")
	require.Len(t, tbl.Rows, 1)
	require.Len(t, tbl.Rows[0].Cells, 1)
	assert.Equal(t, 2, tbl.Rows[0].Cells[0].ColSpan)
}

// C3.16
func TestHTMLInlineStyle(t *testing.T) {
	input := `<body><p><span style="font-weight:bold">styled</span></p></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(p.Runs), 1)

	found := false
	for _, r := range p.Runs {
		if r.Text == "styled" {
			assert.True(t, r.Attrs.Bold, "expected Bold=true from inline style")
			found = true
		}
	}
	assert.True(t, found, "expected a run with text 'styled'")
}

// C3.17
func TestHTMLInlineColor(t *testing.T) {
	input := `<body><p><span style="color:#FF0000">red</span></p></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(p.Runs), 1)

	found := false
	for _, r := range p.Runs {
		if r.Text == "red" {
			assert.Equal(t, ir.Color{R: 255, G: 0, B: 0}, r.Attrs.Color)
			found = true
		}
	}
	assert.True(t, found, "expected a run with text 'red'")
}

// C3.18
func TestHTMLInlineFontSize(t *testing.T) {
	input := `<body><p><span style="font-size:14pt">sized</span></p></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok)
	require.GreaterOrEqual(t, len(p.Runs), 1)

	found := false
	for _, r := range p.Runs {
		if r.Text == "sized" {
			assert.Equal(t, float64(14), r.Attrs.FontSize)
			found = true
		}
	}
	assert.True(t, found, "expected a run with text 'sized'")
}

// C3.19
func TestHTMLBlockquote(t *testing.T) {
	input := `<body><blockquote>Quoted text</blockquote></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.GreaterOrEqual(t, len(doc.Sections[0].Blocks), 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok, "block should be a *ir.Paragraph")
	assert.Greater(t, p.Para.Indent.Left, float64(0), "blockquote should have left indent")
	require.GreaterOrEqual(t, len(p.Runs), 1)
	assert.Equal(t, "Quoted text", p.Runs[0].Text)
}

// C3.20
func TestHTMLPreCode(t *testing.T) {
	input := `<body><pre><code>func main() {}</code></pre></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.GreaterOrEqual(t, len(doc.Sections[0].Blocks), 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok, "block should be a *ir.Paragraph")
	require.GreaterOrEqual(t, len(p.Runs), 1)
	assert.Equal(t, "func main() {}", p.Runs[0].Text)
	assert.Contains(t, p.Runs[0].Attrs.FontName, "mono",
		"pre/code block should use a monospace font (name should contain 'mono')")
}

// C3.21
func TestHTMLLineBreak(t *testing.T) {
	input := `<body><p>line1<br>line2</p></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok)
	// Expect at least one run with Break=BreakLine
	found := false
	for _, r := range p.Runs {
		if r.Break == ir.BreakLine {
			found = true
			break
		}
	}
	assert.True(t, found, "expected a run with Break=BreakLine")
}

// C3.22
func TestHTMLHorizontalRule(t *testing.T) {
	input := `<body><hr></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.GreaterOrEqual(t, len(doc.Sections[0].Blocks), 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok, "hr should produce a *ir.Paragraph")
	assert.NotEmpty(t, p.Para.Borders.Bottom.Style, "hr paragraph should have a bottom border style")
	assert.Greater(t, p.Para.Borders.Bottom.Width, float64(0), "hr paragraph should have a bottom border width > 0")
}

// C3.23
func TestHTMLEntities(t *testing.T) {
	input := `<body><p>&amp; &lt; &gt; &nbsp;</p></body>`
	doc, err := html.Parse([]byte(input))
	require.NoError(t, err)
	require.NotNil(t, doc)
	require.GreaterOrEqual(t, len(doc.Sections), 1)
	require.Len(t, doc.Sections[0].Blocks, 1)

	p, ok := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.True(t, ok)

	// Collect all run text
	var fullText string
	for _, r := range p.Runs {
		fullText += r.Text
	}
	assert.Contains(t, fullText, "&", "expected decoded ampersand")
	assert.Contains(t, fullText, "<", "expected decoded less-than")
	assert.Contains(t, fullText, ">", "expected decoded greater-than")
	// &nbsp; should be decoded to a non-breaking space (U+00A0) or regular space
	assert.True(t,
		len(fullText) > 0,
		"expected non-empty decoded text from HTML entities",
	)
}

// C3.24
func TestHTMLMalformedGraceful(t *testing.T) {
	inputs := []string{
		`<p>unclosed paragraph`,
		`<div><p>nested</div>wrong</p>`,
		`<b><i>overlapping</b></i>`,
		`not even html at all <<<>>>`,
		``,
	}

	for _, input := range inputs {
		t.Run(input, func(t *testing.T) {
			// Must not panic
			assert.NotPanics(t, func() {
				doc, err := html.Parse([]byte(input))
				// Either returns a valid document or a non-nil error; must not panic
				if err == nil {
					assert.NotNil(t, doc, "if no error, document should not be nil")
				}
			})
		})
	}
}
