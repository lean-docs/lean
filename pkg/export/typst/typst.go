package typst

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

func Export(doc *ir.Document) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("typst exporter: document is nil")
	}
	state := exporter{document: doc}
	if len(doc.Sections) == 0 {
		state.output.WriteString("#set page()\n")
	}
	for _, section := range doc.Sections {
		state.writeSection(section)
	}
	return []byte(state.output.String()), nil
}

type exporter struct {
	document *ir.Document
	output   strings.Builder
}

func (e *exporter) writeSection(section ir.Section) {
	e.writePageSetup(section)
	if len(section.Columns) > 1 {
		e.output.WriteString("#columns(" + strconv.Itoa(len(section.Columns)) + ")[\n")
	}
	e.writeBlocks(section.Blocks)
	if len(section.Columns) > 1 {
		e.output.WriteString("]\n")
	}
}

func (e *exporter) writePageSetup(section ir.Section) {
	properties := section.Properties
	var options []string
	if properties.Width > 0 {
		options = append(options, "width: "+points(properties.Width))
	}
	if properties.Height > 0 {
		options = append(options, "height: "+points(properties.Height))
	}
	if properties.MarginTop > 0 || properties.MarginRight > 0 || properties.MarginBottom > 0 || properties.MarginLeft > 0 {
		options = append(options, "margin: (top: "+points(properties.MarginTop)+", right: "+points(properties.MarginRight)+", bottom: "+points(properties.MarginBottom)+", left: "+points(properties.MarginLeft)+")")
	}
	if section.Header != nil {
		options = append(options, "header: ["+blocksText(section.Header.Blocks)+"]")
	}
	if section.Footer != nil {
		options = append(options, "footer: ["+blocksText(section.Footer.Blocks)+"]")
	}
	e.output.WriteString("#set page(" + strings.Join(options, ", ") + ")\n")
}

func (e *exporter) writeBlocks(blocks []ir.Block) {
	for _, block := range blocks {
		e.writeBlock(block)
	}
}

func (e *exporter) writeBlock(block ir.Block) {
	switch value := block.(type) {
	case *ir.Paragraph:
		e.writeParagraph(value)
	case *ir.Table:
		e.writeTable(value)
	case *ir.Image:
		e.writeImage(value)
	}
}

func (e *exporter) writeParagraph(paragraph *ir.Paragraph) {
	if level := headingLevel(paragraph.Style); level > 0 {
		e.output.WriteString(strings.Repeat("=", level) + " ")
	} else if paragraph.Numbering != nil {
		e.output.WriteString(strings.Repeat("  ", paragraph.Numbering.Level))
		if e.numberingFormat(paragraph.Numbering) == ir.NumFormatBullet {
			e.output.WriteString("- ")
		} else {
			e.output.WriteString("+ ")
		}
	}
	for _, run := range paragraph.Runs {
		e.writeRun(run)
	}
	for _, note := range paragraph.Footnotes {
		e.output.WriteString("#footnote[" + blocksText(note.Blocks) + "]")
	}
	e.output.WriteString("\n\n")
}

func (e *exporter) writeRun(run ir.Run) {
	value := escape(run.Text)
	if run.Attrs.Bold {
		value = "*" + value + "*"
	}
	if run.Attrs.Italic {
		value = "_" + value + "_"
	}
	var textOptions []string
	if run.Attrs.Color != (ir.Color{}) {
		textOptions = append(textOptions, fmt.Sprintf("fill: rgb(%d, %d, %d)", run.Attrs.Color.R, run.Attrs.Color.G, run.Attrs.Color.B))
	}
	if run.Attrs.FontSize > 0 {
		textOptions = append(textOptions, "size: "+points(run.Attrs.FontSize))
	}
	if run.Attrs.FontName != "" {
		textOptions = append(textOptions, `font: "`+stringLiteral(run.Attrs.FontName)+`"`)
	}
	if len(textOptions) > 0 {
		value = "#text(" + strings.Join(textOptions, ", ") + ")[" + value + "]"
	}
	if run.Hyperlink != nil && run.Hyperlink.URL != "" {
		value = `#link("` + stringLiteral(run.Hyperlink.URL) + `")[` + value + `]`
	}
	e.output.WriteString(value)
	if run.Break == ir.BreakLine {
		e.output.WriteString(" \\\n")
	}
}

func (e *exporter) writeTable(table *ir.Table) {
	columns := 1
	for _, row := range table.Rows {
		if len(row.Cells) > columns {
			columns = len(row.Cells)
		}
	}
	e.output.WriteString("#table(columns: " + strconv.Itoa(columns))
	for _, row := range table.Rows {
		for _, cell := range row.Cells {
			e.output.WriteString(", [" + blocksText(cell.Blocks) + "]")
		}
	}
	e.output.WriteString(")\n\n")
}

func (e *exporter) writeImage(image *ir.Image) {
	var options []string
	if image.Width > 0 {
		options = append(options, "width: "+points(image.Width))
	}
	if image.Height > 0 {
		options = append(options, "height: "+points(image.Height))
	}
	arguments := `"` + stringLiteral(image.ID+imageExtension(image.Format)) + `"`
	if len(options) > 0 {
		arguments += ", " + strings.Join(options, ", ")
	}
	e.output.WriteString("#image(" + arguments + ")\n\n")
}

func (e *exporter) numberingFormat(reference *ir.NumberingRef) ir.NumberFormat {
	for _, definition := range e.document.Styles.Numbering {
		if definition.ID != reference.ID {
			continue
		}
		for _, level := range definition.Levels {
			if level.Level == reference.Level {
				return level.Format
			}
		}
	}
	return ir.NumFormatBullet
}

func blocksText(blocks []ir.Block) string {
	var result strings.Builder
	for _, block := range blocks {
		switch value := block.(type) {
		case *ir.Paragraph:
			for _, run := range value.Runs {
				result.WriteString(escape(run.Text))
			}
		case *ir.Image:
			result.WriteString("#image(\"" + stringLiteral(value.ID+imageExtension(value.Format)) + "\")")
		}
	}
	return result.String()
}

func headingLevel(style string) int {
	if !strings.HasPrefix(style, "Heading") {
		return 0
	}
	level, err := strconv.Atoi(strings.TrimPrefix(style, "Heading"))
	if err != nil || level < 1 || level > 6 {
		return 0
	}
	return level
}

func points(value float64) string {
	return strconv.FormatFloat(value, 'f', -1, 64) + "pt"
}

func escape(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "#", "\\#", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]", "$", "\\$")
	return replacer.Replace(value)
}

func stringLiteral(value string) string {
	return strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), `"`, `\"`)
}

func imageExtension(format ir.ImageFormat) string {
	switch format {
	case ir.ImageJPEG:
		return ".jpg"
	case ir.ImageGIF:
		return ".gif"
	case ir.ImageSVG:
		return ".svg"
	default:
		return ".png"
	}
}
