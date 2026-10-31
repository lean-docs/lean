package ooxml

import (
	"github.com/lean-docs/lean/pkg/ir"
	"strconv"
)

func drawingAnchor(drawing xmlDrawing, value *xmlDrawingContainer) *ir.DrawingAnchor {
	anchor := &ir.DrawingAnchor{Floating: drawing.Anchor != nil, Behind: xmlDrawingBool(value.Behind), AllowOverlap: xmlDrawingBool(value.Overlap), LayoutInCell: xmlDrawingBool(value.InCell), ZOrder: value.ZOrder, Distance: ir.Padding{Top: emuAttr(value.DistanceTop), Bottom: emuAttr(value.DistanceBottom), Left: emuAttr(value.DistanceLeft), Right: emuAttr(value.DistanceRight)}}
	if value.PositionH != nil {
		anchor.Horizontal = ir.DrawingPosition{Relative: value.PositionH.Relative, Offset: emuAttr(value.PositionH.Offset), Align: value.PositionH.Align}
	}
	if value.PositionV != nil {
		anchor.Vertical = ir.DrawingPosition{Relative: value.PositionV.Relative, Offset: emuAttr(value.PositionV.Offset), Align: value.PositionV.Align}
	}
	switch {
	case value.WrapSquare != nil:
		anchor.Wrap = "square"
		anchor.WrapSide = value.WrapSquare.Text
	case value.WrapTight != nil:
		anchor.Wrap = "tight"
		anchor.WrapSide = value.WrapTight.Text
	case value.WrapThrough != nil:
		anchor.Wrap = "through"
		anchor.WrapSide = value.WrapThrough.Text
	case value.WrapTopBottom != nil:
		anchor.Wrap = "topAndBottom"
	default:
		anchor.Wrap = "none"
	}
	var wrap *xmlDrawingWrap
	if value.WrapTight != nil {
		wrap = value.WrapTight
	} else if value.WrapThrough != nil {
		wrap = value.WrapThrough
	}
	if wrap != nil && wrap.Polygon != nil {
		polygon := wrap.Polygon
		points := [][2]int64{{polygon.Start.X, polygon.Start.Y}}
		for _, point := range polygon.Lines {
			points = append(points, [2]int64{point.X, point.Y})
		}
		if len(points) > 3 && points[0] == points[len(points)-1] {
			points = points[:len(points)-1]
		}
		anchor.Polygon = &ir.DrawingPolygon{Edited: xmlDrawingBool(polygon.Edited), Points: points}
	}
	picture := value.Graphic.Data.Picture
	if props := picture.Properties; props != nil && props.Transform != nil {
		t := props.Transform
		angle, _ := strconv.ParseFloat(t.Rotation, 64)
		anchor.Rotation = angle / 60000
		anchor.FlipHorizontal = xmlDrawingBool(t.FlipH)
		anchor.FlipVertical = xmlDrawingBool(t.FlipV)
	}
	if crop := picture.BlipFill.Crop; crop != nil {
		percentage := func(value string) float64 { number, _ := strconv.ParseFloat(value, 64); return number / 100000 }
		anchor.Crop = ir.Padding{Left: percentage(crop.Left), Right: percentage(crop.Right), Top: percentage(crop.Top), Bottom: percentage(crop.Bottom)}
	}
	return anchor
}
func xmlDrawingBool(value string) bool { return value == "1" || value == "true" || value == "on" }

func rotationValue(value string) float64 {
	number, _ := strconv.ParseFloat(value, 64)
	return number / 60000
}
