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
	Runs []xmlRun `xml:"r"`
}

type xmlRun struct {
	Props *xmlRunProps `xml:"rPr"`
	Text  []xmlText    `xml:"t"`
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
		for _, xr := range xp.Runs {
			run := convertRun(xr)
			p.Runs = append(p.Runs, run)
		}
		section.Blocks = append(section.Blocks, p)
	}
	doc.Sections = append(doc.Sections, section)
	return nil
}

func convertRun(xr xmlRun) ir.Run {
	var run ir.Run
	for _, t := range xr.Text {
		run.Text += t.Value
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
