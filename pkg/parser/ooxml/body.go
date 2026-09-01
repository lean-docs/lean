package ooxml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
	"github.com/lean-docs/lean/pkg/units"
)

// ---------------------------------------------------------------------------
// XML shape (only what the current scope needs)
// ---------------------------------------------------------------------------

type xmlParagraph struct {
	Props *xmlParaProps `xml:"pPr"`
	Runs  []xmlRun      `xml:"r"`
}

func (paragraph *xmlParagraph) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch element := token.(type) {
		case xml.StartElement:
			switch element.Name.Local {
			case "pPr":
				var properties xmlParaProps
				if err := decoder.DecodeElement(&properties, &element); err != nil {
					return err
				}
				paragraph.Props = &properties
			case "r":
				var run xmlRun
				if err := decoder.DecodeElement(&run, &element); err != nil {
					return err
				}
				paragraph.Runs = append(paragraph.Runs, run)
			case "hyperlink":
				var hyperlink xmlHyperlink
				if err := decoder.DecodeElement(&hyperlink, &element); err != nil {
					return err
				}
				for index := range hyperlink.Runs {
					hyperlink.Runs[index].HyperlinkID = hyperlink.ID
					hyperlink.Runs[index].HyperlinkAnchor = hyperlink.Anchor
				}
				paragraph.Runs = append(paragraph.Runs, hyperlink.Runs...)
			default:
				if err := decoder.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if element.Name.Local == start.Name.Local {
				return nil
			}
		}
	}
}

type xmlHyperlink struct {
	ID     string   `xml:"id,attr"`
	Anchor string   `xml:"anchor,attr"`
	Runs   []xmlRun `xml:"r"`
}

type xmlParaProps struct {
	Align           *xmlValAttr `xml:"jc"`
	Spacing         *xmlSpacing `xml:"spacing"`
	Indent          *xmlIndent  `xml:"ind"`
	Tabs            *xmlTabs    `xml:"tabs"`
	KeepLines       *xmlToggle  `xml:"keepLines"`
	KeepNext        *xmlToggle  `xml:"keepNext"`
	PageBreakBefore *xmlToggle  `xml:"pageBreakBefore"`
	OutlineLevel    *xmlValAttr `xml:"outlineLvl"`
	Borders         *xmlBorders `xml:"pBdr"`
	Shading         *xmlShading `xml:"shd"`
	BiDi            *xmlToggle  `xml:"bidi"`
}

type xmlSpacing struct {
	Before   string `xml:"before,attr"`
	After    string `xml:"after,attr"`
	Line     string `xml:"line,attr"`
	LineRule string `xml:"lineRule,attr"`
}

type xmlIndent struct {
	Left      string `xml:"left,attr"`
	Right     string `xml:"right,attr"`
	FirstLine string `xml:"firstLine,attr"`
	Hanging   string `xml:"hanging,attr"`
}

type xmlTabs struct {
	Tabs []xmlTab `xml:"tab"`
}

type xmlTab struct {
	Val    string `xml:"val,attr"`
	Pos    string `xml:"pos,attr"`
	Leader string `xml:"leader,attr"`
}

type xmlRun struct {
	Props           *xmlRunProps `xml:"rPr"`
	Text            []xmlText    `xml:"t"`
	Breaks          []xmlBreak   `xml:"br"`
	Drawings        []xmlDrawing `xml:"drawing"`
	HyperlinkID     string       `xml:"-"`
	HyperlinkAnchor string       `xml:"-"`
}

type xmlDrawing struct {
	Inline *xmlDrawingContainer `xml:"inline"`
	Anchor *xmlDrawingContainer `xml:"anchor"`
}

type xmlDrawingContainer struct {
	Extent  *xmlExtent `xml:"extent"`
	DocPr   *xmlDocPr  `xml:"docPr"`
	Graphic xmlGraphic `xml:"graphic"`
}

type xmlExtent struct {
	CX string `xml:"cx,attr"`
	CY string `xml:"cy,attr"`
}

type xmlDocPr struct {
	Description string `xml:"descr,attr"`
}

type xmlGraphic struct {
	Data xmlGraphicData `xml:"graphicData"`
}

type xmlGraphicData struct {
	Picture xmlPicture `xml:"pic"`
}

type xmlPicture struct {
	BlipFill xmlBlipFill `xml:"blipFill"`
}

type xmlBlipFill struct {
	Blip xmlBlip `xml:"blip"`
}

type xmlBlip struct {
	Embed string `xml:"embed,attr"`
}

type xmlBreak struct {
	Type string `xml:"type,attr"`
}

type xmlText struct {
	Space string `xml:"space,attr"`
	Value string `xml:",chardata"`
}

type xmlRunProps struct {
	Bold      *xmlToggle   `xml:"b"`
	Italic    *xmlToggle   `xml:"i"`
	Underline *xmlValAttr  `xml:"u"`
	Strike    *xmlToggle   `xml:"strike"`
	SmallCaps *xmlToggle   `xml:"smallCaps"`
	AllCaps   *xmlToggle   `xml:"caps"`
	Vert      *xmlValAttr  `xml:"vertAlign"`
	Size      *xmlValAttr  `xml:"sz"`
	Fonts     *xmlRunFonts `xml:"rFonts"`
	Color     *xmlValAttr  `xml:"color"`
	Highlight *xmlValAttr  `xml:"highlight"`
	Spacing   *xmlValAttr  `xml:"spacing"`
	Language  *xmlValAttr  `xml:"lang"`
}

type xmlToggle struct {
	Val string `xml:"val,attr"`
}

type xmlValAttr struct {
	Val string `xml:"val,attr"`
}

type xmlRunFonts struct {
	ASCII    string `xml:"ascii,attr"`
	HAnsi    string `xml:"hAnsi,attr"`
	EastAsia string `xml:"eastAsia,attr"`
}

// --- table ---

type xmlTable struct {
	Props *xmlTblProps  `xml:"tblPr"`
	Grid  *xmlTblGrid   `xml:"tblGrid"`
	Rows  []xmlTableRow `xml:"tr"`
}

type xmlTblProps struct {
	Style   *xmlValAttr `xml:"tblStyle"`
	Align   *xmlValAttr `xml:"jc"`
	Borders *xmlBorders `xml:"tblBorders"`
	Shading *xmlShading `xml:"shd"`
}

type xmlTblGrid struct {
	Cols []xmlGridCol `xml:"gridCol"`
}

type xmlGridCol struct {
	W string `xml:"w,attr"`
}

type xmlTableRow struct {
	Props *xmlRowProps   `xml:"trPr"`
	Cells []xmlTableCell `xml:"tc"`
}

type xmlRowProps struct {
	Header *xmlToggle  `xml:"tblHeader"`
	Height *xmlValAttr `xml:"trHeight"`
}

type xmlTableCell struct {
	Props      *xmlCellProps  `xml:"tcPr"`
	Paragraphs []xmlParagraph `xml:"p"`
	// nested tables parsed via custom UnmarshalXML on xmlTableCell
	Tables []xmlTable `xml:"tbl"`
}

type xmlCellProps struct {
	Width    *xmlCellW   `xml:"tcW"`
	GridSpan *xmlValAttr `xml:"gridSpan"`
	VMerge   *xmlVMerge  `xml:"vMerge"`
	VAlign   *xmlValAttr `xml:"vAlign"`
	Borders  *xmlBorders `xml:"tcBorders"`
	Shading  *xmlShading `xml:"shd"`
	Margin   *xmlTcMar   `xml:"tcMar"`
}

type xmlCellW struct {
	W    string `xml:"w,attr"`
	Type string `xml:"type,attr"`
}

type xmlVMerge struct {
	Val string `xml:"val,attr"` // "restart" or absent (= continue)
}

type xmlBorders struct {
	Top     *xmlBorder `xml:"top"`
	Bottom  *xmlBorder `xml:"bottom"`
	Left    *xmlBorder `xml:"left"`
	Right   *xmlBorder `xml:"right"`
	InsideH *xmlBorder `xml:"insideH"`
	InsideV *xmlBorder `xml:"insideV"`
}

type xmlBorder struct {
	Val   string `xml:"val,attr"`
	Sz    string `xml:"sz,attr"` // eighths of a point
	Color string `xml:"color,attr"`
}

type xmlShading struct {
	Fill  string `xml:"fill,attr"`
	Color string `xml:"color,attr"`
	Val   string `xml:"val,attr"`
}

type xmlTcMar struct {
	Top    *xmlMarEdge `xml:"top"`
	Bottom *xmlMarEdge `xml:"bottom"`
	Left   *xmlMarEdge `xml:"left"`
	Right  *xmlMarEdge `xml:"right"`
}

type xmlMarEdge struct {
	W string `xml:"w,attr"`
}

type xmlSectionProps struct {
	Size    *xmlPageSize    `xml:"pgSz"`
	Margins *xmlPageMargins `xml:"pgMar"`
}

type xmlPageSize struct {
	Width       string `xml:"w,attr"`
	Height      string `xml:"h,attr"`
	Orientation string `xml:"orient,attr"`
}

type xmlPageMargins struct {
	Top    string `xml:"top,attr"`
	Right  string `xml:"right,attr"`
	Bottom string `xml:"bottom,attr"`
	Left   string `xml:"left,attr"`
}

type imageResource struct {
	Data   []byte
	Format ir.ImageFormat
}

// ---------------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------------

// parseDocument reads document.xml and appends a populated Section to doc.
// Body children (w:p, w:tbl) are handled in order; unknown elements skipped.
const wNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

func parseDocument(xmlBytes []byte, doc *ir.Document, images map[string]imageResource, hyperlinks map[string]string) error {
	dec := xml.NewDecoder(bytes.NewReader(xmlBytes))
	if err := advanceToNS(dec, wNS, "document"); err != nil {
		return fmt.Errorf("ooxml: could not find <w:document>: %w", err)
	}
	if err := advanceToNS(dec, wNS, "body"); err != nil {
		return fmt.Errorf("ooxml: could not find <w:body>: %w", err)
	}

	section := ir.Section{}
	ctx := &parseCtx{section: &section, images: images, hyperlinks: hyperlinks}
	if err := parseBlocks(dec, "body", &section.Blocks, ctx); err != nil {
		return err
	}
	doc.Sections = append(doc.Sections, section)
	return nil
}

// parseCtx carries counters used for block ID generation.
type parseCtx struct {
	paraSeq    int
	tableSeq   int
	imageSeq   int
	section    *ir.Section
	images     map[string]imageResource
	hyperlinks map[string]string
}

func (c *parseCtx) nextParaID() string {
	c.paraSeq++
	return fmt.Sprintf("p%d", c.paraSeq)
}
func (c *parseCtx) nextTableID() string {
	c.tableSeq++
	return fmt.Sprintf("t%d", c.tableSeq)
}

func (c *parseCtx) nextImageID() string {
	c.imageSeq++
	return fmt.Sprintf("image%d", c.imageSeq)
}

// advanceToNS consumes tokens until a StartElement with the given namespace
// and local name is seen. An empty ns matches any namespace.
func advanceToNS(dec *xml.Decoder, ns, local string) error {
	for {
		tok, err := dec.Token()
		if err != nil {
			return fmt.Errorf("ooxml: no <%s>: %w", local, err)
		}
		if se, ok := tok.(xml.StartElement); ok && se.Name.Local == local {
			if ns == "" || se.Name.Space == ns {
				return nil
			}
		}
	}
}

// parseBlocks walks children of the current element, emitting IR blocks.
// `parentLocal` is the local name of the element whose children we consume;
// we stop at its matching end token.
func parseBlocks(dec *xml.Decoder, parentLocal string, blocks *[]ir.Block, ctx *parseCtx) error {
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("ooxml: read %s children: %w", parentLocal, err)
		}
		switch t := tok.(type) {
		case xml.StartElement:
			switch t.Name.Local {
			case "p":
				var xp xmlParagraph
				if err := dec.DecodeElement(&xp, &t); err != nil {
					return err
				}
				*blocks = append(*blocks, convertParagraphBlocks(xp, ctx)...)
			case "tbl":
				var xt xmlTable
				if err := dec.DecodeElement(&xt, &t); err != nil {
					return err
				}
				*blocks = append(*blocks, convertTable(xt, ctx))
			case "sectPr":
				var properties xmlSectionProps
				if err := dec.DecodeElement(&properties, &t); err != nil {
					return err
				}
				applySectionProps(ctx.section, &properties)
			default:
				if err := dec.Skip(); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if t.Name.Local == parentLocal {
				return nil
			}
		}
	}
}

func applySectionProps(section *ir.Section, properties *xmlSectionProps) {
	if section == nil || properties == nil {
		return
	}
	if properties.Size != nil {
		section.Properties.Width = twipAttr(properties.Size.Width)
		section.Properties.Height = twipAttr(properties.Size.Height)
		if properties.Size.Orientation == "landscape" {
			section.Properties.Orientation = ir.Landscape
		}
	}
	if properties.Margins != nil {
		section.Properties.MarginTop = twipAttr(properties.Margins.Top)
		section.Properties.MarginRight = twipAttr(properties.Margins.Right)
		section.Properties.MarginBottom = twipAttr(properties.Margins.Bottom)
		section.Properties.MarginLeft = twipAttr(properties.Margins.Left)
	}
}

func convertParagraph(xp xmlParagraph, id string, context *parseCtx) *ir.Paragraph {
	p := &ir.Paragraph{ID: id}
	if xp.Props != nil {
		applyParaProps(&p.Para, xp.Props)
	}
	for _, xr := range xp.Runs {
		p.Runs = append(p.Runs, convertRun(xr, context))
	}
	return p
}

func convertParagraphBlocks(paragraph xmlParagraph, context *parseCtx) []ir.Block {
	blocks := make([]ir.Block, 0, 1)
	if paragraph.Props != nil || paragraphHasText(paragraph) || !paragraphHasDrawing(paragraph) {
		blocks = append(blocks, convertParagraph(paragraph, context.nextParaID(), context))
	}
	for _, run := range paragraph.Runs {
		for _, drawing := range run.Drawings {
			if image := convertDrawing(drawing, context); image != nil {
				blocks = append(blocks, image)
			}
		}
	}
	return blocks
}

func paragraphHasText(paragraph xmlParagraph) bool {
	for _, run := range paragraph.Runs {
		if len(run.Text) > 0 || len(run.Breaks) > 0 {
			return true
		}
	}
	return false
}

func paragraphHasDrawing(paragraph xmlParagraph) bool {
	for _, run := range paragraph.Runs {
		if len(run.Drawings) > 0 {
			return true
		}
	}
	return false
}

func convertDrawing(drawing xmlDrawing, context *parseCtx) *ir.Image {
	container := drawing.Inline
	float := ir.FloatNone
	if container == nil {
		container = drawing.Anchor
		float = ir.FloatLeft
	}
	if container == nil {
		return nil
	}
	resource, ok := context.images[container.Graphic.Data.Picture.BlipFill.Blip.Embed]
	if !ok {
		return nil
	}
	image := &ir.Image{ID: context.nextImageID(), Data: resource.Data, Format: resource.Format, Float: float}
	if container.DocPr != nil {
		image.Alt = container.DocPr.Description
	}
	if container.Extent != nil {
		image.Width = emuAttr(container.Extent.CX)
		image.Height = emuAttr(container.Extent.CY)
	}
	return image
}

func convertRun(xr xmlRun, context *parseCtx) ir.Run {
	var run ir.Run
	for _, t := range xr.Text {
		run.Text += t.Value
	}
	if len(xr.Breaks) > 0 {
		run.Break = breakTypeFromVal(xr.Breaks[len(xr.Breaks)-1].Type)
	}
	if xr.Props != nil {
		applyRunProps(&run.Attrs, xr.Props)
	}
	if xr.HyperlinkID != "" {
		if target := context.hyperlinks[xr.HyperlinkID]; target != "" {
			run.Hyperlink = &ir.Hyperlink{URL: target}
		}
	} else if xr.HyperlinkAnchor != "" {
		run.Hyperlink = &ir.Hyperlink{Bookmark: xr.HyperlinkAnchor}
	}
	return run
}

func toggleOn(t *xmlToggle) bool {
	if t == nil {
		return false
	}
	switch t.Val {
	case "", "1", "true", "on":
		return true
	}
	return false
}

func applyRunProps(a *ir.RunAttrs, p *xmlRunProps) {
	if toggleOn(p.Bold) {
		a.Bold = true
	}
	if toggleOn(p.Italic) {
		a.Italic = true
	}
	if p.Underline != nil {
		a.Underline = underlineFromVal(p.Underline.Val)
	}
	if toggleOn(p.Strike) {
		a.Strike = true
	}
	if toggleOn(p.SmallCaps) {
		a.SmallCaps = true
	}
	if toggleOn(p.AllCaps) {
		a.AllCaps = true
	}
	if p.Vert != nil {
		a.Baseline = baselineFromVal(p.Vert.Val)
	}
	if p.Size != nil {
		if hp, err := strconv.ParseFloat(p.Size.Val, 64); err == nil {
			a.FontSize = units.HalfPointsToPoints(hp)
		}
	}
	if p.Fonts != nil {
		switch {
		case p.Fonts.ASCII != "":
			a.FontName = p.Fonts.ASCII
		case p.Fonts.HAnsi != "":
			a.FontName = p.Fonts.HAnsi
		case p.Fonts.EastAsia != "":
			a.FontName = p.Fonts.EastAsia
		}
	}
	if p.Color != nil && p.Color.Val != "" && p.Color.Val != "auto" {
		if c, ok := parseHexColor(p.Color.Val); ok {
			a.Color = c
		}
	}
	if p.Highlight != nil && p.Highlight.Val != "" {
		a.Highlight = highlightFromName(p.Highlight.Val)
	}
	if p.Spacing != nil {
		if twentieths, err := strconv.ParseFloat(p.Spacing.Val, 64); err == nil {
			a.Tracking = twentieths / 20
		}
	}
	if p.Language != nil {
		a.Language = p.Language.Val
	}
}

func applyParaProps(a *ir.ParaAttrs, p *xmlParaProps) {
	if p.Align != nil {
		a.Align = alignFromVal(p.Align.Val)
	}
	if p.Spacing != nil {
		a.Spacing.Before = twipAttr(p.Spacing.Before)
		a.Spacing.After = twipAttr(p.Spacing.After)
		a.Spacing.Line = twipAttr(p.Spacing.Line)
		a.Spacing.LineRule = lineRuleFromVal(p.Spacing.LineRule)
	}
	if p.Indent != nil {
		a.Indent.Left = twipAttr(p.Indent.Left)
		a.Indent.Right = twipAttr(p.Indent.Right)
		a.Indent.FirstLine = twipAttr(p.Indent.FirstLine)
		a.Indent.Hanging = twipAttr(p.Indent.Hanging)
	}
	if p.Tabs != nil {
		for _, t := range p.Tabs.Tabs {
			a.TabStops = append(a.TabStops, ir.TabStop{
				Position:  twipAttr(t.Pos),
				Alignment: tabAlignFromVal(t.Val),
				Leader:    tabLeaderFromVal(t.Leader),
			})
		}
	}
	if toggleOn(p.KeepLines) {
		a.KeepTogether = true
	}
	if toggleOn(p.KeepNext) {
		a.KeepWithNext = true
	}
	if toggleOn(p.PageBreakBefore) {
		a.PageBreakBefore = true
	}
	if p.OutlineLevel != nil {
		if level, err := strconv.Atoi(p.OutlineLevel.Val); err == nil {
			a.OutlineLevel = level
		}
	}
	if p.Borders != nil {
		a.Borders.Top = borderFrom(p.Borders.Top)
		a.Borders.Bottom = borderFrom(p.Borders.Bottom)
		a.Borders.Left = borderFrom(p.Borders.Left)
		a.Borders.Right = borderFrom(p.Borders.Right)
	}
	if p.Shading != nil {
		if color, ok := parseHexColor(p.Shading.Fill); ok {
			a.Shading = color
		}
	}
	if toggleOn(p.BiDi) {
		a.BiDi = true
	}
}

// --- table conversion ---

func convertTable(xt xmlTable, ctx *parseCtx) *ir.Table {
	tbl := &ir.Table{ID: ctx.nextTableID()}
	if xt.Props != nil {
		applyTableProps(tbl, xt.Props)
	}
	if xt.Grid != nil {
		for _, c := range xt.Grid.Cols {
			tbl.ColumnWidths = append(tbl.ColumnWidths, twipAttr(c.W))
		}
	}
	for _, xr := range xt.Rows {
		row := ir.TableRow{}
		if xr.Props != nil {
			if toggleOn(xr.Props.Header) {
				row.IsHeader = true
			}
			if xr.Props.Height != nil && xr.Props.Height.Val != "" {
				h := twipAttr(xr.Props.Height.Val)
				row.Height = &h
			}
		}
		for _, xc := range xr.Cells {
			row.Cells = append(row.Cells, convertCell(xc, ctx))
		}
		tbl.Rows = append(tbl.Rows, row)
	}
	applyRowSpans(tbl, xt.Rows)
	return tbl
}

// applyRowSpans fills RowSpan on origin cells (vMerge="restart") by counting
// subsequent rows where the same grid column continues the merge. Grid column
// is computed by summing gridSpan values left of each cell, so rows with mixed
// column spans are handled correctly.
func applyRowSpans(tbl *ir.Table, rows []xmlTableRow) {
	gridCol := func(row, cellIdx int) int {
		col := 0
		for c := 0; c < cellIdx && c < len(rows[row].Cells); c++ {
			span := cellGridSpan(rows[row].Cells[c])
			col += span
		}
		return col
	}

	vmergeKind := func(c xmlTableCell) string {
		if c.Props == nil || c.Props.VMerge == nil {
			return ""
		}
		if c.Props.VMerge.Val == "" {
			return "continue"
		}
		return c.Props.VMerge.Val
	}

	// Build a map from grid column → cell index for each row.
	cellAtGridCol := func(row, targetCol int) (int, bool) {
		col := 0
		for c := range rows[row].Cells {
			if col == targetCol {
				return c, true
			}
			col += cellGridSpan(rows[row].Cells[c])
			if col > targetCol {
				return 0, false
			}
		}
		return 0, false
	}

	for i := range rows {
		for j := range rows[i].Cells {
			if vmergeKind(rows[i].Cells[j]) != "restart" {
				continue
			}
			col := gridCol(i, j)
			span := 1
			for k := i + 1; k < len(rows); k++ {
				ci, ok := cellAtGridCol(k, col)
				if !ok || vmergeKind(rows[k].Cells[ci]) != "continue" {
					break
				}
				span++
			}
			tbl.Rows[i].Cells[j].RowSpan = span
		}
	}
}

func cellGridSpan(c xmlTableCell) int {
	if c.Props != nil && c.Props.GridSpan != nil {
		if n, err := strconv.Atoi(c.Props.GridSpan.Val); err == nil && n > 0 {
			return n
		}
	}
	return 1
}

func applyTableProps(t *ir.Table, p *xmlTblProps) {
	if p.Style != nil {
		t.Style = p.Style.Val
	}
	if p.Align != nil {
		t.Align = alignFromVal(p.Align.Val)
	}
	if p.Borders != nil {
		t.Borders.Top = borderFrom(p.Borders.Top)
		t.Borders.Bottom = borderFrom(p.Borders.Bottom)
		t.Borders.Left = borderFrom(p.Borders.Left)
		t.Borders.Right = borderFrom(p.Borders.Right)
		t.Borders.InsideH = borderFrom(p.Borders.InsideH)
		t.Borders.InsideV = borderFrom(p.Borders.InsideV)
	}
	if p.Shading != nil {
		if c, ok := parseHexColor(p.Shading.Fill); ok {
			t.Shading = c
		}
	}
}

func convertCell(xc xmlTableCell, ctx *parseCtx) ir.TableCell {
	cell := ir.TableCell{ColSpan: 1, RowSpan: 1}
	if xc.Props != nil {
		applyCellProps(&cell, xc.Props)
	}
	for _, xp := range xc.Paragraphs {
		cell.Blocks = append(cell.Blocks, convertParagraphBlocks(xp, ctx)...)
	}
	for _, xt := range xc.Tables {
		cell.Blocks = append(cell.Blocks, convertTable(xt, ctx))
	}
	return cell
}

func applyCellProps(c *ir.TableCell, p *xmlCellProps) {
	if p.Width != nil {
		c.Width = twipAttr(p.Width.W)
	}
	if p.GridSpan != nil {
		if n, err := strconv.Atoi(p.GridSpan.Val); err == nil && n > 0 {
			c.ColSpan = n
		}
	}
	if p.VAlign != nil {
		c.VAlign = vAlignFromVal(p.VAlign.Val)
	}
	if p.Borders != nil {
		c.Borders.Top = borderFrom(p.Borders.Top)
		c.Borders.Bottom = borderFrom(p.Borders.Bottom)
		c.Borders.Left = borderFrom(p.Borders.Left)
		c.Borders.Right = borderFrom(p.Borders.Right)
	}
	if p.Shading != nil {
		if col, ok := parseHexColor(p.Shading.Fill); ok {
			c.Shading = col
		}
	}
	if p.Margin != nil {
		c.Padding = padFromTcMar(p.Margin)
	}
}

func padFromTcMar(m *xmlTcMar) ir.Padding {
	pad := ir.Padding{}
	if m.Top != nil {
		pad.Top = twipAttr(m.Top.W)
	}
	if m.Bottom != nil {
		pad.Bottom = twipAttr(m.Bottom.W)
	}
	if m.Left != nil {
		pad.Left = twipAttr(m.Left.W)
	}
	if m.Right != nil {
		pad.Right = twipAttr(m.Right.W)
	}
	return pad
}

func borderFrom(b *xmlBorder) ir.Border {
	if b == nil {
		return ir.Border{}
	}
	out := ir.Border{Style: b.Val}
	if b.Sz != "" {
		if eighths, err := strconv.ParseFloat(b.Sz, 64); err == nil {
			out.Width = eighths / 8.0 // sz in eighths of a point
		}
	}
	if c, ok := parseHexColor(b.Color); ok {
		out.Color = c
	}
	return out
}

func alignFromVal(v string) ir.Alignment {
	switch v {
	case "center":
		return ir.AlignCenter
	case "right", "end":
		return ir.AlignRight
	case "both", "distribute", "justify":
		return ir.AlignJustify
	}
	return ir.AlignLeft
}

func lineRuleFromVal(v string) ir.LineRule {
	switch v {
	case "exact":
		return ir.LineRuleExact
	case "atLeast":
		return ir.LineRuleAtLeast
	}
	return ir.LineRuleAuto
}

func tabAlignFromVal(v string) ir.TabAlignment {
	switch v {
	case "right":
		return ir.TabAlignRight
	case "center":
		return ir.TabAlignCenter
	case "decimal":
		return ir.TabAlignDecimal
	}
	return ir.TabAlignLeft
}

func tabLeaderFromVal(v string) ir.TabLeader {
	switch v {
	case "dot":
		return ir.TabLeaderDot
	case "hyphen":
		return ir.TabLeaderDash
	case "underscore":
		return ir.TabLeaderUnderscore
	}
	return ir.TabLeaderNone
}

func vAlignFromVal(v string) ir.VAlignment {
	switch v {
	case "center":
		return ir.VAlignCenter
	case "bottom":
		return ir.VAlignBottom
	}
	return ir.VAlignTop
}

func twipAttr(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return units.TwipsToPoints(v)
}

func emuAttr(value string) float64 {
	emu, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0
	}
	return emu / 12700
}

func breakTypeFromVal(v string) ir.BreakType {
	switch v {
	case "page":
		return ir.BreakPage
	case "column":
		return ir.BreakColumn
	}
	return ir.BreakLine
}

func underlineFromVal(v string) ir.UnderlineStyle {
	switch v {
	case "", "single":
		return ir.UnderlineSingle
	case "double":
		return ir.UnderlineDouble
	case "wavy", "wave":
		return ir.UnderlineWavy
	case "dash":
		return ir.UnderlineDash
	case "none":
		return ir.UnderlineNone
	}
	return ir.UnderlineSingle
}

func baselineFromVal(v string) ir.BaselineShift {
	switch v {
	case "superscript":
		return ir.BaselineSuperscript
	case "subscript":
		return ir.BaselineSubscript
	}
	return ir.BaselineNone
}

func parseHexColor(s string) (ir.Color, bool) {
	s = strings.TrimPrefix(s, "#")
	if len(s) != 6 {
		return ir.Color{}, false
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return ir.Color{}, false
	}
	return ir.Color{R: uint8(n >> 16), G: uint8(n >> 8), B: uint8(n)}, true
}

func highlightFromName(name string) ir.Color {
	switch name {
	case "yellow":
		return ir.Color{R: 255, G: 255, B: 0}
	case "green":
		return ir.Color{R: 0, G: 255, B: 0}
	case "cyan":
		return ir.Color{R: 0, G: 255, B: 255}
	case "magenta":
		return ir.Color{R: 255, G: 0, B: 255}
	case "blue":
		return ir.Color{R: 0, G: 0, B: 255}
	case "red":
		return ir.Color{R: 255, G: 0, B: 0}
	case "darkBlue":
		return ir.Color{R: 0, G: 0, B: 139}
	case "darkCyan":
		return ir.Color{R: 0, G: 139, B: 139}
	case "darkGreen":
		return ir.Color{R: 0, G: 100, B: 0}
	case "darkMagenta":
		return ir.Color{R: 139, G: 0, B: 139}
	case "darkRed":
		return ir.Color{R: 139, G: 0, B: 0}
	case "darkYellow":
		return ir.Color{R: 128, G: 128, B: 0}
	case "darkGray":
		return ir.Color{R: 169, G: 169, B: 169}
	case "lightGray":
		return ir.Color{R: 211, G: 211, B: 211}
	case "black":
		return ir.Color{R: 0, G: 0, B: 0}
	case "white":
		return ir.Color{R: 255, G: 255, B: 255}
	case "none":
		return ir.Color{None: true}
	}
	return ir.Color{}
}
