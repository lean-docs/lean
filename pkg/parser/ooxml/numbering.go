package ooxml

import (
	"encoding/xml"
	"strconv"

	"github.com/lean-docs/lean/pkg/ir"
)

type xmlNumbering struct {
	Abstract []xmlAbstractNumbering `xml:"abstractNum"`
	Concrete []xmlConcreteNumbering `xml:"num"`
}

type xmlAbstractNumbering struct {
	ID     string              `xml:"abstractNumId,attr"`
	Levels []xmlNumberingLevel `xml:"lvl"`
}

type xmlConcreteNumbering struct {
	ID       string      `xml:"numId,attr"`
	Abstract *xmlValAttr `xml:"abstractNumId"`
}

type xmlNumberingLevel struct {
	Level      string        `xml:"ilvl,attr"`
	Start      *xmlValAttr   `xml:"start"`
	Format     *xmlValAttr   `xml:"numFmt"`
	Text       *xmlValAttr   `xml:"lvlText"`
	Alignment  *xmlValAttr   `xml:"lvlJc"`
	Properties *xmlParaProps `xml:"pPr"`
	Run        *xmlRunProps  `xml:"rPr"`
}

func parseNumbering(content []byte, document *ir.Document) error {
	var numbering xmlNumbering
	if err := xml.Unmarshal(content, &numbering); err != nil {
		return err
	}
	abstract := make(map[string][]ir.NumberingLevel, len(numbering.Abstract))
	for _, definition := range numbering.Abstract {
		levels := make([]ir.NumberingLevel, 0, len(definition.Levels))
		for _, source := range definition.Levels {
			level, _ := strconv.Atoi(source.Level)
			converted := ir.NumberingLevel{Level: level, Start: 1}
			if source.Start != nil {
				converted.Start, _ = strconv.Atoi(source.Start.Val)
			}
			if source.Format != nil {
				converted.Format = numberingFormat(source.Format.Val)
			}
			if source.Text != nil {
				converted.Text = source.Text.Val
			}
			if source.Alignment != nil {
				converted.Align = alignFromVal(source.Alignment.Val)
			}
			if source.Properties != nil {
				if source.Properties.Indent != nil {
					converted.Indent.Left = twipAttr(source.Properties.Indent.Left)
					converted.Indent.Right = twipAttr(source.Properties.Indent.Right)
					converted.Indent.FirstLine = twipAttr(source.Properties.Indent.FirstLine)
					converted.Indent.Hanging = twipAttr(source.Properties.Indent.Hanging)
				}
			}
			if source.Run != nil {
				applyRunProps(&converted.RunAttrs, source.Run)
			}
			levels = append(levels, converted)
		}
		abstract[definition.ID] = levels
	}
	for _, concrete := range numbering.Concrete {
		if concrete.Abstract == nil {
			continue
		}
		levels := append([]ir.NumberingLevel(nil), abstract[concrete.Abstract.Val]...)
		document.Styles.Numbering = append(document.Styles.Numbering, ir.NumberingDef{ID: concrete.ID, Levels: levels})
	}
	return nil
}

func numberingFormat(value string) ir.NumberFormat {
	switch value {
	case "decimal":
		return ir.NumFormatDecimal
	case "lowerLetter":
		return ir.NumFormatLowerAlpha
	case "upperLetter":
		return ir.NumFormatUpperAlpha
	case "lowerRoman":
		return ir.NumFormatLowerRoman
	case "upperRoman":
		return ir.NumFormatUpperRoman
	default:
		return ir.NumFormatBullet
	}
}
