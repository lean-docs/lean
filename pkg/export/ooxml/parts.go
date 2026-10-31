package ooxml

import (
	"strconv"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

type auxiliaryPart struct {
	section                                   int
	partType, element, filename, relationship string
	value                                     *ir.HeaderFooter
}

func (state *packageState) collectBlocks(blocks []ir.Block) {
	walkBlocks(blocks, func(block ir.Block) {
		if block.BlockType() == ir.BlockImage || block.BlockType() == ir.BlockShape {
			if state.drawingIDs[block] == 0 {
				state.drawingIDs[block] = len(state.drawingIDs) + 1
			}
		}
		switch value := block.(type) {
		case *ir.Image:
			if state.imageIDs[value] == 0 {
				state.images = append(state.images, value)
				state.imageIDs[value] = len(state.images)
			}
		case *ir.Paragraph:
			for _, note := range value.Footnotes {
				state.notes = append(state.notes, note)
				state.footnoteIDs[value] = append(state.footnoteIDs[value], len(state.notes))
			}
			for _, run := range value.Runs {
				if run.Watermark != nil && run.Watermark.Image != nil && state.imageIDs[run.Watermark.Image] == 0 {
					state.images = append(state.images, run.Watermark.Image)
					state.imageIDs[run.Watermark.Image] = len(state.images)
				}
				if run.Hyperlink != nil && run.Hyperlink.URL != "" && state.hyperlinkIDs[run.Hyperlink.URL] == 0 {
					state.hyperlinks = append(state.hyperlinks, run.Hyperlink.URL)
					state.hyperlinkIDs[run.Hyperlink.URL] = len(state.hyperlinks)
				}
			}
		}
	})
}

func (state *packageState) addAuxiliary(section int, element string, primary *ir.HeaderFooter, others []*ir.HeaderFooter) {
	values := append([]*ir.HeaderFooter{primary}, others...)
	for _, value := range values {
		if value == nil {
			continue
		}
		partType := "header"
		if element == "ftr" {
			partType = "footer"
		}
		index := 0
		for _, part := range state.auxiliary {
			if part.partType == partType {
				index++
			}
		}
		id := strconv.Itoa(index + 1)
		state.auxiliary = append(state.auxiliary, auxiliaryPart{section: section, partType: partType, element: element, filename: partType + id + ".xml", relationship: "rId" + map[string]string{"header": "Header", "footer": "Footer"}[partType] + id, value: value})
		state.collectBlocks(value.Blocks)
	}
}
func (state *packageState) auxiliaryReferences(section int) string {
	var content strings.Builder
	for _, part := range state.auxiliary {
		if part.section != section {
			continue
		}
		partType := []string{"default", "first", "even"}[part.value.Type]
		content.WriteString(`<w:` + part.partType + `Reference w:type="` + partType + `" r:id="` + part.relationship + `"/>`)
	}
	return content.String()
}
func (state *packageState) hasEvenVariant() bool {
	if state.doc.EvenAndOddHeaders != nil {
		return *state.doc.EvenAndOddHeaders
	}
	for _, part := range state.auxiliary {
		if part.value.Type == ir.HeaderFooterEven {
			return true
		}
	}
	return false
}
