package lean

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"slices"
	"time"

	"github.com/lean-docs/lean/pkg/ir"
)

func inspectRichContent(document *ir.Document, report *FidelityReport) {
	for _, style := range document.Styles.Named {
		if style.Table == nil {
			continue
		}
		inspectRunValues(style.Table.Cell.RunAttrs, report)
		for conditionType, region := range style.Table.Conditions {
			if !slices.Contains([]string{"wholeTable", "firstRow", "lastRow", "firstCol", "lastCol", "band1Horz", "band2Horz", "band1Vert", "band2Vert", "nwCell", "neCell", "swCell", "seCell"}, conditionType) {
				report.addUnsupported("table-style-region")
			}
			inspectRunValues(region.RunAttrs, report)
		}
	}
	inspect := func(blocks []ir.Block) {
		var stack []bool
		var walk func([]ir.Block)
		walk = func(blocks []ir.Block) {
			for _, block := range blocks {
				switch value := block.(type) {
				case *ir.Paragraph:
					for _, run := range value.Runs {
						if run.Change != nil {
							change := run.Change
							if _, err := time.Parse(time.RFC3339, change.Date); err != nil || change.ID < 0 || change.Author == "" || !slices.Contains([]string{"insert", "delete", "format"}, change.Type) || run.Field != nil || run.Watermark != nil || run.Hyperlink != nil || run.Break != ir.BreakNone {
								report.addUnsupported("tracked-change-metadata")
							}
						}
						if run.Change != nil && run.Change.Type == "format" {
							if run.Change.Previous == nil {
								report.addUnsupported("tracked-change-metadata")
							} else {
								inspectRunValues(*run.Change.Previous, report)
							}
						}
						if run.LegacyDrawing {
							report.addUnsupported("legacy-drawings")
						}
						if mark := run.Watermark; mark != nil {
							if mark.Text == "" && mark.Image == nil || !finite(mark.Width) || !finite(mark.Height) || mark.Width <= 0 || mark.Height <= 0 || !finite(mark.Rotation) || !finite(mark.Opacity) || mark.Opacity < 0 || mark.Opacity > 1 {
								report.addUnsupported("watermark-geometry")
							}
						}
						if mark := run.Watermark; mark != nil && (!finite(mark.Left) || !finite(mark.Top) || !slices.Contains([]string{"", "left", "center", "right", "absolute"}, mark.HorizontalAlign) || !slices.Contains([]string{"", "top", "center", "bottom", "absolute"}, mark.VerticalAlign) || !slices.Contains([]string{"", "page", "margin"}, mark.HorizontalRelative) || !slices.Contains([]string{"", "page", "margin"}, mark.VerticalRelative)) {
							report.addUnsupported("watermark-position")
						}
						if run.Watermark != nil && run.Watermark.Image != nil {
							walkDocumentBlocks([]ir.Block{run.Watermark.Image}, report)
							media := run.Watermark.Image
							dimensions, format, err := image.DecodeConfig(bytes.NewReader(media.Data))
							if err != nil || dimensions.Width <= 0 || dimensions.Height <= 0 || !((format == "png" && media.Format == ir.ImagePNG) || (format == "jpeg" && media.Format == ir.ImageJPEG) || (format == "gif" && media.Format == ir.ImageGIF)) {
								report.addUnsupported("watermark-image-format")
							}
						}
						if run.Style != "" {
							style, exists := document.Styles.Named[run.Style]
							if !exists || !style.IsChar {
								report.addUnsupported("character-style-references")
							}
						}
						if run.Field == nil {
							continue
						}
						if run.Text != "" || run.Break != ir.BreakNone {
							report.addUnsupported("field-structure")
						}
						switch run.Field.Type {
						case "begin":
							stack = append(stack, false)
						case "instruction":
							if len(stack) == 0 || stack[len(stack)-1] {
								report.addUnsupported("field-structure")
							}
						case "separate":
							if len(stack) == 0 || stack[len(stack)-1] {
								report.addUnsupported("field-structure")
							} else {
								stack[len(stack)-1] = true
							}
						case "end":
							if len(stack) == 0 {
								report.addUnsupported("field-structure")
							} else {
								stack = stack[:len(stack)-1]
							}
						default:
							report.addUnsupported("field-structure")
						}
					}
					for _, note := range value.Footnotes {
						walk(note.Blocks)
					}
				case *ir.Table:
					if value.Style != "" && value.Style != "TableNormal" {
						style, exists := document.Styles.Named[value.Style]
						if !exists || style.Table == nil {
							report.addUnsupported("table-styles")
						}
					}
					for _, row := range value.Rows {
						if change := row.Change; change != nil {
							if _, err := time.Parse(time.RFC3339, change.Date); err != nil || change.ID < 0 || change.Author == "" || !slices.Contains([]string{"insert", "delete"}, change.Type) {
								report.addUnsupported("tracked-row-metadata")
							}
						}
						for _, cell := range row.Cells {
							walk(cell.Blocks)
						}
					}
				case *ir.Shape:
					walk(value.Blocks)
				}
			}
		}
		walk(blocks)
		if len(stack) != 0 {
			report.addUnsupported("field-structure")
		}
	}
	for _, section := range document.Sections {
		inspect(section.Blocks)
		for _, group := range [][]*ir.HeaderFooter{append([]*ir.HeaderFooter{section.Header}, section.Headers...), append([]*ir.HeaderFooter{section.Footer}, section.Footers...)} {
			for _, part := range group {
				if part != nil {
					inspect(part.Blocks)
				}
			}
		}
	}
}
