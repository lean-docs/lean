package lean

import (
	"bytes"
	"encoding/xml"
)

func inspectDrawingFeatures(content []byte, report *FidelityReport) {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	drawingDepth := 0
	for {
		token, err := decoder.Token()
		if err != nil {
			return
		}
		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "drawing" {
				drawingDepth++
			}
			if drawingDepth == 0 {
				continue
			}
			if element.Name.Local == "graphicData" {
				for _, attribute := range element.Attr {
					if attribute.Name.Local == "uri" && attribute.Value != "http://schemas.openxmlformats.org/drawingml/2006/picture" && attribute.Value != "http://schemas.microsoft.com/office/word/2010/wordprocessingShape" {
						report.addUnsupported("drawing-content")
					}
				}
			}
			switch element.Name.Local {
			case "gradFill", "custGeom", "outerShdw", "innerShdw", "glow", "reflection", "softEdge", "scene3d", "sp3d":
				report.addUnsupported("drawing-effects")
			}
		case xml.EndElement:
			if element.Name.Local == "drawing" {
				drawingDepth--
			}
		}
	}
}
