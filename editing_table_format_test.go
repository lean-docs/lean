package lean_test

import (
	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/require"
	"os"
	"testing"
)

func TestTableFormatCommandsRoundTrip(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{editingParagraph("intro", "Table")}
	doc, err := lean.ApplyDOCXCommands(doc, []lean.DocumentCommand{{ID: "table.insert", At: 1, Rows: 2, Columns: 2}})
	require.NoError(t, err)
	table := doc.Sections[0].Blocks[1].(*ir.Table)
	widths := []float64{110, 130}
	padding := ir.Padding{Top: 5, Bottom: 5, Left: 6, Right: 6}
	align := ir.AlignCenter
	repeat := true
	edited, err := lean.ApplyDOCXCommands(doc, []lean.DocumentCommand{{ID: "table.format", Block: table.ID, TableFormat: &lean.TableFormat{ColumnWidths: &widths, Padding: &padding, Align: &align, RepeatHeader: &repeat}}})
	require.NoError(t, err)
	require.NotEqual(t, widths, table.ColumnWidths)
	content, report, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	require.True(t, report.Editable)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	restored := reopened.Sections[0].Blocks[1].(*ir.Table)
	require.Equal(t, widths, restored.ColumnWidths)
	require.Equal(t, padding, restored.Rows[0].Cells[0].Padding)
	require.Equal(t, align, restored.Align)
	reset, err := lean.ApplyDOCXCommands(edited, []lean.DocumentCommand{{ID: "table.format", Block: table.ID, TableFormat: &lean.TableFormat{ResetStyle: true}}})
	require.NoError(t, err)
	resetTable := reset.Sections[0].Blocks[1].(*ir.Table)
	require.Equal(t, ir.Padding{}, resetTable.Rows[0].Cells[0].Padding)
	require.Equal(t, widths, resetTable.ColumnWidths)
	require.True(t, restored.Rows[0].IsHeader)
	require.False(t, restored.Rows[1].IsHeader)
	bad := []float64{1}
	_, err = lean.ApplyDOCXCommands(doc, []lean.DocumentCommand{{ID: "table.format", Block: table.ID, TableFormat: &lean.TableFormat{ColumnWidths: &bad}}})
	require.Error(t, err)
}

func TestPictureAlternativeTextRoundTrip(t *testing.T) {
	input, err := os.ReadFile("testdata/reference-layouts/nested-textbox.docx")
	require.NoError(t, err)
	doc, _, err := lean.OpenDOCX(input)
	require.NoError(t, err)
	picture := doc.Sections[0].Blocks[1].(*ir.Shape).Blocks[1].(*ir.Image)
	alt := "Reference picture"
	edited, err := lean.ApplyDOCXCommands(doc, []lean.DocumentCommand{{ID: "drawing.layout", Block: picture.ID, Drawing: &lean.DrawingLayout{Alt: &alt}}})
	require.NoError(t, err)
	content, _, err := lean.SaveDOCX(edited)
	require.NoError(t, err)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	restored := reopened.Sections[0].Blocks[1].(*ir.Shape).Blocks[1].(*ir.Image)
	require.Equal(t, alt, restored.Alt)
	require.Equal(t, picture.Data, restored.Data)
	require.Equal(t, picture.Anchor, restored.Anchor)
}
