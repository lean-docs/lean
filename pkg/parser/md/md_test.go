package md_test

import (
	"testing"

	"github.com/lean-docs/lean/pkg/ir"
	"github.com/lean-docs/lean/pkg/parser/md"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// C2.1
func TestMDEmptyDocument(t *testing.T) {
	doc, err := md.Parse([]byte(""))
	require.NoError(t, err)
	require.Len(t, doc.Sections, 1, "empty input produces one section")
	assert.Empty(t, doc.Sections[0].Blocks, "empty input produces no blocks")
}

// C2.2
func TestMDPlainParagraph(t *testing.T) {
	doc, err := md.Parse([]byte("Hello, world"))
	require.NoError(t, err)
	require.Len(t, doc.Sections[0].Blocks, 1)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Len(t, p.Runs, 1)
	assert.Equal(t, "Hello, world", p.Runs[0].Text)
}

// C2.3
func TestMDMultipleParagraphs(t *testing.T) {
	input := "First paragraph\n\nSecond paragraph"
	doc, err := md.Parse([]byte(input))
	require.NoError(t, err)
	require.Len(t, doc.Sections[0].Blocks, 2)

	p1 := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "First paragraph", p1.Runs[0].Text)

	p2 := doc.Sections[0].Blocks[1].(*ir.Paragraph)
	assert.Equal(t, "Second paragraph", p2.Runs[0].Text)
}

// C2.4
func TestMDBold(t *testing.T) {
	doc, err := md.Parse([]byte("**bold text**"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Len(t, p.Runs, 1)
	assert.Equal(t, "bold text", p.Runs[0].Text)
	assert.True(t, p.Runs[0].Attrs.Bold)
}

// C2.5
func TestMDItalic(t *testing.T) {
	doc, err := md.Parse([]byte("*italic text*"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Len(t, p.Runs, 1)
	assert.Equal(t, "italic text", p.Runs[0].Text)
	assert.True(t, p.Runs[0].Attrs.Italic)
}

// C2.6
func TestMDBoldItalic(t *testing.T) {
	doc, err := md.Parse([]byte("***bold italic***"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Len(t, p.Runs, 1)
	assert.True(t, p.Runs[0].Attrs.Bold)
	assert.True(t, p.Runs[0].Attrs.Italic)
}

// C2.7
func TestMDStrike(t *testing.T) {
	doc, err := md.Parse([]byte("~~strikethrough~~"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Len(t, p.Runs, 1)
	assert.True(t, p.Runs[0].Attrs.Strike)
}

// C2.8
func TestMDInlineCode(t *testing.T) {
	doc, err := md.Parse([]byte("`code`"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Len(t, p.Runs, 1)
	assert.Equal(t, "code", p.Runs[0].Text)
	assert.NotEmpty(t, p.Runs[0].Attrs.FontName, "inline code must set a monospace font")
}

// C2.9
func TestMDHeading1(t *testing.T) {
	doc, err := md.Parse([]byte("# Title"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "Heading1", p.Style)
	assert.Equal(t, "Title", p.Runs[0].Text)
}

// C2.10
func TestMDHeading2(t *testing.T) {
	doc, err := md.Parse([]byte("## Subtitle"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, "Heading2", p.Style)
}

// C2.11
func TestMDHeading3to6(t *testing.T) {
	tests := []struct {
		input string
		style string
	}{
		{"### H3", "Heading3"},
		{"#### H4", "Heading4"},
		{"##### H5", "Heading5"},
		{"###### H6", "Heading6"},
	}
	for _, tt := range tests {
		t.Run(tt.style, func(t *testing.T) {
			doc, err := md.Parse([]byte(tt.input))
			require.NoError(t, err)
			p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
			assert.Equal(t, tt.style, p.Style)
		})
	}
}

// C2.12
func TestMDBulletListHyphen(t *testing.T) {
	doc, err := md.Parse([]byte("- item one\n- item two"))
	require.NoError(t, err)
	require.Len(t, doc.Sections[0].Blocks, 2)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.NotNil(t, p.Numbering)
	assert.Equal(t, 0, p.Numbering.Level)
	// Verify numbering def exists and is bullet format
	require.NotEmpty(t, doc.Styles.Numbering)
	found := false
	for _, nd := range doc.Styles.Numbering {
		if nd.ID == p.Numbering.ID {
			assert.Equal(t, ir.NumFormatBullet, nd.Levels[0].Format)
			found = true
		}
	}
	assert.True(t, found, "numbering def for bullet list must exist")
}

// C2.13
func TestMDBulletListAsterisk(t *testing.T) {
	doc, err := md.Parse([]byte("* item one\n* item two"))
	require.NoError(t, err)
	require.Len(t, doc.Sections[0].Blocks, 2)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.NotNil(t, p.Numbering)
}

// C2.14
func TestMDOrderedList(t *testing.T) {
	doc, err := md.Parse([]byte("1. first\n2. second"))
	require.NoError(t, err)
	require.Len(t, doc.Sections[0].Blocks, 2)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.NotNil(t, p.Numbering)
	// Verify decimal format
	found := false
	for _, nd := range doc.Styles.Numbering {
		if nd.ID == p.Numbering.ID {
			assert.Equal(t, ir.NumFormatDecimal, nd.Levels[0].Format)
			found = true
		}
	}
	assert.True(t, found)
}

// C2.15
func TestMDNestedList(t *testing.T) {
	input := "- outer\n  - inner"
	doc, err := md.Parse([]byte(input))
	require.NoError(t, err)
	require.Len(t, doc.Sections[0].Blocks, 2)

	p0 := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	p1 := doc.Sections[0].Blocks[1].(*ir.Paragraph)
	require.NotNil(t, p0.Numbering)
	require.NotNil(t, p1.Numbering)
	assert.Equal(t, 0, p0.Numbering.Level)
	assert.Equal(t, 1, p1.Numbering.Level)
}

// C2.16
func TestMDBlockquote(t *testing.T) {
	doc, err := md.Parse([]byte("> quoted text"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.Greater(t, p.Para.Indent.Left, 0.0, "blockquote must have left indent")
}

// C2.17
func TestMDCodeBlock(t *testing.T) {
	input := "```\nfunc main() {}\n```"
	doc, err := md.Parse([]byte(input))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.NotEmpty(t, p.Runs[0].Attrs.FontName, "code block must set monospace font")
}

// C2.18
func TestMDHorizontalRule(t *testing.T) {
	doc, err := md.Parse([]byte("---"))
	require.NoError(t, err)
	require.Len(t, doc.Sections[0].Blocks, 1)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	assert.NotEmpty(t, p.Para.Borders.Bottom.Style, "horizontal rule must have bottom border")
}

// C2.19
func TestMDLink(t *testing.T) {
	doc, err := md.Parse([]byte("[click](https://example.com)"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	require.NotNil(t, p.Runs[0].Hyperlink)
	assert.Equal(t, "https://example.com", p.Runs[0].Hyperlink.URL)
	assert.Equal(t, "click", p.Runs[0].Text)
}

// C2.20
func TestMDImage(t *testing.T) {
	doc, err := md.Parse([]byte("![alt text](image.png)"))
	require.NoError(t, err)
	require.Len(t, doc.Sections[0].Blocks, 1)
	img := doc.Sections[0].Blocks[0].(*ir.Image)
	assert.Equal(t, "alt text", img.Alt)
}

// C2.21
func TestMDLineBreak(t *testing.T) {
	doc, err := md.Parse([]byte("line one  \nline two"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// Should have runs where the first ends with a line break
	foundBreak := false
	for _, r := range p.Runs {
		if r.Break == ir.BreakLine {
			foundBreak = true
		}
	}
	assert.True(t, foundBreak, "trailing double space must produce line break")
}

// C2.22
func TestMDHardBreak(t *testing.T) {
	doc, err := md.Parse([]byte("line one\\\nline two"))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	foundBreak := false
	for _, r := range p.Runs {
		if r.Break == ir.BreakLine {
			foundBreak = true
		}
	}
	assert.True(t, foundBreak, "backslash line ending must produce line break")
}

// C2.23
func TestMDTable(t *testing.T) {
	input := "| A | B |\n|---|---|\n| 1 | 2 |"
	doc, err := md.Parse([]byte(input))
	require.NoError(t, err)
	require.Len(t, doc.Sections[0].Blocks, 1)
	tbl := doc.Sections[0].Blocks[0].(*ir.Table)
	require.Len(t, tbl.Rows, 2) // header + 1 data row
	require.Len(t, tbl.Rows[0].Cells, 2)
}

// C2.24
func TestMDTableAlignment(t *testing.T) {
	input := "| L | C | R |\n|:--|:--:|--:|\n| a | b | c |"
	doc, err := md.Parse([]byte(input))
	require.NoError(t, err)
	tbl := doc.Sections[0].Blocks[0].(*ir.Table)
	// Check data row cell alignment
	dataRow := tbl.Rows[1]
	p0 := dataRow.Cells[0].Blocks[0].(*ir.Paragraph)
	p1 := dataRow.Cells[1].Blocks[0].(*ir.Paragraph)
	p2 := dataRow.Cells[2].Blocks[0].(*ir.Paragraph)
	assert.Equal(t, ir.AlignLeft, p0.Para.Align)
	assert.Equal(t, ir.AlignCenter, p1.Para.Align)
	assert.Equal(t, ir.AlignRight, p2.Para.Align)
}

// C2.25
func TestMDFrontmatter(t *testing.T) {
	input := "---\ntitle: My Document\nauthor: Jane Doe\n---\n\nContent"
	doc, err := md.Parse([]byte(input))
	require.NoError(t, err)
	assert.Equal(t, "My Document", doc.Meta.Title)
	assert.Equal(t, "Jane Doe", doc.Meta.Author)
}

// C2.26
func TestMDUnicodeText(t *testing.T) {
	inputs := []string{
		"中文测试",         // CJK
		"مرحبا بالعالم",  // Arabic
		"こんにちは",        // Japanese
		"Hello 🌍 World", // Emoji
	}
	for _, input := range inputs {
		t.Run(input[:6], func(t *testing.T) {
			doc, err := md.Parse([]byte(input))
			require.NoError(t, err)
			p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
			assert.Equal(t, input, p.Runs[0].Text)
		})
	}
}

// C2.27
func TestMDMixedInlineFormatting(t *testing.T) {
	// Bold inside italic inside link
	input := "[***bold italic link***](https://example.com)"
	doc, err := md.Parse([]byte(input))
	require.NoError(t, err)
	p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// Should have at least one run that is bold+italic with a hyperlink
	foundComplex := false
	for _, r := range p.Runs {
		if r.Attrs.Bold && r.Attrs.Italic && r.Hyperlink != nil {
			foundComplex = true
			assert.Equal(t, "https://example.com", r.Hyperlink.URL)
		}
	}
	assert.True(t, foundComplex, "must parse bold+italic inside link")
}
