package lean

import (
	"math"
	"slices"
	"unicode/utf16"

	"github.com/lean-docs/lean/pkg/ir"
)

func inspectDocumentStructure(document *ir.Document, report *FidelityReport) {
	for id := range document.Styles.Named {
		if _, ok := document.Styles.ResolveStyle(id); !ok {
			report.addUnsupported("circular-styles")
		}
	}
	definitions := map[string]ir.NumberingDef{}
	for _, definition := range document.Styles.Numbering {
		if definition.ID == "" || definitions[definition.ID].ID != "" {
			report.addUnsupported("invalid-numbering")
		}
		seen := map[int]bool{}
		for _, level := range definition.Levels {
			if level.Level < 0 || level.Level > 8 || seen[level.Level] || level.Format < ir.NumFormatBullet || level.Format > ir.NumFormatUpperRoman || level.Start < 0 {
				report.addUnsupported("invalid-numbering")
			}
			seen[level.Level] = true
			inspectRunValues(level.RunAttrs, report)
		}
		definitions[definition.ID] = definition
	}
	validate := func(block ir.Block) {
		if table, ok := block.(*ir.Table); ok {
			if len(table.Rows) == 0 || table.Align < ir.AlignLeft || table.Align > ir.AlignRight {
				report.addUnsupported("invalid-table")
			}
			for _, width := range table.ColumnWidths {
				if !finite(width) || width < 0 {
					report.addUnsupported("invalid-table-measurements")
				}
			}
			columns := -1
			for _, row := range table.Rows {
				count := 0
				if len(row.Cells) == 0 || row.Height != nil && (!finite(*row.Height) || *row.Height < 0) {
					report.addUnsupported("invalid-table")
				}
				for _, cell := range row.Cells {
					if cell.ColSpan < 0 || cell.RowSpan < 0 || cell.VAlign < ir.VAlignTop || cell.VAlign > ir.VAlignBottom {
						report.addUnsupported("invalid-table")
					}
					count += max(1, cell.ColSpan)
					for _, value := range []float64{cell.Width, cell.Padding.Top, cell.Padding.Bottom, cell.Padding.Left, cell.Padding.Right} {
						if !finite(value) || value < 0 {
							report.addUnsupported("invalid-table-measurements")
						}
					}
				}
				if columns >= 0 && columns != count {
					report.addUnsupported("irregular-table-grid")
				}
				columns = count
			}
		}
		paragraph, ok := block.(*ir.Paragraph)
		if !ok {
			return
		}
		var units []uint16
		for _, run := range paragraph.Runs {
			units = append(units, utf16.Encode([]rune(run.Text))...)
			if run.Break != ir.BreakNone {
				units = append(units, '\n')
			}
		}
		for _, note := range paragraph.Footnotes {
			if note.Offset != nil {
				position := *note.Offset
				if position < 0 || position > len(units) || position > 0 && position < len(units) && units[position-1] >= 0xD800 && units[position-1] <= 0xDBFF {
					report.addUnsupported("footnote-anchors")
				}
			}
		}
		attrs := paragraph.Para
		if attrs.Align < ir.AlignLeft || attrs.Align > ir.AlignJustify || attrs.Spacing.LineRule < ir.LineRuleAuto || attrs.Spacing.LineRule > ir.LineRuleAtLeast || attrs.OutlineLevel < 0 || attrs.OutlineLevel > 9 {
			report.addUnsupported("invalid-paragraph-formatting")
		}
		for _, value := range []float64{attrs.Spacing.Before, attrs.Spacing.After, attrs.Spacing.Line, attrs.Indent.Left, attrs.Indent.Right, attrs.Indent.FirstLine, attrs.Indent.Hanging} {
			if math.IsNaN(value) || math.IsInf(value, 0) {
				report.addUnsupported("invalid-paragraph-measurements")
			}
		}
		if reference := paragraph.Numbering; reference != nil {
			definition := definitions[reference.ID]
			found := false
			for _, level := range definition.Levels {
				if level.Level == reference.Level {
					found = true
				}
			}
			if !found || reference.Restart != nil && *reference.Restart < 0 {
				report.addUnsupported("invalid-numbering")
			}
		}
	}
	var body []ir.Block
	for _, section := range document.Sections {
		if !slices.Contains([]string{"", "nextPage", "continuous", "oddPage", "evenPage", "nextColumn"}, section.Break) {
			report.addUnsupported("section-breaks")
		}
		for _, distance := range []*float64{section.Properties.HeaderDistance, section.Properties.FooterDistance, &section.Properties.Gutter} {
			if distance != nil && (!finite(*distance) || *distance < 0) {
				report.addUnsupported("invalid-page-layout")
			}
		}
		body = append(body, section.Blocks...)
		for _, value := range []float64{section.Properties.Width, section.Properties.Height, section.Properties.MarginTop, section.Properties.MarginBottom, section.Properties.MarginLeft, section.Properties.MarginRight} {
			if math.IsNaN(value) || math.IsInf(value, 0) || value < 0 {
				report.addUnsupported("invalid-page-layout")
			}
		}
		for _, part := range append(append([]*ir.HeaderFooter{section.Header, section.Footer}, section.Headers...), section.Footers...) {
			if part != nil {
				inspectBookmarkRanges(part.Blocks, report)
				walkStructuralBlocks(part.Blocks, validate)
			}
		}
	}
	inspectBookmarkRanges(body, report)
	walkStructuralBlocks(body, validate)
}

func walkStructuralBlocks(blocks []ir.Block, visit func(ir.Block)) {
	for _, block := range blocks {
		if block == nil {
			continue
		}
		visit(block)
		switch value := block.(type) {
		case *ir.Shape:
			walkStructuralBlocks(value.Blocks, visit)
		case *ir.Table:
			for _, row := range value.Rows {
				for _, cell := range row.Cells {
					walkStructuralBlocks(cell.Blocks, visit)
				}
			}
		case *ir.Paragraph:
			for _, note := range value.Footnotes {
				walkStructuralBlocks(note.Blocks, visit)
			}
		}
	}
}

func inspectBookmarkRanges(blocks []ir.Block, report *FidelityReport) {
	type position struct {
		order int
		units []uint16
	}
	paragraphs := map[string]position{}
	ambiguous := map[string]bool{}
	order := 0
	walkStructuralBlocks(blocks, func(block ir.Block) {
		if value, ok := block.(*ir.Paragraph); ok {
			var units []uint16
			for _, run := range value.Runs {
				units = append(units, utf16.Encode([]rune(run.Text))...)
				if run.Break != ir.BreakNone {
					units = append(units, '\n')
				}
			}
			if _, exists := paragraphs[value.ID]; exists {
				ambiguous[value.ID] = true
			}
			paragraphs[value.ID] = position{order: order, units: units}
			order++
		}
	})
	valid := func(anchor *ir.BookmarkPosition) bool {
		if anchor == nil {
			return false
		}
		paragraph, ok := paragraphs[anchor.ParagraphID]
		if !ok || ambiguous[anchor.ParagraphID] || anchor.Offset < 0 || anchor.Offset > len(paragraph.units) {
			return false
		}
		return anchor.Offset == 0 || anchor.Offset == len(paragraph.units) || paragraph.units[anchor.Offset-1] < 0xD800 || paragraph.units[anchor.Offset-1] > 0xDBFF
	}
	walkStructuralBlocks(blocks, func(block ir.Block) {
		var anchor *ir.DrawingAnchor
		switch value := block.(type) {
		case *ir.Image:
			anchor = value.Anchor
		case *ir.Shape:
			anchor = value.Anchor
		}
		if anchor != nil {
			if !slices.Contains([]string{"", "none", "square", "topAndBottom", "tight", "through"}, anchor.Wrap) {
				report.addUnsupported("drawing-wrap")
			}
			if anchor.Polygon != nil && (len(anchor.Polygon.Points) < 3 || len(anchor.Polygon.Points) > 4096) {
				report.addUnsupported("drawing-polygon")
			}

			if anchor.ParagraphID != "" && !valid(&ir.BookmarkPosition{ParagraphID: anchor.ParagraphID, Offset: anchor.Offset}) {
				report.addUnsupported("drawing-anchors")
			}
			for _, number := range []float64{anchor.Horizontal.Offset, anchor.Vertical.Offset, anchor.Rotation, anchor.Distance.Top, anchor.Distance.Bottom, anchor.Distance.Left, anchor.Distance.Right, anchor.Crop.Left, anchor.Crop.Right, anchor.Crop.Top, anchor.Crop.Bottom} {
				if !finite(number) {
					report.addUnsupported("drawing-measurements")
				}
			}
			if anchor.Crop.Left+anchor.Crop.Right >= 1 || anchor.Crop.Top+anchor.Crop.Bottom >= 1 {
				report.addUnsupported("drawing-crop")
			}
		}
		bookmark, ok := block.(*ir.Bookmark)
		if !ok || bookmark.Start == nil {
			return
		}
		if !valid(bookmark.Start) || !valid(bookmark.End) {
			report.addUnsupported("bookmark-ranges")
			return
		}
		a, b := paragraphs[bookmark.Start.ParagraphID], paragraphs[bookmark.End.ParagraphID]
		if a.order > b.order || a.order == b.order && bookmark.Start.Offset > bookmark.End.Offset {
			report.addUnsupported("bookmark-ranges")
		}
	})
}
