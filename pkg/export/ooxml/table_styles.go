package ooxml

import (
	"slices"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

func writeTableStyle(output *strings.Builder, style *ir.TableStyle) {
	output.WriteString("<w:tblPr>")
	if style.RowBandSize > 0 {
		output.WriteString(`<w:tblStyleRowBandSize w:val="` + strconv.Itoa(style.RowBandSize) + `"/>`)
	}
	if style.ColumnBandSize > 0 {
		output.WriteString(`<w:tblStyleColBandSize w:val="` + strconv.Itoa(style.ColumnBandSize) + `"/>`)
	}
	writeTableBorders(output, style.Borders)
	output.WriteString("</w:tblPr>")
	writeStyleCell(output, style.Cell)
	keys := make([]string, 0, len(style.Conditions))
	for key := range style.Conditions {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		region := style.Conditions[key]
		output.WriteString(`<w:tblStylePr w:type="` + escape(key) + `">`)
		writeParagraphProperties(output, &ir.Paragraph{Para: region.ParaAttrs})
		writeRunProperties(output, region.RunAttrs)
		writeStyleCell(output, region)
		output.WriteString("</w:tblStylePr>")
	}
}

func writeStyleCell(output *strings.Builder, cell ir.TableCellStyle) {
	if cell.Borders == (ir.CellBorders{}) && cell.Shading == (ir.Color{}) && cell.Padding == (ir.Padding{}) {
		return
	}
	output.WriteString("<w:tcPr>")
	writeCellBorders(output, cell.Borders)
	if cell.Shading != (ir.Color{}) {
		fill := color(cell.Shading)
		if cell.Shading.None {
			fill = "auto"
		}
		output.WriteString(`<w:shd w:val="clear" w:fill="` + fill + `"/>`)
	}
	writeCellPadding(output, cell.Padding)
	output.WriteString("</w:tcPr>")
}
