package ooxml

import (
	"encoding/xml"
	"github.com/lean-docs/lean/pkg/ir"
	"unicode/utf16"
)

func (run *xmlRun) UnmarshalXML(decoder *xml.Decoder, start xml.StartElement) error {
	*run = xmlRun{}
	offset := 0
	for {
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		switch element := token.(type) {
		case xml.EndElement:
			if element.Name == start.Name {
				return nil
			}
		case xml.StartElement:
			switch element.Name.Local {
			case "pict":
				var picture xmlLegacyPicture
				if err := decoder.DecodeElement(&picture, &element); err != nil {
					return err
				}
				picture.Shapes = append(picture.Shapes, picture.Rectangles...)
				run.Watermark, run.LegacyDrawing = picture.watermark()
				if len(picture.Shapes) == 1 && picture.Shapes[0].ImageData != nil {
					run.WatermarkImageID = picture.Shapes[0].ImageData.ID
					run.WatermarkImageAlt = picture.Shapes[0].ImageData.Title
				}
			case "fldChar":
				var marker struct {
					Type   string `xml:"fldCharType,attr"`
					Locked string `xml:"fldLock,attr"`
					Dirty  string `xml:"dirty,attr"`
				}
				if err := decoder.DecodeElement(&marker, &element); err != nil {
					return err
				}
				run.Field = &ir.FieldElement{Type: marker.Type, Locked: marker.Locked == "true" || marker.Locked == "1", Dirty: marker.Dirty == "true" || marker.Dirty == "1"}
			case "instrText":
				var instruction xmlText
				if err := decoder.DecodeElement(&instruction, &element); err != nil {
					return err
				}
				run.Field = &ir.FieldElement{Type: "instruction", Instruction: instruction.Value}
			case "rPr":
				var properties xmlRunProps
				if err := decoder.DecodeElement(&properties, &element); err != nil {
					return err
				}
				run.Props = &properties
			case "t", "delText":
				var text xmlText
				if err := decoder.DecodeElement(&text, &element); err != nil {
					return err
				}
				run.Text = append(run.Text, text)
				offset += len(utf16.Encode([]rune(text.Value)))
			case "tab":
				run.Text = append(run.Text, xmlText{Value: "\t"})
				offset++
				if err := decoder.Skip(); err != nil {
					return err
				}
			case "br":
				var value xmlBreak
				if err := decoder.DecodeElement(&value, &element); err != nil {
					return err
				}
				run.Breaks = append(run.Breaks, value)
				offset++
			case "drawing":
				var drawing xmlDrawing
				if err := decoder.DecodeElement(&drawing, &element); err != nil {
					return err
				}
				run.Drawings = append(run.Drawings, drawing)
				run.DrawingOffsets = append(run.DrawingOffsets, offset)
			case "footnoteReference":
				var reference xmlIDAttr
				if err := decoder.DecodeElement(&reference, &element); err != nil {
					return err
				}
				run.FootnoteReference = &reference
			default:
				if err := decoder.Skip(); err != nil {
					return err
				}
			}
		}
	}
}
