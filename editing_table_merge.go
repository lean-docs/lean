package lean

import (
	"errors"
	"github.com/lean-docs/lean/pkg/ir"
	"slices"
)

type TableCellRange struct {
	Row     int `json:"row"`
	Column  int `json:"column"`
	Rows    int `json:"rows"`
	Columns int `json:"columns"`
}

func applyTableMerge(table *ir.Table, area *TableCellRange, split bool) error {
	if area == nil || area.Row < 0 || area.Column < 0 || area.Row >= len(table.Rows) {
		return errors.New("select a valid cell range")
	}
	if hasPendingChanges([]ir.Block{table}) {
		return errors.New("resolve tracked changes before merging cells")
	}
	if split {
		index := -1
		column := 0
		for i, cell := range table.Rows[area.Row].Cells {
			if column == area.Column {
				index = i
				break
			}
			column += max(1, cell.ColSpan)
		}
		if index < 0 {
			return errors.New("select the origin of a merged cell")
		}
		cell := table.Rows[area.Row].Cells[index]
		height, width := max(1, cell.RowSpan), max(1, cell.ColSpan)
		if area.Row+height > len(table.Rows) {
			return errors.New("invalid merged cell extent")
		}
		for row := area.Row; row < area.Row+height; row++ {
			position, logical := -1, 0
			for i, candidate := range table.Rows[row].Cells {
				if logical == area.Column {
					position = i
					break
				}
				logical += max(1, candidate.ColSpan)
			}
			if position < 0 || max(1, table.Rows[row].Cells[position].ColSpan) != width {
				return errors.New("invalid merged cell continuation")
			}
			cells := make([]ir.TableCell, width)
			for i := range cells {
				cells[i] = newEditingCell()
			}
			if row == area.Row {
				cells[0] = cell
				cells[0].RowSpan = 1
				cells[0].ColSpan = 1
			}
			table.Rows[row].Cells = slices.Replace(table.Rows[row].Cells, position, position+1, cells...)
		}
		return nil
	}
	if area.Rows < 1 || area.Columns < 1 || area.Row+area.Rows > len(table.Rows) {
		return errors.New("select a nonempty rectangular range")
	}
	for _, row := range table.Rows {
		for _, cell := range row.Cells {
			if cell.ColSpan > 1 || cell.RowSpan > 1 {
				return errors.New("split existing merged cells before merging a new range")
			}
		}
	}
	var content []ir.Block
	for row := area.Row; row < area.Row+area.Rows; row++ {
		if area.Column+area.Columns > len(table.Rows[row].Cells) {
			return errors.New("cell range exceeds the table")
		}
		for _, cell := range table.Rows[row].Cells[area.Column : area.Column+area.Columns] {
			content = append(content, cell.Blocks...)
		}
	}
	for row := area.Row; row < area.Row+area.Rows; row++ {
		cell := table.Rows[row].Cells[area.Column]
		cell.ColSpan = area.Columns
		cell.RowSpan = 1
		cell.Blocks = []ir.Block{newEditingParagraph("")}
		if row == area.Row {
			cell.RowSpan = area.Rows
			cell.Blocks = content
		}
		table.Rows[row].Cells = slices.Replace(table.Rows[row].Cells, area.Column, area.Column+area.Columns, cell)
	}
	return nil
}
