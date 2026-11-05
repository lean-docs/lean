package lean_test

import (
	lean "github.com/lean-docs/lean"
	"github.com/lean-docs/lean/pkg/ir"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTableMergeSplitSurviveDOCX(t *testing.T) {
	doc := ir.NewDocument()
	doc.Sections[0].Blocks = []ir.Block{editingParagraph("intro", "Table")}
	doc, err := lean.ApplyDOCXCommands(doc, []lean.DocumentCommand{{ID: "table.insert", At: 1, Rows: 3, Columns: 3}})
	require.NoError(t, err)
	table := doc.Sections[0].Blocks[1].(*ir.Table)
	var edits []lean.DocumentCommand
	for row := 0; row < 2; row++ {
		for column := 0; column < 2; column++ {
			edits = append(edits, lean.DocumentCommand{ID: "text.replace", Block: table.Rows[row].Cells[column].Blocks[0].BlockID(), Text: "Cell content"})
		}
	}
	edits = append(edits, lean.DocumentCommand{ID: "table.merge", Block: table.ID, CellRange: &lean.TableCellRange{Row: 0, Column: 0, Rows: 2, Columns: 2}})
	merged, err := lean.ApplyDOCXCommands(doc, edits)
	require.NoError(t, err)
	require.Len(t, table.Rows[0].Cells, 3)
	content, report, err := lean.SaveDOCX(merged)
	require.NoError(t, err)
	require.True(t, report.Editable, report.Unsupported)
	reopened, _, err := lean.OpenDOCX(content)
	require.NoError(t, err)
	restored := reopened.Sections[0].Blocks[1].(*ir.Table)
	require.Equal(t, 2, restored.Rows[0].Cells[0].RowSpan)
	require.Equal(t, 2, restored.Rows[0].Cells[0].ColSpan)
	require.Len(t, restored.Rows[0].Cells[0].Blocks, 4)
	split, err := lean.ApplyDOCXCommands(reopened, []lean.DocumentCommand{{ID: "table.split", Block: restored.ID, CellRange: &lean.TableCellRange{Row: 0, Column: 0}}})
	require.NoError(t, err)
	after := split.Sections[0].Blocks[1].(*ir.Table)
	require.Len(t, after.Rows[0].Cells, 3)
	require.Len(t, after.Rows[1].Cells, 3)
	require.Len(t, after.Rows[0].Cells[0].Blocks, 4)
	_, err = lean.ApplyDOCXCommands(doc, []lean.DocumentCommand{{ID: "table.merge", Block: table.ID, CellRange: &lean.TableCellRange{Row: 0, Column: 2, Rows: 2, Columns: 2}}})
	require.Error(t, err)
}
