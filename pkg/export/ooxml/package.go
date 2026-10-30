package ooxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

var ErrNilDocument = errors.New("ooxml exporter: document is nil")

type packageState struct {
	anchoredShapes    map[*ir.Shape]bool
	drawingIDs        map[ir.Block]int
	anchoredImages    map[*ir.Image]bool
	bookmarkEvents    map[*ir.Paragraph][]bookmarkEvent
	anchoredBookmarks map[*ir.Bookmark]bool
	doc               *ir.Document
	images            []*ir.Image
	hyperlinks        []string
	hyperlinkIDs      map[string]int
	hasFootnotes      bool
	auxiliary         []auxiliaryPart
	footnoteIDs       map[*ir.Paragraph][]int
	notes             []ir.Footnote
	imageIDs          map[*ir.Image]int
	numbering         []ir.NumberingDef
	numberingIDs      map[*ir.Paragraph]string
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
	if len(state.numbering) > 0 {
		parts["word/numbering.xml"] = numberingXML(state.numbering)
	}
	if len(doc.Styles.Named) > 0 || doc.Styles.Defaults != (ir.RunAttrs{}) {
		parts["word/styles.xml"] = stylesXML(doc.Styles)
	}
	for _, part := range state.auxiliary {
		parts["word/"+part.filename] = headerFooterXML(part.element, part.value, state)
		parts["word/_rels/"+part.filename+".rels"] = state.contentRelationships()
	}
	if state.hasFootnotes {
		parts["word/footnotes.xml"] = state.footnotesXML()
		parts["word/_rels/footnotes.xml.rels"] = state.contentRelationships()
	}
	if state.hasEvenVariant() {
		parts["word/settings.xml"] = xmlDeclaration + `<w:settings xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main"><w:evenAndOddHeaders/></w:settings>`
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
	state := &packageState{drawingIDs: make(map[ir.Block]int), doc: doc, hyperlinkIDs: make(map[string]int), footnoteIDs: make(map[*ir.Paragraph][]int), imageIDs: make(map[*ir.Image]int)}
	for sectionIndex, section := range doc.Sections {
		state.collectBlocks(section.Blocks)
		state.addAuxiliary(sectionIndex, "hdr", section.Header, section.Headers)
		state.addAuxiliary(sectionIndex, "ftr", section.Footer, section.Footers)
	}
	for index := 0; index < len(state.notes); index++ {
		state.collectBlocks(state.notes[index].Blocks)
	}
	state.hasFootnotes = len(state.notes) > 0
	state.prepareNumbering()
	state.prepareBookmarks()

	return state
}

func walkBlocks(blocks []ir.Block, visit func(ir.Block)) {
	for _, block := range blocks {
		visit(block)
		if shape, ok := block.(*ir.Shape); ok {
			walkBlocks(shape.Blocks, visit)
		}
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
	for sectionIndex, section := range state.doc.Sections {
		var properties strings.Builder
		writeSectionProperties(&properties, section, state.auxiliaryReferences(sectionIndex))
		for index, block := range section.Blocks {
			if sectionIndex < len(state.doc.Sections)-1 && index == len(section.Blocks)-1 {
				if paragraph, ok := block.(*ir.Paragraph); ok {
					writeParagraph(&body, paragraph, state, properties.String())
					continue
				}
			}
			writeBlock(&body, block, &imageIndex, state)
		}
		if sectionIndex == len(state.doc.Sections)-1 {
			body.WriteString(properties.String())
		} else if len(section.Blocks) == 0 || section.Blocks[len(section.Blocks)-1].BlockType() != ir.BlockParagraph {
			body.WriteString("<w:p><w:pPr>" + properties.String() + "</w:pPr></w:p>")
		}
	}

	if len(state.doc.Sections) == 0 {
		body.WriteString("<w:sectPr/>")
	}
	return xmlDeclaration + `<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture" xmlns:wps="http://schemas.microsoft.com/office/word/2010/wordprocessingShape"><w:body>` + body.String() + `</w:body></w:document>`
}

func writeBlock(output *strings.Builder, block ir.Block, imageIndex *int, state *packageState) {
	switch value := block.(type) {
	case *ir.Paragraph:
		writeParagraph(output, value, state)
	case *ir.Table:
		writeTable(output, value, imageIndex, state)
	case *ir.Image:
		if state != nil && state.anchoredImages[value] {
			return
		}
		*imageIndex = *imageIndex + 1
		index := *imageIndex
		if state != nil {
			index = state.imageIDs[value]
		}
		writeImage(output, value, index, state)
	case *ir.Shape:
		if state != nil && state.anchoredShapes[value] {
			return
		}
		output.WriteString("<w:p>")
		writeShapeRun(output, value, state)
		output.WriteString("</w:p>")
	case *ir.Bookmark:
		if state != nil && state.anchoredBookmarks[value] {
			return
		}
		output.WriteString(`<w:p><w:bookmarkStart w:id="` + escape(value.ID) + `" w:name="` + escape(value.Name) + `"/><w:bookmarkEnd w:id="` + escape(value.ID) + `"/></w:p>`)
	}
}

func writeParagraph(output *strings.Builder, paragraph *ir.Paragraph, state *packageState, section ...string) {
	output.WriteString("<w:p>")
	var properties strings.Builder
	formatted := *paragraph
	if state != nil && paragraph.Numbering != nil {
		reference := *paragraph.Numbering
		reference.ID = state.numberingIDs[paragraph]
		formatted.Numbering = &reference
	}
	writeParagraphProperties(&properties, &formatted)
	content := properties.String()
	if len(section) > 0 {
		if content == "" {
			content = "<w:pPr>" + section[0] + "</w:pPr>"
		} else {
			content = strings.Replace(content, "</w:pPr>", section[0]+"</w:pPr>", 1)
		}
	}
	output.WriteString(content)
	if state != nil && len(state.bookmarkEvents[paragraph]) > 0 {
		state.writeBookmarkedRuns(output, paragraph)
	} else {
		for _, run := range paragraph.Runs {
			writeRun(output, run, state)
		}
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
	if attrs.AlignmentSpecified() {
		output.WriteString(`<w:jc w:val="` + alignment(attrs.Align) + `"/>`)
	}
	if attrs.SpacingSpecified() {
		output.WriteString(`<w:spacing w:before="` + twips(attrs.Spacing.Before) + `" w:after="` + twips(attrs.Spacing.After) + `" w:line="` + twips(attrs.Spacing.Line) + `" w:lineRule="` + lineRule(attrs.Spacing.LineRule) + `"/>`)
	}
	if attrs.IndentSpecified() {
		output.WriteString(`<w:ind w:left="` + twips(attrs.Indent.Left) + `" w:right="` + twips(attrs.Indent.Right) + `" w:firstLine="` + twips(attrs.Indent.FirstLine) + `" w:hanging="` + twips(attrs.Indent.Hanging) + `"/>`)
	}
	if len(attrs.TabStops) > 0 {
		output.WriteString("<w:tabs>")
		for _, tab := range attrs.TabStops {
			output.WriteString(`<w:tab w:val="` + tabAlignment(tab.Alignment) + `" w:pos="` + twips(tab.Position) + `" w:leader="` + tabLeader(tab.Leader) + `"/>`)
		}
		output.WriteString("</w:tabs>")
	}
	if attrs.KeepTogetherSpecified() {
		if attrs.KeepTogether {
			output.WriteString("<w:keepLines/>")
		} else {
			output.WriteString(`<w:keepLines w:val="0"/>`)
		}
	}
	if attrs.KeepWithNextSpecified() {
		if attrs.KeepWithNext {
			output.WriteString("<w:keepNext/>")
		} else {
			output.WriteString(`<w:keepNext w:val="0"/>`)
		}
	}
	if attrs.PageBreakBeforeSpecified() {
		if attrs.PageBreakBefore {
			output.WriteString("<w:pageBreakBefore/>")
		} else {
			output.WriteString(`<w:pageBreakBefore w:val="0"/>`)
		}
	}
	if attrs.OutlineLevelSpecified() {
		output.WriteString(`<w:outlineLvl w:val="` + strconv.Itoa(attrs.OutlineLevel) + `"/>`)
	}
	writeParagraphBorders(output, attrs.Borders)
	if attrs.Shading != (ir.Color{}) {
		output.WriteString(`<w:shd w:val="clear" w:fill="` + color(attrs.Shading) + `"/>`)
	}
	if attrs.BiDiSpecified() {
		if attrs.BiDi {
			output.WriteString("<w:bidi/>")
		} else {
			output.WriteString(`<w:bidi w:val="0"/>`)
		}
	}
	output.WriteString("</w:pPr>")
}

func hasParagraphProperties(attrs ir.ParaAttrs) bool {
	return attrs.AlignmentSpecified() || attrs.SpacingSpecified() || attrs.IndentSpecified() || len(attrs.TabStops) > 0 || attrs.KeepTogetherSpecified() || attrs.KeepWithNextSpecified() || attrs.PageBreakBeforeSpecified() || attrs.OutlineLevelSpecified() || attrs.Borders != (ir.ParaBorders{}) || attrs.Shading != (ir.Color{}) || attrs.BiDiSpecified()
}

func writeRun(output *strings.Builder, run ir.Run, state *packageState) {
	if run.Hyperlink != nil {
		if run.Hyperlink.URL != "" && state != nil {
			if id := state.hyperlinkIDs[run.Hyperlink.URL]; id > 0 {
				output.WriteString(`<w:hyperlink r:id="rIdHyperlink` + strconv.Itoa(id) + `">`)
				writeRunContent(output, run)
				output.WriteString(`</w:hyperlink>`)
				return
			}
		} else if run.Hyperlink.Bookmark != "" {
			output.WriteString(`<w:hyperlink w:anchor="` + escape(run.Hyperlink.Bookmark) + `">`)
			writeRunContent(output, run)
			output.WriteString(`</w:hyperlink>`)
			return
		}
	}
	writeRunContent(output, run)
}

func writeRunContent(output *strings.Builder, run ir.Run) {
	output.WriteString("<w:r>")
	writeRunProperties(output, run.Attrs)
	if run.Text != "" {
		for index, text := range strings.Split(run.Text, "\t") {
			if index > 0 {
				output.WriteString("<w:tab/>")
			}
			if text != "" {
				output.WriteString(`<w:t xml:space="preserve">` + escape(text) + `</w:t>`)
			}
		}
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
	if attrs.BoldSpecified() {
		if attrs.Bold {
			output.WriteString("<w:b/>")
		} else {
			output.WriteString(`<w:b w:val="0"/>`)
		}
	}
	if attrs.ItalicSpecified() {
		if attrs.Italic {
			output.WriteString("<w:i/>")
		} else {
			output.WriteString(`<w:i w:val="0"/>`)
		}
	}
	if attrs.UnderlineSpecified() {
		output.WriteString(`<w:u w:val="` + underline(attrs.Underline) + `"/>`)
	}
	if attrs.StrikeSpecified() {
		if attrs.Strike {
			output.WriteString("<w:strike/>")
		} else {
			output.WriteString(`<w:strike w:val="0"/>`)
		}
	}
	if attrs.SmallCapsSpecified() {
		if attrs.SmallCaps {
			output.WriteString("<w:smallCaps/>")
		} else {
			output.WriteString(`<w:smallCaps w:val="0"/>`)
		}
	}
	if attrs.AllCapsSpecified() {
		if attrs.AllCaps {
			output.WriteString("<w:caps/>")
		} else {
			output.WriteString(`<w:caps w:val="0"/>`)
		}
	}
	if attrs.FontSize > 0 {
		output.WriteString(`<w:sz w:val="` + strconv.FormatFloat(attrs.FontSize*2, 'f', -1, 64) + `"/>`)
	}
	if attrs.FontName != "" {
		output.WriteString(`<w:rFonts w:ascii="` + escape(attrs.FontName) + `" w:hAnsi="` + escape(attrs.FontName) + `"/>`)
	}
	if attrs.ColorSpecified() {
		output.WriteString(`<w:color w:val="` + color(attrs.Color) + `"/>`)
	}
	if attrs.BaselineSpecified() {
		output.WriteString(`<w:vertAlign w:val="` + baseline(attrs.Baseline) + `"/>`)
	}
	if attrs.Highlight != (ir.Color{}) {
		output.WriteString(`<w:highlight w:val="` + highlight(attrs.Highlight) + `"/>`)
	}
	if attrs.TrackingSpecified() {
		output.WriteString(`<w:spacing w:val="` + strconv.FormatInt(int64(attrs.Tracking*20), 10) + `"/>`)
	}
	if attrs.Language != "" {
		output.WriteString(`<w:lang w:val="` + escape(attrs.Language) + `"/>`)
	}
	output.WriteString("</w:rPr>")
}

func writeTable(output *strings.Builder, table *ir.Table, imageIndex *int, state *packageState) {
	output.WriteString("<w:tbl>")
	if table.Style != "" || table.Align != ir.AlignLeft || table.Borders != (ir.TableBorders{}) || table.Shading != (ir.Color{}) {
		output.WriteString("<w:tblPr>")
		if table.Style != "" {
			output.WriteString(`<w:tblStyle w:val="` + escape(table.Style) + `"/>`)
		}
		if table.Align != ir.AlignLeft {
			output.WriteString(`<w:jc w:val="` + alignment(table.Align) + `"/>`)
		}
		writeTableBorders(output, table.Borders)
		if table.Shading != (ir.Color{}) {
			output.WriteString(`<w:shd w:val="clear" w:fill="` + color(table.Shading) + `"/>`)
		}
		output.WriteString("</w:tblPr>")
	}
	if len(table.ColumnWidths) > 0 {
		output.WriteString("<w:tblGrid>")
		for _, width := range table.ColumnWidths {
			output.WriteString(`<w:gridCol w:w="` + twips(width) + `"/>`)
		}
		output.WriteString("</w:tblGrid>")
	}
	for _, row := range table.Rows {
		output.WriteString("<w:tr>")
		if row.IsHeader || row.Height != nil {
			output.WriteString("<w:trPr>")
			if row.IsHeader {
				output.WriteString("<w:tblHeader/>")
			}
			if row.Height != nil {
				output.WriteString(`<w:trHeight w:val="` + twips(*row.Height) + `"/>`)
			}
			output.WriteString("</w:trPr>")
		}
		for _, cell := range row.Cells {
			output.WriteString("<w:tc>")
			if cell.ColSpan > 1 || cell.Width > 0 || cell.VAlign != ir.VAlignTop || cell.Borders != (ir.CellBorders{}) || cell.Shading != (ir.Color{}) || cell.Padding != (ir.Padding{}) {
				output.WriteString("<w:tcPr>")
				if cell.Width > 0 {
					output.WriteString(`<w:tcW w:w="` + twips(cell.Width) + `" w:type="dxa"/>`)
				}
				if cell.ColSpan > 1 {
					output.WriteString(`<w:gridSpan w:val="` + strconv.Itoa(cell.ColSpan) + `"/>`)
				}
				if cell.VAlign != ir.VAlignTop {
					output.WriteString(`<w:vAlign w:val="` + verticalAlignment(cell.VAlign) + `"/>`)
				}
				writeCellBorders(output, cell.Borders)
				if cell.Shading != (ir.Color{}) {
					output.WriteString(`<w:shd w:val="clear" w:fill="` + color(cell.Shading) + `"/>`)
				}
				writeCellPadding(output, cell.Padding)
				output.WriteString("</w:tcPr>")
			}
			for _, block := range cell.Blocks {
				writeBlock(output, block, imageIndex, state)
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

func writeImage(output *strings.Builder, image *ir.Image, index int, state *packageState) {
	output.WriteString("<w:p>")
	writeDrawingRun(output, image, index, state)
	output.WriteString("</w:p>")
}

func writeSectionProperties(output *strings.Builder, section ir.Section, references string) {
	output.WriteString("<w:sectPr>" + references)
	if section.Break != "" {
		output.WriteString(`<w:type w:val="` + escape(section.Break) + `"/>`)
	}
	properties := section.Properties
	if properties.Width > 0 || properties.Height > 0 {
		orientation := ""
		if properties.Orientation == ir.Landscape {
			orientation = ` w:orient="landscape"`
		}
		output.WriteString(`<w:pgSz w:w="` + twips(properties.Width) + `" w:h="` + twips(properties.Height) + `"` + orientation + `/>`)
	}
	distances := ` w:gutter="` + twips(properties.Gutter) + `"`
	if properties.HeaderDistance != nil {
		distances += ` w:header="` + twips(*properties.HeaderDistance) + `"`
	}
	if properties.FooterDistance != nil {
		distances += ` w:footer="` + twips(*properties.FooterDistance) + `"`
	}
	output.WriteString(`<w:pgMar w:top="` + twips(properties.MarginTop) + `" w:right="` + twips(properties.MarginRight) + `" w:bottom="` + twips(properties.MarginBottom) + `" w:left="` + twips(properties.MarginLeft) + `"` + distances + `/>`)
	if len(section.Columns) > 0 {
		output.WriteString(`<w:cols w:num="` + strconv.Itoa(len(section.Columns)) + `" w:equalWidth="0">`)
		for _, column := range section.Columns {
			output.WriteString(`<w:col w:w="` + twips(column.Width) + `" w:space="` + twips(column.Spacing) + `"/>`)
		}
		output.WriteString(`</w:cols>`)
	}
	differentFirst := false
	if properties.DifferentFirstPage != nil {
		differentFirst = *properties.DifferentFirstPage
	} else {
		for _, value := range append(append([]*ir.HeaderFooter{section.Header, section.Footer}, section.Headers...), section.Footers...) {
			if value != nil && value.Kind == ir.HeaderFooterFirst {
				differentFirst = true
				break
			}
		}
	}
	if differentFirst {
		output.WriteString("<w:titlePg/>")
	}

	output.WriteString("</w:sectPr>")
}

func (state *packageState) contentTypes() string {
	content := `<Default Extension="rels" ContentType="application/vnd.openxmlformats-package.relationships+xml"/><Default Extension="xml" ContentType="application/xml"/><Override PartName="/word/document.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.document.main+xml"/>`
	if len(state.numbering) > 0 {
		content += `<Override PartName="/word/numbering.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.numbering+xml"/>`
	}
	if len(state.doc.Styles.Named) > 0 || state.doc.Styles.Defaults != (ir.RunAttrs{}) {
		content += `<Override PartName="/word/styles.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.styles+xml"/>`
	}
	for _, part := range state.auxiliary {
		content += `<Override PartName="/word/` + part.filename + `" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.` + part.kind + `+xml"/>`
	}
	if state.hasEvenVariant() {
		content += `<Override PartName="/word/settings.xml" ContentType="application/vnd.openxmlformats-officedocument.wordprocessingml.settings+xml"/>`
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
	if len(state.numbering) > 0 {
		content += `<Relationship Id="rIdNumbering" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/numbering" Target="numbering.xml"/>`
	}
	if len(state.doc.Styles.Named) > 0 || state.doc.Styles.Defaults != (ir.RunAttrs{}) {
		content += `<Relationship Id="rIdStyles" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/styles" Target="styles.xml"/>`
	}
	for _, part := range state.auxiliary {
		content += `<Relationship Id="` + part.relationship + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/` + part.kind + `" Target="` + part.filename + `"/>`
	}
	if state.hasEvenVariant() {
		content += `<Relationship Id="rIdSettings" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/settings" Target="settings.xml"/>`
	}
	if state.hasFootnotes {
		content += `<Relationship Id="rIdFootnotes" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/footnotes" Target="footnotes.xml"/>`
	}
	return relationshipXML(content + state.contentRelationshipEntries())
}
func (state *packageState) contentRelationshipEntries() string {
	content := ""
	for index := range state.images {
		content += `<Relationship Id="rIdImage` + strconv.Itoa(index+1) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/image" Target="media/image` + strconv.Itoa(index+1) + `.` + imageExtension(state.images[index].Format) + `"/>`
	}
	for index, target := range state.hyperlinks {
		content += `<Relationship Id="rIdHyperlink` + strconv.Itoa(index+1) + `" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/hyperlink" Target="` + escape(target) + `" TargetMode="External"/>`
	}
	return content
}
func relationshipXML(content string) string {
	return xmlDeclaration + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships">` + content + `</Relationships>`
}
func (state *packageState) contentRelationships() string {
	return relationshipXML(state.contentRelationshipEntries())
}

func numberingXML(definitions []ir.NumberingDef) string {
	var content strings.Builder
	for index, definition := range definitions {
		id := index + 1
		content.WriteString(`<w:abstractNum w:abstractNumId="` + strconv.Itoa(id) + `">`)
		for _, level := range definition.Levels {
			start := level.Start
			if start == 0 {
				start = 1
			}
			content.WriteString(`<w:lvl w:ilvl="` + strconv.Itoa(level.Level) + `"><w:start w:val="` + strconv.Itoa(start) + `"/><w:numFmt w:val="` + numberFormat(level.Format) + `"/><w:lvlText w:val="` + escape(level.Text) + `"/><w:lvlJc w:val="` + alignment(level.Align) + `"/>`)
			writeParagraphProperties(&content, &ir.Paragraph{Para: ir.ParaAttrs{Indent: level.Indent}})
			writeRunProperties(&content, level.RunAttrs)
			content.WriteString(`</w:lvl>`)
		}
		content.WriteString(`</w:abstractNum><w:num w:numId="` + escape(definition.ID) + `"><w:abstractNumId w:val="` + strconv.Itoa(id) + `"/></w:num>`)
	}
	return xmlDeclaration + `<w:numbering xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` + content.String() + `</w:numbering>`
}

func stylesXML(sheet ir.StyleSheet) string {
	var content strings.Builder
	if sheet.Defaults != (ir.RunAttrs{}) {
		content.WriteString("<w:docDefaults><w:rPrDefault>")
		writeRunProperties(&content, sheet.Defaults)
		content.WriteString("</w:rPrDefault></w:docDefaults>")
	}
	ids := make([]string, 0, len(sheet.Named))
	for id := range sheet.Named {
		ids = append(ids, id)
	}
	slices.Sort(ids)
	for _, id := range ids {
		style := sheet.Named[id]
		kind := "paragraph"
		if style.IsChar {
			kind = "character"
		}
		content.WriteString(`<w:style w:type="` + kind + `" w:styleId="` + escape(style.ID) + `">`)
		if style.Name != "" {
			content.WriteString(`<w:name w:val="` + escape(style.Name) + `"/>`)
		}
		if style.BasedOn != "" {
			content.WriteString(`<w:basedOn w:val="` + escape(style.BasedOn) + `"/>`)
		}
		properties := style.ParaAttrs
		if !style.IsChar && strings.EqualFold(style.ID, "Heading1") && !properties.OutlineLevelSpecified() {
			properties.SetOutlineLevel(0)
		}
		writeParagraphProperties(&content, &ir.Paragraph{Para: properties})
		writeRunProperties(&content, style.RunAttrs)
		content.WriteString("</w:style>")
	}
	return xmlDeclaration + `<w:styles xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">` + content.String() + `</w:styles>`
}

func headerFooterXML(element string, value *ir.HeaderFooter, state *packageState) string {
	var content strings.Builder
	if value != nil {
		index := 0
		for _, block := range value.Blocks {
			writeBlock(&content, block, &index, state)
		}
	}
	return xmlDeclaration + `<w:` + element + ` xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture" xmlns:wps="http://schemas.microsoft.com/office/word/2010/wordprocessingShape">` + content.String() + `</w:` + element + `>`
}

func (state *packageState) footnotesXML() string {
	var content strings.Builder
	for index, note := range state.notes {
		content.WriteString(`<w:footnote w:id="` + strconv.Itoa(index+1) + `">`)
		image := 0
		for _, block := range note.Blocks {
			writeBlock(&content, block, &image, state)
		}
		content.WriteString(`</w:footnote>`)
	}
	return xmlDeclaration + `<w:footnotes xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main" xmlns:r="http://schemas.openxmlformats.org/officeDocument/2006/relationships" xmlns:wp="http://schemas.openxmlformats.org/drawingml/2006/wordprocessingDrawing" xmlns:a="http://schemas.openxmlformats.org/drawingml/2006/main" xmlns:pic="http://schemas.openxmlformats.org/drawingml/2006/picture" xmlns:wps="http://schemas.microsoft.com/office/word/2010/wordprocessingShape">` + content.String() + `</w:footnotes>`
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
	switch value {
	case ir.UnderlineNone:
		return "none"
	case ir.UnderlineSingle:
		return "single"
	case ir.UnderlineDouble:
		return "double"
	case ir.UnderlineWavy:
		return "wave"
	case ir.UnderlineDash:
		return "dash"
	case ir.UnderlineWords:
		return "words"
	case ir.UnderlineThick:
		return "thick"
	case ir.UnderlineDotted:
		return "dotted"
	case ir.UnderlineDottedHeavy:
		return "dottedHeavy"
	case ir.UnderlineDashedHeavy:
		return "dashedHeavy"
	case ir.UnderlineDashLong:
		return "dashLong"
	case ir.UnderlineDashLongHeavy:
		return "dashLongHeavy"
	case ir.UnderlineDotDash:
		return "dotDash"
	case ir.UnderlineDashDotHeavy:
		return "dashDotHeavy"
	case ir.UnderlineDotDotDash:
		return "dotDotDash"
	case ir.UnderlineDashDotDotHeavy:
		return "dashDotDotHeavy"
	case ir.UnderlineWavyHeavy:
		return "wavyHeavy"
	case ir.UnderlineWavyDouble:
		return "wavyDouble"
	}
	return "single"
}

func baseline(value ir.BaselineShift) string {
	if value == ir.BaselineNone {
		return "baseline"
	}
	if value == ir.BaselineSubscript {
		return "subscript"
	}
	return "superscript"
}
func lineRule(value ir.LineRule) string {
	if value == ir.LineRuleExact {
		return "exact"
	}
	if value == ir.LineRuleAtLeast {
		return "atLeast"
	}
	return "auto"
}
func verticalAlignment(value ir.VAlignment) string {
	if value == ir.VAlignCenter {
		return "center"
	}
	if value == ir.VAlignBottom {
		return "bottom"
	}
	return "top"
}
func highlight(value ir.Color) string {
	colors := map[ir.Color]string{
		{R: 255, G: 255, B: 0}:   "yellow",
		{R: 0, G: 255, B: 0}:     "green",
		{R: 0, G: 255, B: 255}:   "cyan",
		{R: 255, G: 0, B: 255}:   "magenta",
		{R: 0, G: 0, B: 255}:     "blue",
		{R: 255, G: 0, B: 0}:     "red",
		{R: 0, G: 0, B: 0}:       "black",
		{R: 255, G: 255, B: 255}: "white",
		{R: 0, G: 0, B: 139}:     "darkBlue",
		{R: 0, G: 139, B: 139}:   "darkCyan",
		{R: 0, G: 100, B: 0}:     "darkGreen",
		{R: 139, G: 0, B: 139}:   "darkMagenta",
		{R: 139, G: 0, B: 0}:     "darkRed",
		{R: 128, G: 128, B: 0}:   "darkYellow",
		{R: 169, G: 169, B: 169}: "darkGray",
		{R: 211, G: 211, B: 211}: "lightGray",
	}
	if name, ok := colors[value]; ok {
		return name
	}
	return "yellow"
}
func writeParagraphBorders(output *strings.Builder, borders ir.ParaBorders) {
	if borders == (ir.ParaBorders{}) {
		return
	}
	output.WriteString("<w:pBdr>")
	writeBorder(output, "top", borders.Top)
	writeBorder(output, "bottom", borders.Bottom)
	writeBorder(output, "left", borders.Left)
	writeBorder(output, "right", borders.Right)
	output.WriteString("</w:pBdr>")
}
func writeTableBorders(output *strings.Builder, borders ir.TableBorders) {
	if borders == (ir.TableBorders{}) {
		return
	}
	output.WriteString("<w:tblBorders>")
	writeBorder(output, "top", borders.Top)
	writeBorder(output, "bottom", borders.Bottom)
	writeBorder(output, "left", borders.Left)
	writeBorder(output, "right", borders.Right)
	writeBorder(output, "insideH", borders.InsideH)
	writeBorder(output, "insideV", borders.InsideV)
	output.WriteString("</w:tblBorders>")
}
func writeCellBorders(output *strings.Builder, borders ir.CellBorders) {
	if borders == (ir.CellBorders{}) {
		return
	}
	output.WriteString("<w:tcBorders>")
	writeBorder(output, "top", borders.Top)
	writeBorder(output, "bottom", borders.Bottom)
	writeBorder(output, "left", borders.Left)
	writeBorder(output, "right", borders.Right)
	output.WriteString("</w:tcBorders>")
}
func writeBorder(output *strings.Builder, edge string, border ir.Border) {
	if border == (ir.Border{}) {
		return
	}
	style := border.Style
	if style == "" {
		style = "single"
	}
	width := border.Width * 8
	if width == 0 {
		width = 4
	}
	output.WriteString(`<w:` + edge + ` w:val="` + escape(style) + `" w:sz="` + strconv.FormatFloat(width, 'f', -1, 64) + `" w:color="` + color(border.Color) + `"/>`)
}
func writeCellPadding(output *strings.Builder, padding ir.Padding) {
	if padding == (ir.Padding{}) {
		return
	}
	output.WriteString("<w:tcMar>")
	for _, edge := range []struct {
		name  string
		value float64
	}{{"top", padding.Top}, {"bottom", padding.Bottom}, {"left", padding.Left}, {"right", padding.Right}} {
		if edge.value > 0 {
			output.WriteString(`<w:` + edge.name + ` w:w="` + twips(edge.value) + `" w:type="dxa"/>`)
		}
	}
	output.WriteString("</w:tcMar>")
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
	switch value {
	case ir.NumFormatDecimal:
		return "decimal"
	case ir.NumFormatLowerAlpha:
		return "lowerLetter"
	case ir.NumFormatUpperAlpha:
		return "upperLetter"
	case ir.NumFormatLowerRoman:
		return "lowerRoman"
	case ir.NumFormatUpperRoman:
		return "upperRoman"
	default:
		return "bullet"
	}
}

const xmlDeclaration = `<?xml version="1.0" encoding="UTF-8" standalone="yes"?>`
const rootRelationships = xmlDeclaration + `<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"><Relationship Id="rId1" Type="http://schemas.openxmlformats.org/officeDocument/2006/relationships/officeDocument" Target="word/document.xml"/></Relationships>`
