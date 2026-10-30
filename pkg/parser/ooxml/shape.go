package ooxml

import (
	"encoding/xml"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

type xmlWordShape struct {
	NonVisual struct {
		TextBox string `xml:"txBox,attr"`
	} `xml:"cNvSpPr"`
	Properties xmlShapeProperties `xml:"spPr"`
	TextBox    *xmlTextBox        `xml:"txbx"`
	Body       *xmlTextBody       `xml:"bodyPr"`
}
type xmlShapeProperties struct {
	Transform *xmlPictureTransform `xml:"xfrm"`
	Geometry  struct {
		Preset string `xml:"prst,attr"`
	} `xml:"prstGeom"`
	Fill *xmlShapeFill `xml:"solidFill"`
	Line *xmlShapeLine `xml:"ln"`
}
type xmlShapeFill struct {
	RGB *struct {
		Value string `xml:"val,attr"`
	} `xml:"srgbClr"`
}
type xmlShapeLine struct {
	Width  string        `xml:"w,attr"`
	Fill   *xmlShapeFill `xml:"solidFill"`
	NoFill *struct{}     `xml:"noFill"`
}
type xmlTextBox struct {
	Content string `xml:",innerxml"`
}
type xmlTextBody struct {
	Left    string    `xml:"lIns,attr"`
	Right   string    `xml:"rIns,attr"`
	Top     string    `xml:"tIns,attr"`
	Bottom  string    `xml:"bIns,attr"`
	Anchor  string    `xml:"anchor,attr"`
	AutoFit *struct{} `xml:"spAutoFit"`
}

func shapeColor(fill *xmlShapeFill) *ir.Color {
	if fill == nil || fill.RGB == nil {
		return nil
	}
	value, ok := parseHexColor(fill.RGB.Value)
	if !ok {
		return nil
	}
	return &value
}
func convertShape(drawing xmlDrawing, context *parseCtx) *ir.Shape {
	container := drawing.Inline
	if container == nil {
		container = drawing.Anchor
	}
	if container == nil || container.Graphic.Data.Shape == nil {
		return nil
	}
	value := container.Graphic.Data.Shape
	shape := &ir.Shape{ID: context.nextImageID(), Geometry: value.Properties.Geometry.Preset, Anchor: drawingAnchor(drawing, container), TextBox: xmlDrawingBool(value.NonVisual.TextBox), Fill: shapeColor(value.Properties.Fill)}
	if shape.Geometry == "" {
		shape.Geometry = "rect"
	}
	if container.Extent != nil {
		shape.Width = emuAttr(container.Extent.CX)
		shape.Height = emuAttr(container.Extent.CY)
	}
	if value.Properties.Transform != nil {
		transform := value.Properties.Transform
		shape.Anchor.Rotation = rotationValue(transform.Rotation)
		shape.Anchor.FlipHorizontal = xmlDrawingBool(transform.FlipH)
		shape.Anchor.FlipVertical = xmlDrawingBool(transform.FlipV)
	}
	if line := value.Properties.Line; line != nil && line.NoFill == nil {
		shape.Stroke = &ir.Border{Style: "single", Width: emuAttr(line.Width)}
		if color := shapeColor(line.Fill); color != nil {
			shape.Stroke.Color = *color
		}
	}
	if body := value.Body; body != nil {
		shape.Padding = ir.Padding{Left: emuAttr(body.Left), Right: emuAttr(body.Right), Top: emuAttr(body.Top), Bottom: emuAttr(body.Bottom)}
		shape.AutoFit = body.AutoFit != nil
		shape.VerticalAlignment = body.Anchor
	}
	if value.TextBox != nil {
		decoder := xml.NewDecoder(strings.NewReader(`<root xmlns:w="` + wNS + `">` + value.TextBox.Content + `</root>`))
		if advanceToNS(decoder, "", "txbxContent") == nil {
			if parseBlocks(decoder, "txbxContent", &shape.Blocks, context) != nil {
				return nil
			}
		}
	}
	return shape
}
