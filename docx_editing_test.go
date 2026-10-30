package lean_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"testing"

	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/require"
)

func editingParagraph(id, text string) *ir.Paragraph {
	return &ir.Paragraph{ID: id, Runs: []ir.Run{{Text: text}}}
}
func editingPart(t *testing.T, content []byte, name string) string {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(content), int64(len(content)))
	require.NoError(t, err)
	for _, file := range reader.File {
		if file.Name == name {
			stream, err := file.Open()
			require.NoError(t, err)
			defer stream.Close()
			data, err := io.ReadAll(stream)
			require.NoError(t, err)
			return string(data)
		}
	}
	t.Fatalf("Missing package part %s", name)
	return ""
}

func TestEditingParagraphResetsSurviveStyleInheritanceAndDOCX(t *testing.T) {
	document := ir.NewDocument()
	parent := ir.Style{ID: "Parent", ParaAttrs: ir.ParaAttrs{Align: ir.AlignCenter, Spacing: ir.Spacing{Before: 12}, Indent: ir.Indent{Left: 24}, KeepWithNext: true}, RunAttrs: ir.RunAttrs{Color: ir.Color{R: 255}}}
	child := ir.Style{ID: "Heading1", BasedOn: "Parent"}
	child.ParaAttrs.SetAlignment(ir.AlignLeft)
	child.ParaAttrs.SetSpacing(ir.Spacing{})
	child.ParaAttrs.SetIndent(ir.Indent{})
	child.ParaAttrs.SetKeepWithNext(false)
	child.ParaAttrs.SetOutlineLevel(0)
	child.RunAttrs.SetColor(ir.Color{})
	document.Styles.Named = map[string]ir.Style{"Parent": parent, "Heading1": child}
	paragraph := editingParagraph("p1", "Heading with explicit resets")
	paragraph.Style = "Heading1"
	document.Sections[0].Blocks = []ir.Block{paragraph}
	encoded, err := json.Marshal(document)
	require.NoError(t, err)
	var cloned ir.Document
	require.NoError(t, json.Unmarshal(encoded, &cloned))
	resolved, ok := cloned.Styles.ResolveStyle("Heading1")
	require.True(t, ok)
	require.Equal(t, ir.AlignLeft, resolved.ParaAttrs.Align)
	require.True(t, resolved.ParaAttrs.AlignmentSpecified())
	require.Zero(t, resolved.ParaAttrs.Spacing.Before)
	require.Zero(t, resolved.ParaAttrs.Indent.Left)
	require.False(t, resolved.ParaAttrs.KeepWithNext)
	require.True(t, resolved.RunAttrs.ColorSpecified())
	require.Equal(t, ir.Color{}, resolved.RunAttrs.Color)
	content, report, err := lean.SaveDOCX(&cloned)
	require.NoError(t, err)
	require.True(t, report.Editable)
	styles := editingPart(t, content, "word/styles.xml")
	for _, expected := range []string{`<w:jc w:val="left"/>`, `<w:outlineLvl w:val="0"/>`, `<w:keepNext w:val="0"/>`, `<w:color w:val="000000"/>`} {
		require.Contains(t, styles, expected)
	}
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	resolved, ok = reopened.Styles.ResolveStyle("Heading1")
	require.True(t, ok)
	require.True(t, resolved.ParaAttrs.OutlineLevelSpecified())
	require.True(t, resolved.RunAttrs.ColorSpecified())
	require.False(t, resolved.ParaAttrs.KeepWithNext)
}

func TestEditingListStartsAndRestartProduceValidOOXML(t *testing.T) {
	document := ir.NewDocument()
	document.Styles.Numbering = []ir.NumberingDef{{ID: "browser-list", Levels: []ir.NumberingLevel{{Level: 0, Format: ir.NumFormatDecimal, Text: "%1.", Start: 3, Indent: ir.Indent{Left: 36, Hanging: 18}, RunAttrs: ir.RunAttrs{Bold: true, FontName: "Arial"}}}}}
	first := editingParagraph("one", "Third item")
	first.Numbering = &ir.NumberingRef{ID: "browser-list"}
	restart := 1
	second := editingParagraph("two", "Restarted item")
	second.Numbering = &ir.NumberingRef{ID: "browser-list", Restart: &restart}
	third := editingParagraph("three", "Continued item")
	third.Numbering = &ir.NumberingRef{ID: "browser-list"}
	document.Sections[0].Blocks = []ir.Block{first, second, third}
	content, report, err := lean.SaveDOCX(document)
	require.NoError(t, err)
	require.True(t, report.Editable)
	require.NotContains(t, editingPart(t, content, "word/document.xml"), `w:numId w:val="browser-list"`)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	require.Len(t, reopened.Styles.Numbering, 2)
	paragraphs := reopened.Sections[0].Blocks
	one := paragraphs[0].(*ir.Paragraph).Numbering
	two := paragraphs[1].(*ir.Paragraph).Numbering
	three := paragraphs[2].(*ir.Paragraph).Numbering
	require.NotEqual(t, one.ID, two.ID)
	require.Equal(t, two.ID, three.ID)
	require.Equal(t, 3, reopened.Styles.Numbering[0].Levels[0].Start)
	require.Equal(t, 1, reopened.Styles.Numbering[1].Levels[0].Start)
	require.Equal(t, ir.Indent{Left: 36, Hanging: 18}, reopened.Styles.Numbering[0].Levels[0].Indent)
	require.True(t, reopened.Styles.Numbering[0].Levels[0].RunAttrs.Bold)
	require.Equal(t, "browser-list", first.Numbering.ID)
}

func TestEditingSectionsKeepIndependentHeadersFootersAndColumns(t *testing.T) {
	document := ir.NewDocument()
	properties := document.Sections[0].Properties
	header := func(kind ir.HeaderFooterKind, text string) *ir.HeaderFooter {
		return &ir.HeaderFooter{Kind: kind, Blocks: []ir.Block{editingParagraph(text, text)}}
	}
	document.Sections = []ir.Section{
		{Properties: properties, Blocks: []ir.Block{editingParagraph("first", "First section")}, Header: header(ir.HeaderFooterDefault, "First header"), Headers: []*ir.HeaderFooter{header(ir.HeaderFooterFirst, "Title header"), header(ir.HeaderFooterEven, "Even header")}, Footer: header(ir.HeaderFooterDefault, "First footer"), Columns: []ir.Column{{Width: 216, Spacing: 36}, {Width: 216}}},
		{Properties: properties, Blocks: []ir.Block{editingParagraph("second", "Second section")}, Header: header(ir.HeaderFooterDefault, "Second header"), Footer: header(ir.HeaderFooterDefault, "Second footer")},
	}
	encoded, err := json.Marshal(document)
	require.NoError(t, err)
	var cloned ir.Document
	require.NoError(t, json.Unmarshal(encoded, &cloned))
	require.Len(t, cloned.Sections[0].Headers, 2)
	content, report, err := lean.SaveDOCX(&cloned)
	require.NoError(t, err)
	require.True(t, report.Editable)
	require.Contains(t, editingPart(t, content, "word/settings.xml"), "evenAndOddHeaders")
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	require.Len(t, reopened.Sections, 2)
	for index, text := range []string{"First header", "Second header"} {
		require.Equal(t, text, reopened.Sections[index].Header.Blocks[0].(*ir.Paragraph).Runs[0].Text)
		require.Len(t, reopened.Sections[index].Blocks, 1)
	}
	require.Len(t, reopened.Sections[0].Headers, 2)
	require.Equal(t, document.Sections[0].Columns, reopened.Sections[0].Columns)
	require.Equal(t, "Second footer", reopened.Sections[1].Footer.Blocks[0].(*ir.Paragraph).Runs[0].Text)
}

func TestEditingFootnotesRetainTheirOwnReferencesAndHyperlinks(t *testing.T) {
	document := ir.NewDocument()
	first := editingParagraph("first", "First note anchor")
	second := editingParagraph("second", "Second note anchor")
	first.Footnotes = []ir.Footnote{{ID: "first-note", Blocks: []ir.Block{&ir.Paragraph{ID: "note-one", Runs: []ir.Run{{Text: "First note", Hyperlink: &ir.Hyperlink{URL: "https://example.org/reference"}}}}}}}
	second.Footnotes = []ir.Footnote{{ID: "second-note", Blocks: []ir.Block{editingParagraph("note-two", "Second note")}}}
	document.Sections[0].Blocks = []ir.Block{first, second}
	content, report, err := lean.SaveDOCX(document)
	require.NoError(t, err)
	require.True(t, report.Editable)
	xml := editingPart(t, content, "word/document.xml")
	require.Contains(t, xml, `footnoteReference w:id="1"`)
	require.Contains(t, xml, `footnoteReference w:id="2"`)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	one := reopened.Sections[0].Blocks[0].(*ir.Paragraph).Footnotes[0]
	two := reopened.Sections[0].Blocks[1].(*ir.Paragraph).Footnotes[0]
	require.Equal(t, "First note", one.Blocks[0].(*ir.Paragraph).Runs[0].Text)
	require.Equal(t, "Second note", two.Blocks[0].(*ir.Paragraph).Runs[0].Text)
	require.Equal(t, "https://example.org/reference", one.Blocks[0].(*ir.Paragraph).Runs[0].Hyperlink.URL)
}
