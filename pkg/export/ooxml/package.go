package ooxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

var ErrNilDocument = errors.New("ooxml exporter: document is nil")

type packageState struct {
	doc          *ir.Document
	images       []*ir.Image
	hasHeader    bool
	hasFooter    bool
	hasFootnotes bool
}

func buildPackage(doc *ir.Document) ([]byte, error) {
	if doc == nil {
		return nil, ErrNilDocument
	}
	state := inspect(doc)
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	parts := map[string]string{
		"[Content_Types].xml":          state.contentTypes(),
		"_rels/.rels":                  rootRelationships,
		"word/document.xml":            state.documentXML(),
		"word/_rels/document.xml.rels": state.documentRelationships(),
	}
	if len(doc.Styles.Numbering) > 0 {
		parts["word/numbering.xml"] = numberingXML(doc.Styles.Numbering)
	}
	if state.hasHeader {
		parts["word/header1.xml"] = headerFooterXML("hdr", firstHeader(doc))
	}
	if state.hasFooter {
		parts["word/footer1.xml"] = headerFooterXML("ftr", firstFooter(doc))
	}
	if state.hasFootnotes {
		parts["word/footnotes.xml"] = footnotesXML(doc)
	}
	for name, content := range parts {
		if err := writePart(writer, name, []byte(content)); err != nil {
			return nil, err
		}
	}
	for index, image := range state.images {
		name := fmt.Sprintf("word/media/image%d.%s", index+1, imageExtension(image.Format))
		if err := writePart(writer, name, image.Data); err != nil {
			return nil, err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("ooxml exporter: close package: %w", err)
	}
	return output.Bytes(), nil
}

func inspect(doc *ir.Document) *packageState {
	state := &packageState{doc: doc}
	for _, section := range doc.Sections {
		state.hasHeader = state.hasHeader || section.Header != nil
		state.hasFooter = state.hasFooter || section.Footer != nil
		walkBlocks(section.Blocks, func(block ir.Block) {
			switch value := block.(type) {
			case *ir.Image:
				state.images = append(state.images, value)
			case *ir.Paragraph:
				state.hasFootnotes = state.hasFootnotes || len(value.Footnotes) > 0
			}
		})
	}
	return state
}

func walkBlocks(blocks []ir.Block, visit func(ir.Block)) {
	for _, block := range blocks {
		visit(block)
		if table, ok := block.(*ir.Table); ok {
			for _, row := range table.Rows {
				for _, cell := range row.Cells {
					walkBlocks(cell.Blocks, visit)
				}
			}
		}
	}
}

func (state *packageState) documentXML() string {
	var body strings.Builder
	imageIndex := 0
	for _, section := range state.doc.Sections {
		for _, block := range section.Blocks {
			writeBlock(&body, block, &imageIndex)
		}
		writeSectionProperties(&body, section, state.hasHeader, state.hasFooter)
	}
	if len(state.doc.Sections) == 0 {
		body.WriteString("<w:sectPr/>")
	}
	return xmlDeclaration + `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture"><w:body>` + body.String() + `</w:body></w:document>`
}

func writeBlock(output *strings.Builder, block ir.Block, imageIndex *int) {
	switch value := block.(type) {
	case *ir.Paragraph:
		writeParagraph(output, value)
	case *ir.Table:
		writeTable(output, value, imageIndex)
	case *ir.Image:
		*imageIndex = *imageIndex + 1
		writeImage(output, value, *imageIndex)
	case *ir.Bookmark:
		output.WriteString(`<w:p><w:bookmarkStart w:id="` + escape(value.ID) + `" w:name="` + escape(value.Name) + `"/><w:bookmarkEnd w:id="` + escape(value.ID) + `"/></w:p>`)
	}
}

func writeParagraph(output *strings.Builder, paragraph *ir.Paragraph) {
	output.WriteString("<w:p>")
	writeParagraphProperties(output, paragraph)
	for _, run := range paragraph.Runs {
		writeRun(output, run)
	}
	for index := range paragraph.Footnotes {
		output.WriteString(`<w:r><w:footnoteReference w:id="` + strconv.Itoa(index+1) + `"/></w:r>`)
	}
	output.WriteString("</w:p>")
}

func writeParagraphProperties(output *strings.Builder, paragraph *ir.Paragraph) {
	attrs := paragraph.Para
	if paragraph.Style == "" && paragraph.Numbering == nil && !hasParagraphProperties(attrs) {
		return
	}
	output.WriteString("<w:pPr>")
	if paragraph.Style != "" {
		output.WriteString(`<w:pStyle w:val="` + escape(paragraph.Style) + `"/>`)
	}
	if paragraph.Numbering != nil {
		output.WriteString(`<w:numPr><w:ilvl w:val="` + strconv.Itoa(paragraph.Numbering.Level) + `"/><w:numId w:val="` + escape(paragraph.Numbering.ID) + `"/></w:numPr>`)
	}
	if attrs.Align != ir.AlignLeft {
		output.WriteString(`<w:jc w:val="` + alignment(attrs.Align) + `"/>`)
	}
	if attrs.Spacing != (ir.Spacing{}) {
		output.WriteString(`<w:spacing w:before="` + twips(attrs.Spacing.Before) + `" w:after="` + twips(attrs.Spacing.After) + `" w:line="` + twips(attrs.Spacing.Line) + `"/>`)
	}
	if attrs.Indent != (ir.Indent{}) {
		output.WriteString(`<w:ind w:left="` + twips(attrs.Indent.Left) + `" w:right="` + twips(attrs.Indent.Right) + `" w:firstLine="` + twips(attrs.Indent.FirstLine) + `" w:hanging="` + twips(attrs.Indent.Hanging) + `"/>`)
	}
	if len(attrs.TabStops) > 0 {
		output.WriteString("<w:tabs>")
		for _, tab := range attrs.TabStops {
			output.WriteString(`<w:tab w:val="` + tabAlignment(tab.Alignment) + `" w:pos="` + twips(tab.Position) + `" w:leader="` + tabLeader(tab.Leader) + `"/>`)
		}
		output.WriteString("</w:tabs>")
	}
	if attrs.KeepTogether {
		output.WriteString("<w:keepLines/>")
	}
	if attrs.KeepWithNext {
		output.WriteString("<w:keepNext/>")
	}
	if attrs.PageBreakBefore {
		output.WriteString("<w:pageBreakBefore/>")
	}
	output.WriteString("</w:pPr>")
}

func hasParagraphProperties(attrs ir.ParaAttrs) bool {
	return attrs.Align != ir.AlignLeft || attrs.Spacing != (ir.Spacing{}) || attrs.Indent != (ir.Indent{}) || len(attrs.TabStops) > 0 || attrs.KeepTogether || attrs.KeepWithNext || attrs.PageBreakBefore
}

func writeRun(output *strings.Builder, run ir.Run) {
	output.WriteString("<w:r>")
	writeRunProperties(output, run.Attrs)
	if run.Text != "" {
		output.WriteString(`<w:t xml:space="preserve">` + escape(run.Text) + `</w:t>`)
	}
	if run.Break != ir.BreakNone {
		output.WriteString(`<w:br w:type="` + breakType(run.Break) + `"/>`)
	}
	output.WriteString("</w:r>")
}

func writeRunProperties(output *strings.Builder, attrs ir.RunAttrs) {
	if attrs == (ir.RunAttrs{}) {
		return
	}
	output.WriteString("<w:rPr>")
	if attrs.Bold {
		output.WriteString("<w:b/>")
	}
	if attrs.Italic {
		output.WriteString("<w:i/>")
	}
	if attrs.Underline != ir.UnderlineNone {
		output.WriteString(`<w:u w:val="` + underline(attrs.Underline) + `"/>`)
	}
	if attrs.Strike {
		output.WriteString("<w:strike/>")
	}
	if attrs.SmallCaps {
		output.WriteString("<w:smallCaps/>")
	}
	if attrs.AllCaps {
		output.WriteString("<w:caps/>")
	}
	if attrs.FontSize > 0 {
		output.WriteString(`<w:sz w:val="` + strconv.FormatFloat(attrs.FontSize*2, 'f', -1, 64) + `"/>`)
	}
	if attrs.FontName != "" {
		output.WriteString(`<w:rFonts w:ascii="` + escape(attrs.FontName) + `" w:hAnsi="` + escape(attrs.FontName) + `"/>`)
	}
	if attrs.Color != (ir.Color{}) {
		output.WriteString(`<w:color w:val="` + color(attrs.Color) + `"/>`)
	}
	if attrs.Baseline != ir.BaselineNone {
		output.WriteString(`<w:vertAlign w:val="` + baseline(attrs.Baseline) + `"/>`)
	}
	output.WriteString("</w:rPr>")
}

func writeTable(output *strings.Builder, table *ir.Table, imageIndex *int) {
	output.WriteString("<w:tbl>")
	if len(table.ColumnWidths) > 0 {
		output.WriteString("<w:tblGrid>")
		for _, width := range table.ColumnWidths {
			output.WriteString(`<w:gridCol w:w="` + twips(width) + `"/>`)
		}
		output.WriteString("</w:tblGrid>")
	}
	for _, row := range table.Rows {
		output.WriteString("<w:tr>")
		for _, cell := range row.Cells {
			output.WriteString("<w:tc>")
			if cell.ColSpan > 1 || cell.Width > 0 {
				output.WriteString("<w:tcPr>")
				if cell.Width > 0 {
					output.WriteString(`<w:tcW w:w="` + twips(cell.Width) + `" w:type="dxa"/>`)
				}
				if cell.ColSpan > 1 {
					output.WriteString(`<w:gridSpan w:val="` + strconv.Itoa(cell.ColSpan) + `"/>`)
				}
				output.WriteString("</w:tcPr>")
			}
			for _, block := range cell.Blocks {
				writeBlock(output, block, imageIndex)
			}
			if len(cell.Blocks) == 0 {
				output.WriteString("<w:p/>")
			}
			output.WriteString("</w:tc>")
		}
		output.WriteString("</w:tr>")
	}
	output.WriteString("</w:tbl>")
}

func writeImage(output *strings.Builder, image *ir.Image, index int) {
	cx := int64(image.Width * 12700)
	cy := int64(image.Height * 12700)
	if cx == 0 {
		cx = 914400
	}
	if cy == 0 {
		cy = 914400
	}
	output.WriteString(fmt.Sprintf(`<w:p><w:r><w:drawing><wp:inline><wp:extent cx="%d" cy="%d"/><wp:docPr id="%d" name="Image %d" descr="%s"/><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic><pic:blipFill><a:blip r:embed="rIdImage%d"/></pic:blipFill><pic:spPr><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic></wp:inline></w:drawing></w:r></w:p>`, cx, cy, index, index, escape(image.Alt), index))
}

func writeSectionProperties(output *strings.Builder, section ir.Section, header, footer bool) {
	output.WriteString("<w:sectPr>")
	if header {
		output.WriteString(`<w:headerReference w:type="default" r:id="rIdHeader"/>`)
	}
	if footer {
		output.WriteString(`<w:footerReference w:type="default" r:id="rIdFooter"/>`)
	}
	properties := section.Properties
	if properties.Width > 0 || properties.Height > 0 {
		output.WriteString(`<w:pgSz w:w="` + twips(properties.Width) + `" w:h="` + twips(properties.Height) + `"/>`)
	}
	output.WriteString(`<w:pgMar w:top="` + twips(properties.MarginTop) + `" w:right="` + twips(properties.MarginRight) + `" w:bottom="` + twips(properties.MarginBottom) + `" w:left="` + twips(properties.MarginLeft) + `"/>`)
	output.WriteString("</w:sectPr>")
}

func (state *packageState) contentTypes() string {
	content := `<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>`
	if len(state.doc.Styles.Numbering) > 0 {
		content += `<Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>`
	}
	if state.hasHeader {
		content += `<Override PartName="/word/header1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.header+xml"/>`
	}
	if state.hasFooter {
		content += `<Override PartName="/word/footer1.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footer+xml"/>`
	}
	if state.hasFootnotes {
		content += `<Override PartName="/word/footnotes.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.footnotes+xml"/>`
	}
	for _, image := range state.images {
		extension := imageExtension(image.Format)
		content += `<Default Extension="` + extension + `" ContentType="` + imageContentType(extension) + `"/>`
	}
	return xmlDeclaration + `<Types xmlns="http://schemas.openxmlformats.org/package/2006/content-types">` + content + `</Types>`
}

func (state *packageState) documentRelationships() string {
	content := ""
	if len(state.doc.Styles.Numbering) > 0 {
		content += `<Relationship Id="rIdNumbering" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/>`
	}
	if state.hasHeader {
		content += `<Relationship Id="rIdHeader" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/header" Target="header1.xml"/>`
	}
	if state.hasFooter {
		content += `<Relationship Id="rIdFooter" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footer" Target="footer1.xml"/>`
	}
	if state.hasFootnotes {
		content += `<Relationship Id="rIdFootnotes" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footnotes" Target="footnotes.xml"/>`
	}
	for index := range state.images {
		content += `<Relationship Id="rIdImage` + strconv.Itoa(index+1) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/image` + strconv.Itoa(index+1) + `.` + imageExtension(state.images[index].Format) + `"/>`
	}
	return xmlDeclaration + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` + content + `</Relationships>`
}

func numberingXML(definitions []ir.NumberingDef) string {
	var content strings.Builder
	for index, definition := range definitions {
		id := index + 1
		content.WriteString(`<w:abstractNum w:abstractNumId="` + strconv.Itoa(id) + `">`)
		for _, level := range definition.Levels {
			content.WriteString(`<w:lvl w:ilvl="` + strconv.Itoa(level.Level) + `"><w:start w:val="1"/><w:numFmt w:val="` + numberFormat(level.Format) + `"/><w:lvlText w:val="` + escape(level.Text) + `"/></w:lvl>`)
		}
		content.WriteString(`</w:abstractNum><w:num w:numId="` + escape(definition.ID) + `"><w:abstractNumId w:val="` + strconv.Itoa(id) + `"/></w:num>`)
	}
	return xmlDeclaration + `<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` + content.String() + `</w:numbering>`
}

func headerFooterXML(element string, value *ir.HeaderFooter) string {
	var content strings.Builder
	if value != nil {
		index := 0
		for _, block := range value.Blocks {
			writeBlock(&content, block, &index)
		}
	}
	return xmlDeclaration + `<w:` + element + ` xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` + content.String() + `</w:` + element + `>`
}

func footnotesXML(doc *ir.Document) string {
	var content strings.Builder
	index := 0
	for _, section := range doc.Sections {
		walkBlocks(section.Blocks, func(block ir.Block) {
			paragraph, ok := block.(*ir.Paragraph)
			if !ok {
				return
			}
			for _, note := range paragraph.Footnotes {
				index++
				content.WriteString(`<w:footnote w:id="` + strconv.Itoa(index) + `">`)
				image := 0
				for _, nested := range note.Blocks {
					writeBlock(&content, nested, &image)
				}
				content.WriteString(`</w:footnote>`)
			}
		})
	}
	return xmlDeclaration + `<w:footnotes xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` + content.String() + `</w:footnotes>`
}

func firstHeader(doc *ir.Document) *ir.HeaderFooter {
	for _, section := range doc.Sections {
		if section.Header != nil {
			return section.Header
		}
	}
	return nil
}
func firstFooter(doc *ir.Document) *ir.HeaderFooter {
	for _, section := range doc.Sections {
		if section.Footer != nil {
			return section.Footer
		}
	}
	return nil
}
func writePart(writer *zip.Writer, name string, data []byte) error {
	part, err := writer.Create(name)
	if err != nil {
		return err
	}
	_, err = part.Write(data)
	return err
}
func escape(value string) string {
	var output bytes.Buffer
	_ = xml.EscapeText(&output, []byte(value))
	return output.String()
}
func twips(points float64) string { return strconv.FormatInt(int64(points*20), 10) }
func imageExtension(value ir.ImageFormat) string {
	if value == ir.ImageJPEG {
		return "jpg"
	}
	if value == ir.ImageGIF {
		return "gif"
	}
	if value == ir.ImageSVG {
		return "svg"
	}
	return "png"
}
func imageContentType(extension string) string {
	if extension == "jpg" {
		return "image/jpeg"
	}
	if extension == "gif" {
		return "image/gif"
	}
	if extension == "svg" {
		return "image/svg+xml"
	}
	return "image/png"
}
func alignment(value ir.Alignment) string {
	if value == ir.AlignCenter {
		return "center"
	}
	if value == ir.AlignRight {
		return "right"
	}
	if value == ir.AlignJustify {
		return "both"
	}
	return "left"
}
func tabAlignment(value ir.TabAlignment) string {
	if value == ir.TabAlignRight {
		return "right"
	}
	if value == ir.TabAlignCenter {
		return "center"
	}
	if value == ir.TabAlignDecimal {
		return "decimal"
	}
	return "left"
}
func tabLeader(value ir.TabLeader) string {
	if value == ir.TabLeaderDot {
		return "dot"
	}
	if value == ir.TabLeaderDash {
		return "hyphen"
	}
	if value == ir.TabLeaderUnderscore {
		return "underscore"
	}
	return "none"
}
func underline(value ir.UnderlineStyle) string {
	if value == ir.UnderlineDouble {
		return "double"
	}
	if value == ir.UnderlineWavy {
		return "wave"
	}
	if value == ir.UnderlineDash {
		return "dash"
	}
	return "single"
}
func baseline(value ir.BaselineShift) string {
	if value == ir.BaselineSubscript {
		return "subscript"
	}
	return "superscript"
}
func breakType(value ir.BreakType) string {
	if value == ir.BreakPage {
		return "page"
	}
	if value == ir.BreakColumn {
		return "column"
	}
	return "textWrapping"
}
func color(value ir.Color) string { return fmt.Sprintf("%02X%02X%02X", value.R, value.G, value.B) }
func numberFormat(value ir.NumberFormat) string {
	if value == ir.NumFormatDecimal {
		return "decimal"
	}
	return "bullet"
}

const xmlDeclaration = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
const rootRelationships = xmlDeclaration + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`
