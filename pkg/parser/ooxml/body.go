package ooxml

import (
	"encoding/xml"
	"fmt"
	"strconv"

	"github.com/lean-docs/lean/pkg/ir"
	"github.com/lean-docs/lean/pkg/units"
)

// ---------------------------------------------------------------------------
// XML shape (only what the current scope needs)
// ---------------------------------------------------------------------------

const wNS = "http://schemas.openxmlformats.org/wordprocessingml/2006/main"

type xmlDocument struct {
	XMLName xml.Name `xml:"document"`
	Body    xmlBody  `xml:"body"`
}

type xmlBody struct {
	Paragraphs []xmlParagraph `xml:"p"`
}

type xmlParagraph struct {
	Props *xmlParaProps `xml:"pPr"`
	Runs  []xmlRun      `xml:"r"`
}

type xmlParaProps struct {
	Align           *xmlValAttr `xml:"jc"`
	Spacing         *xmlSpacing `xml:"spacing"`
	Indent          *xmlIndent  `xml:"ind"`
	Tabs            *xmlTabs    `xml:"tabs"`
	KeepLines       *xmlToggle  `xml:"keepLines"`
	KeepNext        *xmlToggle  `xml:"keepNext"`
	PageBreakBefore *xmlToggle  `xml:"pageBreakBefore"`
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
	Props  *xmlRunProps `xml:"rPr"`
	Text   []xmlText    `xml:"t"`
	Breaks []xmlBreak   `xml:"br"`
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
	Size      *xmlValAttr  `xml:"sz"`     // half-points
	Fonts     *xmlRunFonts `xml:"rFonts"`
	Color     *xmlValAttr  `xml:"color"`
	Highlight *xmlValAttr  `xml:"highlight"`
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

// ---------------------------------------------------------------------------
// Parsing
// ---------------------------------------------------------------------------

func parseDocument(xmlBytes []byte, doc *ir.Document) error {
	var x xmlDocument
	if err := xml.Unmarshal(xmlBytes, &x); err != nil {
		return fmt.Errorf("ooxml: parse document.xml: %w", err)
	}

	section := ir.Section{}
	for i, xp := range x.Body.Paragraphs {
		p := &ir.Paragraph{ID: fmt.Sprintf("p%d", i+1)}
		if xp.Props != nil {
			applyParaProps(&p.Para, xp.Props)
		}
		for _, xr := range xp.Runs {
			run := convertRun(xr)
			p.Runs = append(p.Runs, run)
		}
		section.Blocks = append(section.Blocks, p)
	}
	doc.Sections = append(doc.Sections, section)
	return nil
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

// twipAttr reads a twips-valued string attribute and returns its point value.
// Missing or malformed input returns 0.
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

func breakTypeFromVal(v string) ir.BreakType {
	switch v {
	case "page":
		return ir.BreakPage
	case "column":
		return ir.BreakColumn
	}
	return ir.BreakLine
}

func convertRun(xr xmlRun) ir.Run {
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
	return run
}

// toggleOn interprets an OOXML toggle property. Presence with no val or
// val="1"/"true"/"on" → true; val="0"/"false"/"off" → false.
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
	if len(s) != 6 {
		return ir.Color{}, false
	}
	n, err := strconv.ParseUint(s, 16, 32)
	if err != nil {
		return ir.Color{}, false
	}
	return ir.Color{R: uint8(n >> 16), G: uint8(n >> 8), B: uint8(n)}, true
}

// highlightFromName maps OOXML w:highlight keywords to approximate RGB.
// Word's highlight palette is a fixed set of named colors.
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

// silence unused warning for wNS when we add namespace-validated parsing later
var _ = wNS
