package lean

import (
	"errors"
	"github.com/lean-docs/lean/pkg/ir"
)

type TableFormat struct {
	ResetStyle    bool            `json:"reset_style,omitempty"`
	ColumnWidths  *[]float64      `json:"column_widths,omitempty"`
	Align         *ir.Alignment   `json:"align,omitempty"`
	Padding       *ir.Padding     `json:"padding,omitempty"`
	VerticalAlign *ir.VAlignment  `json:"vertical_align,omitempty"`
	Shading       *ir.Color       `json:"shading,omitempty"`
	Borders       *ir.CellBorders `json:"borders,omitempty"`
	RepeatHeader  *bool           `json:"repeat_header,omitempty"`
}

func applyTableFormat(table *ir.Table, format *TableFormat) error {
	if format == nil {
		return errors.New("table formatting is required")
	}
	if hasPendingChanges([]ir.Block{table}) {
		return errors.New("resolve tracked changes before formatting a table")
	}
	if format.ResetStyle {
		table.Borders = ir.TableBorders{}
		table.Shading = ir.Color{}
		for rowIndex := range table.Rows {
			for cellIndex := range table.Rows[rowIndex].Cells {
				cell := &table.Rows[rowIndex].Cells[cellIndex]
				cell.Borders = ir.CellBorders{}
				cell.Shading = ir.Color{}
				cell.Padding = ir.Padding{}
				cell.VAlign = ir.VAlignTop
			}
		}
	}
	if format.ColumnWidths != nil {
		count := 0
		if len(table.Rows) > 0 {
			for _, cell := range table.Rows[0].Cells {
				count += max(1, cell.ColSpan)
			}
		}
		if len(*format.ColumnWidths) != count {
			return errors.New("provide one width for each table column")
		}
		table.ColumnWidths = append([]float64(nil), (*format.ColumnWidths)...)
	}
	if format.Align != nil {
		table.Align = *format.Align
	}
	for index := range table.Rows {
		row := &table.Rows[index]
		if format.RepeatHeader != nil {
			row.IsHeader = index == 0 && *format.RepeatHeader
		}
		for index := range row.Cells {
			cell := &row.Cells[index]
			if format.Padding != nil {
				cell.Padding = *format.Padding
			}
			if format.VerticalAlign != nil {
				cell.VAlign = *format.VerticalAlign
			}
			if format.Shading != nil {
				cell.Shading = *format.Shading
			}
			if format.Borders != nil {
				cell.Borders = *format.Borders
			}
		}
	}
	return nil
}
