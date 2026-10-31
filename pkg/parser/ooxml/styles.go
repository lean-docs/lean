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
	Type       string                `xml:"type,attr"`
	ID         string                `xml:"styleId,attr"`
	Name       *xmlValAttr           `xml:"name"`
	BasedOn    *xmlValAttr           `xml:"basedOn"`
	Paragraph  *xmlParaProps         `xml:"pPr"`
	Run        *xmlRunProps          `xml:"rPr"`
	Table      *xmlTblProps          `xml:"tblPr"`
	Cell       *xmlCellProps         `xml:"tcPr"`
	Conditions []xmlTableStyleRegion `xml:"tblStylePr"`
}

type xmlTableStyleRegion struct {
	Type      string        `xml:"type,attr"`
	Run       *xmlRunProps  `xml:"rPr"`
	Paragraph *xmlParaProps `xml:"pPr"`
	Cell      *xmlCellProps `xml:"tcPr"`
}

func tableStyleRegion(run *xmlRunProps, paragraph *xmlParaProps, cell *xmlCellProps) ir.TableCellStyle {
	var value ir.TableCellStyle
	if run != nil {
		applyRunProps(&value.RunAttrs, run)
	}
	if paragraph != nil {
		applyParaProps(&value.ParaAttrs, paragraph)
	}
	if cell != nil {
		var target ir.TableCell
		applyCellProps(&target, cell)
		value.Borders, value.Shading, value.Padding = target.Borders, target.Shading, target.Padding
	}
	return value
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
		if value.Type == "table" {
			style.Table = &ir.TableStyle{Cell: tableStyleRegion(nil, nil, value.Cell), Conditions: make(map[string]ir.TableCellStyle)}
			if value.Table != nil {
				var table ir.Table
				applyTableProps(&table, value.Table)
				style.Table.Borders = table.Borders
				if value.Table.RowBandSize != nil {
					style.Table.RowBandSize, _ = strconv.Atoi(value.Table.RowBandSize.Val)
				}
				if value.Table.ColumnBandSize != nil {
					style.Table.ColumnBandSize, _ = strconv.Atoi(value.Table.ColumnBandSize.Val)
				}
			}
			for _, region := range value.Conditions {
				style.Table.Conditions[region.Type] = tableStyleRegion(region.Run, region.Paragraph, region.Cell)
			}
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
