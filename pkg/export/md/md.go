package md

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

func Export(doc *ir.Document) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("md exporter: document is nil")
	}
	state := exporter{document: doc, counters: make(map[string]int)}
	for _, section := range doc.Sections {
		for _, block := range section.Blocks {
			state.writeBlock(block)
		}
	}
	return []byte(strings.TrimSpace(state.output.String())), nil
}

type exporter struct {
	document *ir.Document
	output   strings.Builder
	counters map[string]int
}

func (e *exporter) writeBlock(block ir.Block) {
	if e.output.Len() > 0 {
		e.output.WriteString("\n\n")
	}
	switch value := block.(type) {
	case *ir.Paragraph:
		e.writeParagraph(value)
	case *ir.Table:
		e.writeTable(value)
	case *ir.Image:
		e.output.WriteString("![" + escapeText(value.Alt) + "](" + imageTarget(value) + ")")
	}
}

func (e *exporter) writeParagraph(paragraph *ir.Paragraph) {
	if paragraph.Style == "Code" {
		e.output.WriteString("```\n" + paragraphText(paragraph) + "\n```")
		return
	}
	if level := headingLevel(paragraph.Style); level > 0 {
		e.output.WriteString(strings.Repeat("#", level) + " ")
	}
	if paragraph.Numbering != nil {
		e.output.WriteString(strings.Repeat("  ", paragraph.Numbering.Level))
		if e.numberingFormat(paragraph.Numbering) == ir.NumFormatBullet {
			e.output.WriteString("- ")
		} else {
			key := paragraph.Numbering.ID + ":" + strconv.Itoa(paragraph.Numbering.Level)
			e.counters[key]++
			start := e.numberingStart(paragraph.Numbering)
			e.output.WriteString(strconv.Itoa(start+e.counters[key]-1) + ". ")
		}
	}
	for _, run := range paragraph.Runs {
		e.writeRun(run)
	}
}

func (e *exporter) writeRun(run ir.Run) {
	text := escapeText(run.Text)
	if run.Attrs.FontName == "monospace" && text != "" {
		text = "`" + strings.ReplaceAll(text, "`", "\\`") + "`"
	}
	if run.Attrs.Strike {
		text = "~~" + text + "~~"
	}
	if run.Attrs.Bold && run.Attrs.Italic {
		text = "***" + text + "***"
	} else if run.Attrs.Bold {
		text = "**" + text + "**"
	} else if run.Attrs.Italic {
		text = "*" + text + "*"
	}
	if run.Hyperlink != nil && run.Hyperlink.URL != "" {
		text = "[" + text + "](" + strings.ReplaceAll(run.Hyperlink.URL, ")", "\\)") + ")"
	}
	e.output.WriteString(text)
	if run.Break == ir.BreakLine {
		e.output.WriteString("  \n")
	}
}

func (e *exporter) writeTable(table *ir.Table) {
	if len(table.Rows) == 0 {
		return
	}
	for index, row := range table.Rows {
		e.output.WriteString("|")
		for _, cell := range row.Cells {
			e.output.WriteString(" " + escapeTableCell(cellText(cell)) + " |")
		}
		e.output.WriteString("\n")
		if index == 0 {
			e.output.WriteString("|")
			for range row.Cells {
				e.output.WriteString(" --- |")
			}
			e.output.WriteString("\n")
		}
	}
	e.output.WriteString("")
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

func (e *exporter) numberingStart(reference *ir.NumberingRef) int {
	for _, definition := range e.document.Styles.Numbering {
		if definition.ID != reference.ID {
			continue
		}
		for _, level := range definition.Levels {
			if level.Level == reference.Level && level.Start > 0 {
				return level.Start
			}
		}
	}
	return 1
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

func paragraphText(paragraph *ir.Paragraph) string {
	var text strings.Builder
	for _, run := range paragraph.Runs {
		text.WriteString(run.Text)
		if run.Break == ir.BreakLine {
			text.WriteByte('\n')
		}
	}
	return text.String()
}

func cellText(cell ir.TableCell) string {
	var text strings.Builder
	for _, block := range cell.Blocks {
		if paragraph, ok := block.(*ir.Paragraph); ok {
			if text.Len() > 0 {
				text.WriteString("<br>")
			}
			text.WriteString(paragraphText(paragraph))
		}
	}
	return text.String()
}

func imageTarget(image *ir.Image) string {
	extension := "png"
	switch image.Format {
	case ir.ImageJPEG:
		extension = "jpg"
	case ir.ImageGIF:
		extension = "gif"
	case ir.ImageSVG:
		extension = "svg"
	}
	return image.ID + "." + extension
}

func escapeText(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "*", "\\*", "_", "\\_", "[", "\\[", "]", "\\]")
	return replacer.Replace(value)
}

func escapeTableCell(value string) string {
	return strings.ReplaceAll(value, "|", "\\|")
}
