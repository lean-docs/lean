package ooxml

import (
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

func writeShapeRun(output *strings.Builder, shape *ir.Shape, state *packageState) {
	anchor := shape.Anchor
	if anchor == nil {
		anchor = &ir.DrawingAnchor{}
	}
	output.WriteString(`<w:r><w:drawing>`)
	tag := "inline"
	if anchor.Floating {
		tag = "anchor"
		output.WriteString(`<wp:anchor simplePos="0" relativeHeight="` + strconv.FormatUint(uint64(anchor.ZOrder), 10) + `" behindDoc="` + drawingFlag(anchor.Behind) + `" locked="0" layoutInCell="` + drawingFlag(anchor.LayoutInCell) + `" allowOverlap="` + drawingFlag(anchor.AllowOverlap) + `" distT="` + drawingEMU(anchor.Distance.Top) + `" distB="` + drawingEMU(anchor.Distance.Bottom) + `" distL="` + drawingEMU(anchor.Distance.Left) + `" distR="` + drawingEMU(anchor.Distance.Right) + `"><wp:simplePos x="0" y="0"/>`)
		writeDrawingPosition(output, "H", anchor.Horizontal, "column")
		writeDrawingPosition(output, "V", anchor.Vertical, "paragraph")
	} else {
		output.WriteString(`<wp:inline>`)
	}
	output.WriteString(`<wp:extent cx="` + drawingEMU(shape.Width) + `" cy="` + drawingEMU(shape.Height) + `"/>`)
	if anchor.Floating {
		writeDrawingWrap(output, anchor)
	}
	id := 1
	if state != nil {
		id = state.drawingIDs[shape]
	}
	output.WriteString(`<wp:docPr id="` + strconv.Itoa(id) + `" name="` + escape(shape.ID) + `"/><a:graphic><a:graphicData uri="http://schemas.microsoft.com/office/word/2010/wordprocessingShape"><wps:wsp><wps:cNvSpPr txBox="` + drawingFlag(shape.TextBox) + `"/><wps:spPr><a:xfrm rot="` + strconv.FormatInt(int64(anchor.Rotation*60000), 10) + `" flipH="` + drawingFlag(anchor.FlipHorizontal) + `" flipV="` + drawingFlag(anchor.FlipVertical) + `"><a:off x="0" y="0"/><a:ext cx="` + drawingEMU(shape.Width) + `" cy="` + drawingEMU(shape.Height) + `"/></a:xfrm><a:prstGeom prst="` + escape(shape.Geometry) + `"><a:avLst/></a:prstGeom>`)
	if shape.Fill != nil {
		output.WriteString(`<a:solidFill><a:srgbClr val="` + color(*shape.Fill) + `"/></a:solidFill>`)
	} else {
		output.WriteString(`<a:noFill/>`)
	}
	if shape.Stroke != nil {
		output.WriteString(`<a:ln w="` + drawingEMU(shape.Stroke.Width) + `"><a:solidFill><a:srgbClr val="` + color(shape.Stroke.Color) + `"/></a:solidFill></a:ln>`)
	} else {
		output.WriteString(`<a:ln><a:noFill/></a:ln>`)
	}
	output.WriteString(`</wps:spPr>`)
	if shape.TextBox || len(shape.Blocks) > 0 {
		output.WriteString(`<wps:txbx><w:txbxContent>`)
		imageIndex := 0
		for _, block := range shape.Blocks {
			writeBlock(output, block, &imageIndex, state)
		}
		if len(shape.Blocks) == 0 {
			output.WriteString(`<w:p/>`)
		}
		output.WriteString(`</w:txbxContent></wps:txbx>`)
	}
	output.WriteString(`<wps:bodyPr lIns="` + drawingEMU(shape.Padding.Left) + `" tIns="` + drawingEMU(shape.Padding.Top) + `" rIns="` + drawingEMU(shape.Padding.Right) + `" bIns="` + drawingEMU(shape.Padding.Bottom) + `"`)
	if shape.VerticalAlignment != "" {
		output.WriteString(` anchor="` + escape(shape.VerticalAlignment) + `"`)
	}
	output.WriteString(`>`)
	if shape.AutoFit {
		output.WriteString(`<a:spAutoFit/>`)
	}
	output.WriteString(`</wps:bodyPr></wps:wsp></a:graphicData></a:graphic></wp:` + tag + `></w:drawing></w:r>`)
}
