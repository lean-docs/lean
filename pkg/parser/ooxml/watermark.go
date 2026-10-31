package ooxml

import (
	"encoding/xml"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

type xmlLegacyPicture struct {
	Shapes     []xmlLegacyShape `xml:"shape"`
	Rectangles []xmlLegacyShape `xml:"rect"`
}
type xmlLegacyShape struct {
	ImageData *struct {
		ID    string     `xml:"id,attr"`
		Title string     `xml:"title,attr"`
		Extra []xml.Attr `xml:",any,attr"`
	} `xml:"imagedata"`
	Type      string `xml:"type,attr"`
	Style     string `xml:"style,attr"`
	FillColor string `xml:"fillcolor,attr"`
	Fill      struct {
		Opacity string `xml:"opacity,attr"`
	} `xml:"fill"`
	TextPath *struct {
		Text  string `xml:"string,attr"`
		Style string `xml:"style,attr"`
	} `xml:"textpath"`
}

func legacyCSS(source string) map[string]string {
	values := make(map[string]string)
	for _, entry := range strings.Split(source, ";") {
		if key, value, ok := strings.Cut(entry, ":"); ok {
			values[strings.TrimSpace(key)] = strings.Trim(strings.TrimSpace(value), `"'`)
		}
	}
	return values
}

func legacyPoints(source string) float64 {
	multiplier := 1.0
	for _, unit := range []struct {
		suffix string
		factor float64
	}{{"pt", 1}, {"px", 0.75}, {"in", 72}, {"cm", 72 / 2.54}, {"mm", 72 / 25.4}} {
		if strings.HasSuffix(source, unit.suffix) {
			multiplier = unit.factor
			source = strings.TrimSuffix(source, unit.suffix)
			break
		}
	}
	value, _ := strconv.ParseFloat(strings.TrimSpace(source), 64)
	return value * multiplier
}

func (picture xmlLegacyPicture) watermark() (*ir.Watermark, bool) {
	if len(picture.Shapes) != 1 {
		return nil, true
	}
	shape := picture.Shapes[0]
	if shape.ImageData == nil && (shape.TextPath == nil || !strings.HasSuffix(shape.Type, "_t136")) {
		return nil, true
	}
	style := legacyCSS(shape.Style)
	if shape.ImageData != nil {
		if style["position"] != "absolute" || !strings.HasPrefix(style["z-index"], "-") {
			return nil, true
		}
		for _, attribute := range shape.ImageData.Extra {
			if attribute.Name.Space != "xmlns" && attribute.Name.Local != "xmlns" {
				return nil, true
			}
		}
	}
	color, ok := parseHexColor(strings.TrimPrefix(shape.FillColor, "#"))
	if !ok && shape.ImageData == nil {
		return nil, true
	}
	opacity := 1.0
	if shape.Fill.Opacity != "" {
		var err error
		opacity, err = strconv.ParseFloat(strings.TrimSuffix(shape.Fill.Opacity, "%"), 64)
		if err != nil {
			return nil, true
		}
		if strings.HasSuffix(shape.Fill.Opacity, "%") {
			opacity /= 100
		}
	}
	rotation := 0.0
	if source := style["rotation"]; source != "" {
		var err error
		rotation, err = strconv.ParseFloat(source, 64)
		if err != nil {
			return nil, true
		}
	}
	text, font := "", ""
	if shape.TextPath != nil {
		text, font = shape.TextPath.Text, legacyCSS(shape.TextPath.Style)["font-family"]
	}
	return &ir.Watermark{Text: text, FontName: font, Width: legacyPoints(style["width"]), Height: legacyPoints(style["height"]), Rotation: rotation, Color: color, Opacity: opacity, HorizontalAlign: style["mso-position-horizontal"], VerticalAlign: style["mso-position-vertical"], HorizontalRelative: style["mso-position-horizontal-relative"], VerticalRelative: style["mso-position-vertical-relative"], Left: legacyPoints(style["margin-left"]), Top: legacyPoints(style["margin-top"])}, false
}
