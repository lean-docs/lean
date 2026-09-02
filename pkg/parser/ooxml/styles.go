package ooxml

import (
	"encoding/xml"
	"strconv"

	"github.com/lean-docs/lean/pkg/ir"
)

type xmlStyles struct {
	Defaults *xmlDocumentDefaults `xml:"docDefaults"`
	Styles   []xmlStyle           `xml:"style"`
}

type xmlDocumentDefaults struct {
	Run *xmlRunDefault `xml:"rPrDefault"`
}

type xmlRunDefault struct {
	Properties *xmlRunProps `xml:"rPr"`
}

type xmlStyle struct {
	Type      string        `xml:"type,attr"`
	ID        string        `xml:"styleId,attr"`
	Name      *xmlValAttr   `xml:"name"`
	BasedOn   *xmlValAttr   `xml:"basedOn"`
	Paragraph *xmlParaProps `xml:"pPr"`
	Run       *xmlRunProps  `xml:"rPr"`
}

func parseStyles(content []byte, document *ir.Document) error {
	var source xmlStyles
	if err := xml.Unmarshal(content, &source); err != nil {
		return err
	}
	if source.Defaults != nil && source.Defaults.Run != nil && source.Defaults.Run.Properties != nil {
		applyRunProps(&document.Styles.Defaults, source.Defaults.Run.Properties)
	}
	if document.Styles.Named == nil {
		document.Styles.Named = make(map[string]ir.Style)
	}
	for _, value := range source.Styles {
		style := ir.Style{ID: value.ID, IsChar: value.Type == "character"}
		if value.Name != nil {
			style.Name = value.Name.Val
		}
		if value.BasedOn != nil {
			style.BasedOn = value.BasedOn.Val
		}
		if value.Paragraph != nil {
			applyParaProps(&style.ParaAttrs, value.Paragraph)
		}
		if value.Run != nil {
			applyRunProps(&style.RunAttrs, value.Run)
		}
		document.Styles.Named[value.ID] = style
	}
	return nil
}

func parseStyleNumbering(content []byte) (map[string]*ir.NumberingRef, error) {
	var source xmlStyles
	if err := xml.Unmarshal(content, &source); err != nil {
		return nil, err
	}
	references := make(map[string]*ir.NumberingRef)
	for _, style := range source.Styles {
		if style.Paragraph == nil || style.Paragraph.Numbering == nil || style.Paragraph.Numbering.ID == nil {
			continue
		}
		level := 0
		if style.Paragraph.Numbering.Level != nil {
			level, _ = strconv.Atoi(style.Paragraph.Numbering.Level.Val)
		}
		references[style.ID] = &ir.NumberingRef{ID: style.Paragraph.Numbering.ID.Val, Level: level}
	}
	return references, nil
}
