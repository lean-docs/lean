package html

import (
	"encoding/base64"
	"fmt"
	stdhtml "html"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

func Export(doc *ir.Document) ([]byte, error) {
	if doc == nil {
		return nil, fmt.Errorf("html exporter: document is nil")
	}
	var output strings.Builder
	language := doc.Meta.Language
	if language == "" {
		language = "en"
	}
	output.WriteString(`<!DOCTYPE html><html lang="` + attribute(language) + `"><head><meta charset="utf-8"><title>` + text(doc.Meta.Title) + `</title></head><body>`)
	state := exporter{document: doc, output: &output}
	for _, section := range doc.Sections {
		state.writeBlocks(section.Blocks)
	}
	output.WriteString("</body></html>")
	return []byte(output.String()), nil
}

type exporter struct {
	document *ir.Document
	output   *strings.Builder
}

func (e *exporter) writeBlocks(blocks []ir.Block) {
	for index := 0; index < len(blocks); {
		paragraph, ok := blocks[index].(*ir.Paragraph)
		if ok && paragraph.Numbering != nil {
			index = e.writeList(blocks, index, paragraph.Numbering)
			continue
		}
		e.writeBlock(blocks[index])
		index++
	}
}

func (e *exporter) writeBlock(block ir.Block) {
	switch value := block.(type) {
	case *ir.Paragraph:
		tag := "p"
		if level := headingLevel(value.Style); level > 0 {
			tag = "h" + strconv.Itoa(level)
		}
		e.output.WriteString("<" + tag + ">")
		e.writeRuns(value.Runs)
		e.output.WriteString("</" + tag + ">")
	case *ir.Table:
		e.writeTable(value)
	case *ir.Image:
		e.writeImage(value)
	}
}

func (e *exporter) writeRuns(runs []ir.Run) {
	for _, run := range runs {
		value := text(run.Text)
		if run.Attrs.Underline != ir.UnderlineNone {
			value = "<u>" + value + "</u>"
		}
		if run.Attrs.Strike {
			value = "<s>" + value + "</s>"
		}
		if run.Attrs.Italic {
			value = "<em>" + value + "</em>"
		}
		if run.Attrs.Bold {
			value = "<strong>" + value + "</strong>"
		}
		style := runStyle(run.Attrs)
		if style != "" {
			value = `<span style="` + attribute(style) + `">` + value + `</span>`
		}
		if run.Hyperlink != nil {
			if run.Hyperlink.URL != "" {
				value = `<a href="` + attribute(run.Hyperlink.URL) + `">` + value + `</a>`
			} else if run.Hyperlink.Bookmark != "" {
				value = `<a href="#` + attribute(run.Hyperlink.Bookmark) + `">` + value + `</a>`
			}
		}
		e.output.WriteString(value)
		if run.Break == ir.BreakLine {
			e.output.WriteString("<br>")
		}
	}
}

func (e *exporter) writeList(blocks []ir.Block, start int, reference *ir.NumberingRef) int {
	tag := "ul"
	if e.numberingFormat(reference) != ir.NumFormatBullet {
		tag = "ol"
	}
	e.output.WriteString("<" + tag + ">")
	index := start
	for index < len(blocks) {
		paragraph, ok := blocks[index].(*ir.Paragraph)
		if !ok || paragraph.Numbering == nil || paragraph.Numbering.ID != reference.ID || paragraph.Numbering.Level != reference.Level {
			break
		}
		e.output.WriteString("<li>")
		e.writeRuns(paragraph.Runs)
		e.output.WriteString("</li>")
		index++
	}
	e.output.WriteString("</" + tag + ">")
	return index
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

func (e *exporter) writeTable(table *ir.Table) {
	e.output.WriteString("<table>")
	for _, header := range []bool{true, false} {
		container := "tbody"
		if header {
			container = "thead"
		}
		var rows []ir.TableRow
		for _, row := range table.Rows {
			if row.IsHeader == header {
				rows = append(rows, row)
			}
		}
		if len(rows) == 0 {
			continue
		}
		e.output.WriteString("<" + container + ">")
		for _, row := range rows {
			e.output.WriteString("<tr>")
			for _, cell := range row.Cells {
				tag := "td"
				if header {
					tag = "th"
				}
				e.output.WriteString("<" + tag)
				if cell.ColSpan > 1 {
					e.output.WriteString(` colspan="` + strconv.Itoa(cell.ColSpan) + `"`)
				}
				if cell.RowSpan > 1 {
					e.output.WriteString(` rowspan="` + strconv.Itoa(cell.RowSpan) + `"`)
				}
				e.output.WriteString(">")
				e.writeBlocks(cell.Blocks)
				e.output.WriteString("</" + tag + ">")
			}
			e.output.WriteString("</tr>")
		}
		e.output.WriteString("</" + container + ">")
	}
	e.output.WriteString("</table>")
}

func (e *exporter) writeImage(image *ir.Image) {
	source := "data:" + imageMediaType(image.Format) + ";base64," + base64.StdEncoding.EncodeToString(image.Data)
	e.output.WriteString(`<img src="` + source + `" alt="` + attribute(image.Alt) + `"`)
	if image.Width > 0 {
		e.output.WriteString(` width="` + strconv.FormatFloat(image.Width, 'f', -1, 64) + `"`)
	}
	if image.Height > 0 {
		e.output.WriteString(` height="` + strconv.FormatFloat(image.Height, 'f', -1, 64) + `"`)
	}
	e.output.WriteString(">")
}

func runStyle(attrs ir.RunAttrs) string {
	var declarations []string
	if attrs.FontName != "" {
		declarations = append(declarations, "font-family:"+attrs.FontName)
	}
	if attrs.FontSize > 0 {
		declarations = append(declarations, "font-size:"+strconv.FormatFloat(attrs.FontSize, 'f', -1, 64)+"pt")
	}
	if attrs.Color != (ir.Color{}) {
		declarations = append(declarations, fmt.Sprintf("color:#%02X%02X%02X", attrs.Color.R, attrs.Color.G, attrs.Color.B))
	}
	return strings.Join(declarations, ";")
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

func imageMediaType(format ir.ImageFormat) string {
	switch format {
	case ir.ImageJPEG:
		return "image/jpeg"
	case ir.ImageGIF:
		return "image/gif"
	case ir.ImageSVG:
		return "image/svg+xml"
	default:
		return "image/png"
	}
}

func text(value string) string {
	return stdhtml.EscapeString(value)
}

func attribute(value string) string {
	return stdhtml.EscapeString(value)
}
