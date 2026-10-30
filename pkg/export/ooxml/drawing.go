package ooxml

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

func drawingEMU(value float64) string { return strconv.FormatInt(int64(value*12700), 10) }
func drawingFlag(value bool) string {
	if value {
		return "1"
	}
	return "0"
}
func writeDrawingPosition(output *strings.Builder, axis string, position ir.DrawingPosition, fallback string) {
	relative := position.Relative
	if relative == "" {
		relative = fallback
	}
	output.WriteString(`<wp:position` + axis + ` relativeFrom="` + escape(relative) + `">`)
	if position.Align != "" {
		output.WriteString(`<wp:align>` + escape(position.Align) + `</wp:align>`)
	} else {
		output.WriteString(`<wp:posOffset>` + drawingEMU(position.Offset) + `</wp:posOffset>`)
	}
	output.WriteString(`</wp:position` + axis + `>`)
}
func writeDrawingRun(output *strings.Builder, image *ir.Image, index int, states ...*packageState) {
	drawingID := index
	if len(states) > 0 && states[0] != nil {
		drawingID = states[0].drawingIDs[image]
	}
	anchor := image.Anchor
	if anchor == nil {
		anchor = &ir.DrawingAnchor{}
		if image.Float != ir.FloatNone {
			anchor.Floating = true
			anchor.Wrap = "square"
			anchor.Horizontal.Align = "left"
			if image.Float == ir.FloatRight {
				anchor.Horizontal.Align = "right"
			}
		}
	}
	extent := `<wp:extent cx="` + drawingEMU(image.Width) + `" cy="` + drawingEMU(image.Height) + `"/>`
	distance := ` distT="` + drawingEMU(anchor.Distance.Top) + `" distB="` + drawingEMU(anchor.Distance.Bottom) + `" distL="` + drawingEMU(anchor.Distance.Left) + `" distR="` + drawingEMU(anchor.Distance.Right) + `"`
	output.WriteString(`<w:r><w:drawing>`)
	if anchor.Floating {
		output.WriteString(`<wp:anchor` + distance + ` simplePos="0" relativeHeight="` + strconv.FormatUint(uint64(anchor.ZOrder), 10) + `" behindDoc="` + drawingFlag(anchor.Behind) + `" locked="0" layoutInCell="` + drawingFlag(anchor.LayoutInCell) + `" allowOverlap="` + drawingFlag(anchor.AllowOverlap) + `"><wp:simplePos x="0" y="0"/>`)
		writeDrawingPosition(output, "H", anchor.Horizontal, "column")
		writeDrawingPosition(output, "V", anchor.Vertical, "paragraph")
		output.WriteString(extent)
		writeDrawingWrap(output, anchor)
	} else {
		output.WriteString(`<wp:inline` + distance + `>` + extent)
	}
	output.WriteString(fmt.Sprintf(`<wp:docPr id="%d" name="Image %d" descr="%s"/><wp:cNvGraphicFramePr><a:graphicFrameLocks noChangeAspect="1"/></wp:cNvGraphicFramePr><a:graphic><a:graphicData uri="http://schemas.openxmlformats.org/drawingml/2006/picture"><pic:pic><pic:nvPicPr><pic:cNvPr id="0" name="Image %d"/><pic:cNvPicPr/></pic:nvPicPr><pic:blipFill><a:blip r:embed="rIdImage%d"/>`, drawingID, drawingID, escape(image.Alt), drawingID, index))
	if anchor.Crop != (ir.Padding{}) {
		percent := func(value float64) string { return strconv.FormatInt(int64(value*100000), 10) }
		output.WriteString(`<a:srcRect l="` + percent(anchor.Crop.Left) + `" t="` + percent(anchor.Crop.Top) + `" r="` + percent(anchor.Crop.Right) + `" b="` + percent(anchor.Crop.Bottom) + `"/>`)
	}
	output.WriteString(`<a:stretch><a:fillRect/></a:stretch></pic:blipFill><pic:spPr><a:xfrm rot="` + strconv.FormatInt(int64(anchor.Rotation*60000), 10) + `" flipH="` + drawingFlag(anchor.FlipHorizontal) + `" flipV="` + drawingFlag(anchor.FlipVertical) + `"><a:off x="0" y="0"/><a:ext cx="` + drawingEMU(image.Width) + `" cy="` + drawingEMU(image.Height) + `"/></a:xfrm><a:prstGeom prst="rect"><a:avLst/></a:prstGeom></pic:spPr></pic:pic></a:graphicData></a:graphic>`)
	if anchor.Floating {
		output.WriteString(`</wp:anchor>`)
	} else {
		output.WriteString(`</wp:inline>`)
	}
	output.WriteString(`</w:drawing></w:r>`)
}

func writeDrawingWrap(output *strings.Builder, anchor *ir.DrawingAnchor) {
	side := anchor.WrapSide
	if side == "" {
		side = "bothSides"
	}
	switch anchor.Wrap {
	case "square":
		output.WriteString(`<wp:wrapSquare wrapText="` + escape(side) + `"/>`)
	case "topAndBottom":
		output.WriteString(`<wp:wrapTopAndBottom/>`)
	case "tight", "through":
		tag := "wrapTight"
		if anchor.Wrap == "through" {
			tag = "wrapThrough"
		}
		polygon := anchor.Polygon
		if polygon == nil {
			polygon = &ir.DrawingPolygon{Points: [][2]int64{{0, 0}, {0, 21600}, {21600, 21600}, {21600, 0}}}
		}
		output.WriteString(`<wp:` + tag + ` wrapText="` + escape(side) + `"><wp:wrapPolygon edited="` + drawingFlag(polygon.Edited) + `">`)
		for index, point := range append(append([][2]int64{}, polygon.Points...), polygon.Points[0]) {
			name := "lineTo"
			if index == 0 {
				name = "start"
			}
			output.WriteString(`<wp:` + name + ` x="` + strconv.FormatInt(point[0], 10) + `" y="` + strconv.FormatInt(point[1], 10) + `"/>`)
		}
		output.WriteString(`</wp:wrapPolygon></wp:` + tag + `>`)
	default:
		output.WriteString(`<wp:wrapNone/>`)
	}
}
