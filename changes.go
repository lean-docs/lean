package lean

import (
	"encoding/xml"
	"errors"
	"io"
	"strings"
	"unicode/utf16"

	"github.com/lean-docs/lean/pkg/ir"
)

func inspectTrackedText(source string, report *FidelityReport) {
	decoder := xml.NewDecoder(strings.NewReader("<root>" + source + "</root>"))
	var stack []string
	tracked := 0
	formatDepth := 0
	rowChanges := 0
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			return
		}
		if err != nil {
			report.addUnsupported("tracked-change-structure")
			return
		}
		switch element := token.(type) {
		case xml.StartElement:
			name := element.Name.Local
			if name == "trPr" {
				rowChanges = 0
			}
			if name == "rPrChange" {
				formatDepth++
				if formatDepth > 1 {
					report.addUnsupported("tracked-change-structure")
				}
			}
			parent := ""
			if len(stack) > 0 {
				parent = stack[len(stack)-1]
			}
			switch name {
			case "cellIns", "cellDel", "cellMerge", "tblGridChange", "numberingChange", "conflictIns", "conflictDel":
				report.addUnsupported("tracked-change-structure")
			}
			if strings.HasPrefix(name, "customXml") && strings.Contains(name, "Range") {
				report.addUnsupported("tracked-change-structure")
			}
			if tracked > 0 && parent == "r" && name != "rPr" && name != "t" && name != "delText" && name != "tab" {
				report.addUnsupported("tracked-change-structure")
			}
			if (tracked > 0 || formatDepth > 0) && parent == "rPr" {
				switch name {
				case "rPrChange", "rStyle", "b", "i", "u", "strike", "smallCaps", "caps", "vertAlign", "sz", "rFonts", "color", "highlight", "spacing", "lang", "shd":
				default:
					report.addUnsupported("tracked-change-structure")
				}
			}
			if name == "ins" || name == "del" {
				if (parent != "p" && parent != "trPr") || tracked > 0 {
					report.addUnsupported("tracked-change-structure")
				}
				if parent == "trPr" {
					rowChanges++
					if rowChanges > 1 {
						report.addUnsupported("tracked-change-structure")
					}
				}
				tracked++
			} else if (strings.HasSuffix(name, "PrChange") && name != "rPrChange") || strings.HasPrefix(name, "moveFrom") || strings.HasPrefix(name, "moveTo") {
				report.addUnsupported("tracked-change-structure")
			}
			if name == "rPrChange" && (parent != "rPr" || tracked > 0) {
				report.addUnsupported("tracked-change-structure")
			}
			if tracked > 0 && name != "ins" && name != "del" {
				if parent == "ins" || parent == "del" {
					if name != "r" {
						report.addUnsupported("tracked-change-structure")
					}
				}
				switch name {
				case "drawing", "pict", "fldChar", "instrText", "br", "hyperlink":
					report.addUnsupported("tracked-change-structure")
				}
			}
			stack = append(stack, name)
		case xml.EndElement:
			if element.Name.Local == "rPrChange" {
				formatDepth--
			}
			if element.Name.Local == "ins" || element.Name.Local == "del" {
				tracked--
			}
			if len(stack) > 0 {
				stack = stack[:len(stack)-1]
			}
		}
	}
}

func visitTextChanges(document *ir.Document, visit func([]ir.Block, *ir.Paragraph)) {
	for _, section := range document.Sections {
		roots := [][]ir.Block{section.Blocks}
		seen := map[*ir.HeaderFooter]bool{}
		parts := append([]*ir.HeaderFooter{section.Header, section.Footer}, section.Headers...)
		parts = append(parts, section.Footers...)
		for _, part := range parts {
			if part != nil && !seen[part] {
				seen[part] = true
				roots = append(roots, part.Blocks)
			}
		}
		for _, blocks := range roots {
			walkBlocksForEditing(blocks, func(p *ir.Paragraph) { visit(blocks, p) })
		}
	}
}
func nextTextChangeID(document *ir.Document) int {
	next := 0
	visitTextChanges(document, func(_ []ir.Block, p *ir.Paragraph) {
		for _, run := range p.Runs {
			if run.Change != nil && run.Change.ID >= next {
				next = run.Change.ID + 1
			}
		}
	})
	visitChangeRoots(document, func(blocks *[]ir.Block) {
		walkChangedRows(*blocks, func(row *ir.TableRow) {
			if row.Change != nil && row.Change.ID >= next {
				next = row.Change.ID + 1
			}
		})
	})
	return next
}
func resolveTextChange(document *ir.Document, command DocumentCommand) error {
	if command.ChangeID == nil || *command.ChangeID < 0 {
		return errors.New("a tracked change ID is required")
	}
	found := false
	visitTextChanges(document, func(blocks []ir.Block, p *ir.Paragraph) {
		var runs []ir.Run
		offset := 0
		for _, run := range p.Runs {
			if run.Change == nil || run.Change.ID != *command.ChangeID {
				runs = append(runs, run)
				offset += lenUTF16(run.Text)
				continue
			}
			found = true
			if run.Change.Type == "format" {
				if command.ID == "change.reject" && run.Change.Previous != nil {
					run.Attrs = *run.Change.Previous
					run.Style = run.Change.PreviousStyle
				}
				run.Change = nil
				runs = append(runs, run)
				offset += lenUTF16(run.Text)
				continue
			}
			keep := command.ID == "change.accept" && run.Change.Type == "insert" || command.ID == "change.reject" && run.Change.Type == "delete"
			if keep {
				run.Change = nil
				runs = append(runs, run)
				offset += lenUTF16(run.Text)
			} else {
				replaceEditingBookmarkText(blocks, p.ID, offset, offset+lenUTF16(run.Text), 0)
			}
		}
		p.Runs = runs
	})
	visitChangeRoots(document, func(blocks *[]ir.Block) {
		if resolveChangedRows(blocks, *command.ChangeID, command.ID == "change.accept") {
			found = true
		}
	})
	if !found {
		return errors.New("tracked change does not exist")
	}
	return nil
}

func lenUTF16(value string) int { return len(utf16.Encode([]rune(value))) }

func visitChangeRoots(document *ir.Document, visit func(*[]ir.Block)) {
	for index := range document.Sections {
		section := &document.Sections[index]
		visit(&section.Blocks)
		seen := map[*ir.HeaderFooter]bool{}
		parts := append([]*ir.HeaderFooter{section.Header, section.Footer}, section.Headers...)
		parts = append(parts, section.Footers...)
		for _, part := range parts {
			if part != nil && !seen[part] {
				seen[part] = true
				visit(&part.Blocks)
			}
		}
	}
}
func walkChangedRows(blocks []ir.Block, visit func(*ir.TableRow)) {
	for _, block := range blocks {
		switch value := block.(type) {
		case *ir.Table:
			for index := range value.Rows {
				row := &value.Rows[index]
				visit(row)
				for _, cell := range row.Cells {
					walkChangedRows(cell.Blocks, visit)
				}
			}
		case *ir.Shape:
			walkChangedRows(value.Blocks, visit)
		}
	}
}
func resolveChangedRows(blocks *[]ir.Block, id int, accept bool) bool {
	found := false
	var retained []ir.Block
	for _, block := range *blocks {
		switch value := block.(type) {
		case *ir.Table:
			var rows []ir.TableRow
			for _, row := range value.Rows {
				if row.Change != nil && row.Change.ID == id {
					found = true
					keep := accept && row.Change.Type == "insert" || !accept && row.Change.Type == "delete"
					if !keep {
						continue
					}
					row.Change = nil
				}
				for index := range row.Cells {
					if resolveChangedRows(&row.Cells[index].Blocks, id, accept) {
						found = true
					}
				}
				rows = append(rows, row)
			}
			value.Rows = rows
			if len(rows) == 0 {
				continue
			}
		case *ir.Shape:
			if resolveChangedRows(&value.Blocks, id, accept) {
				found = true
			}
		}
		retained = append(retained, block)
	}
	if found && len(retained) == 0 {
		retained = []ir.Block{newEditingParagraph("")}
	}
	*blocks = retained
	return found
}

func hasPendingChanges(blocks []ir.Block) bool {
	found := false
	walkBlocksForEditing(blocks, func(p *ir.Paragraph) {
		for _, run := range p.Runs {
			if run.Change != nil {
				found = true
			}
		}
	})
	walkChangedRows(blocks, func(row *ir.TableRow) {
		if row.Change != nil {
			found = true
		}
	})
	return found
}
