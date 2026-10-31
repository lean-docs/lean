package lean_test

import (
	"encoding/json"
	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/require"
	"math"
	"testing"
)

func TestDocumentCommandsEditCanonicalContentAtomically(t *testing.T) {
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{editingParagraph("intro", "Hello 😀 world")}
	document.Styles.Named["Title"] = ir.Style{ID: "Title", RunAttrs: ir.RunAttrs{Bold: true, FontSize: 24}}
	var plain ir.RunAttrs
	plain.SetBold(false)
	plain.SetColor(ir.Color{})
	header := &ir.HeaderFooter{Type: ir.HeaderFooterDefault, Blocks: []ir.Block{editingParagraph("header", "Header")}}
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{
		{ID: "text.replace", Block: "intro", Start: 9, End: 14, Text: "document"},
		{ID: "text.format", Block: "intro", Start: 0, End: 5, Run: &plain},
		{ID: "text.link", Block: "intro", Start: 0, End: 5, Hyperlink: &ir.Hyperlink{URL: "https://example.org"}},
		{ID: "paragraph.style", Block: "intro", Style: "Title"},
		{ID: "header.set", HeaderFooter: header},
		{ID: "text.replace", Part: "header", Block: "header", Start: 0, End: 6, Text: "Updated header"},
		{ID: "table.insert", At: 1, Rows: 2, Columns: 2},
	})
	require.NoError(t, err)
	require.Equal(t, "Header", header.Blocks[0].(*ir.Paragraph).Runs[0].Text)
	require.Equal(t, "Hello 😀 world", document.Sections[0].Blocks[0].(*ir.Paragraph).Runs[0].Text)
	table := edited.Sections[0].Blocks[1].(*ir.Table)
	require.Len(t, table.Rows, 2)
	edited, err = lean.ApplyDOCXCommands(edited, []lean.DocumentCommand{{ID: "table.row.insert", Block: table.ID, At: 1}, {ID: "table.column.insert", Block: table.ID, At: 1}, {ID: "table.row.delete", Block: table.ID, At: 2}, {ID: "table.column.delete", Block: table.ID, At: 2}})
	require.NoError(t, err)
	content, report, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	require.True(t, report.Editable)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	paragraph := reopened.Sections[0].Blocks[0].(*ir.Paragraph)
	text := ""
	for _, run := range paragraph.Runs {
		text += run.Text
	}
	require.Equal(t, "Hello 😀 document", text)
	require.Equal(t, "Title", paragraph.Style)
	require.True(t, paragraph.Runs[0].Attrs.BoldSpecified())
	require.False(t, paragraph.Runs[0].Attrs.Bold)
	require.Equal(t, "https://example.org", paragraph.Runs[0].Hyperlink.URL)
	require.Equal(t, "Updated header", reopened.Sections[0].Header.Blocks[0].(*ir.Paragraph).Runs[0].Text)
	require.Len(t, reopened.Sections[0].Blocks[1].(*ir.Table).Rows, 2)
}

func TestDocumentCommandsRejectPartialAndInvalidEdits(t *testing.T) {
	document := ir.NewDocument()
	document.Sections[0].Blocks = []ir.Block{editingParagraph("intro", "😀 original")}
	before, err := json.Marshal(document)
	require.NoError(t, err)
	for _, commands := range [][]lean.DocumentCommand{
		{{ID: "text.replace", Block: "intro", Start: 0, End: 2, Text: "Changed"}, {ID: "page.layout", Page: &ir.PageProperties{Width: 100, Height: 100, MarginLeft: 90, MarginRight: 90}}},
		{{ID: "text.replace", Block: "intro", Start: 1, End: 2, Text: "Broken"}},
		{{ID: "paragraph.split", Block: "intro", At: -1}},
		{{ID: "table.insert", Rows: 1000, Columns: 1000}},
		{{ID: "paragraph.style", Block: "intro", Style: "Missing"}},
		{{ID: "page.columns", ColumnsLayout: []ir.Column{{Width: 10000}}}},
	} {
		edited, err := lean.ApplyDOCXCommands(document, commands)
		require.ErrorIs(t, err, lean.ErrInvalidCommand)
		require.Nil(t, edited)
		after, err := json.Marshal(document)
		require.NoError(t, err)
		require.Equal(t, string(before), string(after))
	}
}

func TestDocumentCommandsPreserveParagraphIdentityAcrossSplitAndJoin(t *testing.T) {
	document := ir.NewDocument()
	paragraph := editingParagraph("intro", "First second")
	paragraph.Numbering = &ir.NumberingRef{ID: "list"}
	document.Styles.Numbering = []ir.NumberingDef{{ID: "list", Levels: []ir.NumberingLevel{{Format: ir.NumFormatBullet, Text: "•", Start: 1}}}}
	document.Sections[0].Blocks = []ir.Block{paragraph}
	split, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "paragraph.split", Block: "intro", At: 6}})
	require.NoError(t, err)
	require.Len(t, split.Sections[0].Blocks, 2)
	require.Equal(t, "intro", split.Sections[0].Blocks[0].BlockID())
	require.NotEqual(t, "intro", split.Sections[0].Blocks[1].BlockID())
	require.NotNil(t, split.Sections[0].Blocks[1].(*ir.Paragraph).Numbering)
	joined, err := lean.ApplyDOCXCommands(split, []lean.DocumentCommand{{ID: "paragraph.join", Block: "intro"}})
	require.NoError(t, err)
	require.Len(t, joined.Sections[0].Blocks, 1)
	require.Equal(t, "intro", joined.Sections[0].Blocks[0].BlockID())
	content, _, err := lean.SaveDOCX(joined)
	require.NoError(t, err)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	text := ""
	for _, run := range reopened.Sections[0].Blocks[0].(*ir.Paragraph).Runs {
		text += run.Text
	}
	require.Equal(t, "First second", text)
}

func TestDocumentCommandsCountInlineBreaksInTextPositions(t *testing.T) {
	document := lean.NewDOCX("")
	paragraph := document.Sections[0].Blocks[0].(*ir.Paragraph)
	paragraph.Runs = []ir.Run{{Text: "First", Break: ir.BreakLine}, {Text: "second"}}
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "text.replace", Block: paragraph.ID, Start: 6, End: 12, Text: "last"}})
	require.NoError(t, err)
	content, _, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	runs := reopened.Sections[0].Blocks[0].(*ir.Paragraph).Runs
	require.Equal(t, "First", runs[0].Text)
	require.Equal(t, ir.BreakLine, runs[0].Break)
	require.Equal(t, "last", runs[len(runs)-1].Text)
}

func TestDocumentCommandsKeepStylesUsedByHeaders(t *testing.T) {
	document := lean.NewDOCX("Body")
	document.Sections[0].Header = &ir.HeaderFooter{Blocks: []ir.Block{&ir.Paragraph{ID: "header", Style: "Title", Runs: []ir.Run{{Text: "Header"}}}}}
	_, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "style.remove", Style: "Title"}})
	require.ErrorContains(t, err, "still used")
	require.Contains(t, document.Styles.Named, "Title")
}

func TestDocumentCommandsPreserveFootnoteAnchors(t *testing.T) {
	document := lean.NewDOCX("First second")
	paragraph := document.Sections[0].Blocks[0].(*ir.Paragraph)
	position := 6
	var err error
	document, err = lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "footnote.set", Block: paragraph.ID, Footnote: &ir.Footnote{ID: "note", Offset: &position, Blocks: []ir.Block{editingParagraph("note-text", "Reference")}}}})
	require.NoError(t, err)
	content, _, err := lean.SaveDOCX(document)
	require.NoError(t, err)
	document, _, err = lean.OpenDOCX(content)
	require.NoError(t, err)
	paragraph = document.Sections[0].Blocks[0].(*ir.Paragraph)
	require.Equal(t, 6, *paragraph.Footnotes[0].Offset)
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "text.replace", Block: paragraph.ID, Start: 0, End: 0, Text: "New "}, {ID: "paragraph.split", Block: paragraph.ID, At: 10}})
	require.NoError(t, err)
	first := edited.Sections[0].Blocks[0].(*ir.Paragraph)
	second := edited.Sections[0].Blocks[1].(*ir.Paragraph)
	require.Empty(t, first.Footnotes)
	require.Equal(t, 0, *second.Footnotes[0].Offset)
	edited, err = lean.ApplyDOCXCommands(edited, []lean.DocumentCommand{{ID: "paragraph.join", Block: first.ID}})
	require.NoError(t, err)
	content, _, err = lean.SaveDOCX(edited)
	require.NoError(t, err)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	require.Equal(t, 10, *reopened.Sections[0].Blocks[0].(*ir.Paragraph).Footnotes[0].Offset)
}

func TestSectionSplitDoesNotShareBlockStorage(t *testing.T) {
	for _, command := range []lean.DocumentCommand{
		{ID: "paragraph.insert", At: 1, Text: "Added"},
		{ID: "paragraph.split", Block: "one", At: 1},
		{ID: "table.insert", At: 1, Rows: 1, Columns: 1},
	} {
		t.Run(command.ID, func(t *testing.T) {
			document := lean.NewDOCX("First")
			document.Sections[0].Blocks = []ir.Block{editingParagraph("one", "First"), editingParagraph("two", "Second"), editingParagraph("three", "Third"), editingParagraph("four", "Fourth")}
			edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "section.insert", At: 2}, command})
			require.NoError(t, err)
			require.Equal(t, "Third", edited.Sections[1].Blocks[0].(*ir.Paragraph).Runs[0].Text)
			require.Equal(t, "Fourth", edited.Sections[1].Blocks[1].(*ir.Paragraph).Runs[0].Text)
			content, _, err := lean.SaveDOCX(edited)
			require.NoError(t, err)
			reopened, _, err := lean.OpenDOCX(content)
			require.NoError(t, err)
			require.Equal(t, "Third", reopened.Sections[1].Blocks[0].(*ir.Paragraph).Runs[0].Text)
		})
	}
}

func TestCanonicalSaveRejectsNonFiniteColumnSpacing(t *testing.T) {
	for _, spacing := range []float64{math.NaN(), math.Inf(1), math.Inf(-1)} {
		document := lean.NewDOCX("Columns")
		document.Sections[0].Columns = []ir.Column{{Width: 120, Spacing: spacing}, {Width: 120}}
		content, report, err := lean.SaveDOCX(document)
		require.ErrorIs(t, err, lean.ErrUnsupportedDocument)
		require.Nil(t, content)
		require.False(t, report.Editable)
		require.Contains(t, report.Unsupported, "page-columns")
	}
}
