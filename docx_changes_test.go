package lean_test

import (
	"os"
	"testing"

	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/require"
)

func TestTrackedTextSurvivesWordcraftDOCXExportAndResolution(t *testing.T) {
	input, err := os.ReadFile("testdata/wordcraft/handbook.docx")
	require.NoError(t, err)
	original, report, err := lean.OpenDOCX(input)
	require.NoError(t, err)
	require.True(t, report.Editable)
	var paragraph *ir.Paragraph
	for _, block := range original.Sections[0].Blocks {
		if value, ok := block.(*ir.Paragraph); ok && len(value.Runs) > 0 && len(value.Runs[0].Text) > 0 {
			paragraph = value
			break
		}
	}
	require.NotNil(t, paragraph)
	before := ""
	for _, run := range paragraph.Runs {
		before += run.Text
	}
	edited, err := lean.ApplyDOCXCommands(original, []lean.DocumentCommand{{ID: "text.suggest", Block: paragraph.ID, Start: 0, End: 1, Text: "Updated", Change: &ir.Change{Author: "Reviewer", Date: "2026-10-31T13:00:00Z"}}})
	require.NoError(t, err)
	output, report, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	require.True(t, report.Editable)
	xml := editingPart(t, output, "word/document.xml")
	require.Contains(t, xml, "<w:ins ")
	require.Contains(t, xml, "<w:del ")
	require.Contains(t, xml, "<w:delText")
	reopened, report, err := lean.OpenDOCX(output)
	require.NoError(t, err)
	require.True(t, report.Editable)
	var ids []int
	for _, block := range reopened.Sections[0].Blocks {
		if p, ok := block.(*ir.Paragraph); ok {
			for _, run := range p.Runs {
				if run.Change != nil {
					ids = append(ids, run.Change.ID)
					require.Equal(t, "Reviewer", run.Change.Author)
				}
			}
		}
	}
	require.Len(t, ids, 2)
	for _, decision := range []string{"change.accept", "change.reject"} {
		var commands []lean.DocumentCommand
		for _, id := range ids {
			value := id
			commands = append(commands, lean.DocumentCommand{ID: decision, ChangeID: &value})
		}
		resolved, err := lean.ApplyDOCXCommands(reopened, commands)
		require.NoError(t, err)
		content, _, err := lean.SaveDOCX(resolved)
		require.NoError(t, err)
		require.NotContains(t, editingPart(t, content, "word/document.xml"), "<w:ins ")
		require.NotContains(t, editingPart(t, content, "word/document.xml"), "<w:del ")
		if decision == "change.reject" {
			found := false
			for _, block := range resolved.Sections[0].Blocks {
				if p, ok := block.(*ir.Paragraph); ok {
					text := ""
					for _, run := range p.Runs {
						text += run.Text
					}
					if text == before {
						found = true
					}
				}
			}
			require.True(t, found, "Reject restores the original text")
		}
	}
}

func TestTrackedFormattingPreservesPreviousProperties(t *testing.T) {
	input, err := os.ReadFile("testdata/wordcraft/handbook.docx")
	require.NoError(t, err)
	document, _, err := lean.OpenDOCX(input)
	require.NoError(t, err)
	var paragraph *ir.Paragraph
	for _, block := range document.Sections[0].Blocks {
		if value, ok := block.(*ir.Paragraph); ok && len(value.Runs) > 0 && len(value.Runs[0].Text) > 0 {
			paragraph = value
			break
		}
	}
	require.NotNil(t, paragraph)
	previous := paragraph.Runs[0].Attrs
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "text.format.suggest", Block: paragraph.ID, Start: 0, End: 1, Run: &ir.RunAttrs{Bold: true}, Change: &ir.Change{Author: "Reviewer", Date: "2026-10-31T13:00:00Z"}}})
	require.NoError(t, err)
	content, _, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	require.Contains(t, editingPart(t, content, "word/document.xml"), "<w:rPrChange ")
	reopened, report, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	require.True(t, report.Editable)
	var revision *ir.Change
	for _, block := range reopened.Sections[0].Blocks {
		if p, ok := block.(*ir.Paragraph); ok {
			for _, run := range p.Runs {
				if run.Change != nil {
					revision = run.Change
				}
			}
		}
	}
	require.NotNil(t, revision)
	require.Equal(t, "format", revision.Type)
	require.Equal(t, previous, *revision.Previous)
	rejected, err := lean.ApplyDOCXCommands(reopened, []lean.DocumentCommand{{ID: "change.reject", ChangeID: &revision.ID}})
	require.NoError(t, err)
	output, _, err := lean.SaveDOCX(rejected)
	require.NoError(t, err)
	require.NotContains(t, editingPart(t, output, "word/document.xml"), "<w:rPrChange ")
}

func TestTrackedTableRowsSurviveDOCXAndReject(t *testing.T) {
	input, err := os.ReadFile("testdata/wordcraft/handbook.docx")
	require.NoError(t, err)
	document, _, err := lean.OpenDOCX(input)
	require.NoError(t, err)
	edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{{ID: "table.insert.suggest", At: 0, Rows: 2, Columns: 2, Change: &ir.Change{Author: "Reviewer", Date: "2026-10-31T13:00:00Z"}}})
	require.NoError(t, err)
	content, _, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	require.Contains(t, editingPart(t, content, "word/document.xml"), "<w:trPr><w:ins ")
	reopened, report, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	require.True(t, report.Editable)
	table, ok := reopened.Sections[0].Blocks[0].(*ir.Table)
	require.True(t, ok)
	require.Len(t, table.Rows, 2)
	var commands []lean.DocumentCommand
	for _, row := range table.Rows {
		require.NotNil(t, row.Change)
		id := row.Change.ID
		commands = append(commands, lean.DocumentCommand{ID: "change.reject", ChangeID: &id})
	}
	rejected, err := lean.ApplyDOCXCommands(reopened, commands)
	require.NoError(t, err)
	_, ok = rejected.Sections[0].Blocks[0].(*ir.Table)
	require.False(t, ok, "Rejecting all inserted rows removes the proposed table")
	_, _, err = lean.SaveDOCX(rejected)
	require.NoError(t, err)
}
