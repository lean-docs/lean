package lean

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"slices"
	"strings"
	"unicode/utf16"
	"unicode/utf8"

	"github.com/lean-docs/lean/pkg/ir"
)

var ErrInvalidCommand = errors.New("lean: invalid document command")

type DocumentCommand struct {
	Change             *ir.Change          `json:"change,omitempty"`
	ChangeID           *int                `json:"change_id,omitempty"`
	Watermark          *ir.Watermark       `json:"watermark,omitempty"`
	EvenAndOddHeaders  *bool               `json:"even_and_odd_headers,omitempty"`
	DifferentFirstPage *bool               `json:"different_first_page,omitempty"`
	Image              *ir.Image           `json:"image,omitempty"`
	Shape              *ir.Shape           `json:"shape,omitempty"`
	Drawing            *DrawingLayout      `json:"drawing,omitempty"`
	SectionBreak       string              `json:"section_break,omitempty"`
	Footnote           *ir.Footnote        `json:"footnote,omitempty"`
	StyleDefinition    *ir.Style           `json:"style_definition,omitempty"`
	Bookmark           *ir.Bookmark        `json:"bookmark,omitempty"`
	ID                 string              `json:"id"`
	Section            int                 `json:"section,omitempty"`
	Part               string              `json:"part,omitempty"`
	Variant            ir.HeaderFooterType `json:"variant,omitempty"`
	Block              string              `json:"block,omitempty"`
	At                 int                 `json:"at,omitempty"`
	Start              int                 `json:"start,omitempty"`
	End                int                 `json:"end,omitempty"`
	Text               string              `json:"text,omitempty"`
	Style              string              `json:"style,omitempty"`
	Run                *ir.RunAttrs        `json:"run,omitempty"`
	Paragraph          *ir.ParaAttrs       `json:"paragraph,omitempty"`
	Numbering          *ir.NumberingDef    `json:"numbering,omitempty"`
	Level              int                 `json:"level,omitempty"`
	Restart            *int                `json:"restart,omitempty"`
	Rows               int                 `json:"rows,omitempty"`
	Columns            int                 `json:"columns,omitempty"`
	Page               *ir.PageProperties  `json:"page,omitempty"`
	ColumnsLayout      []ir.Column         `json:"columns_layout,omitempty"`
	HeaderFooter       *ir.HeaderFooter    `json:"header_footer,omitempty"`
	Hyperlink          *ir.Hyperlink       `json:"hyperlink,omitempty"`
}

type DocumentCommandInfo struct {
	InputSchema map[string]any `json:"input_schema"`
	ID          string         `json:"id"`
	Capability  string         `json:"capability"`
}

func DOCXCommands() []DocumentCommandInfo {
	commands := []DocumentCommandInfo{
		{ID: "text.format.suggest", Capability: "tracked-changes"}, {ID: "text.suggest", Capability: "tracked-changes"}, {ID: "change.accept", Capability: "tracked-changes"}, {ID: "change.reject", Capability: "tracked-changes"}, {ID: "text.replace", Capability: "text"}, {ID: "text.format", Capability: "run-formatting"}, {ID: "text.link", Capability: "hyperlinks"},
		{ID: "bookmark.set", Capability: "bookmarks"}, {ID: "bookmark.remove", Capability: "bookmarks"},
		{ID: "footnote.set", Capability: "footnotes"}, {ID: "footnote.remove", Capability: "footnotes"},
		{ID: "style.set", Capability: "paragraph-formatting"}, {ID: "style.remove", Capability: "paragraph-formatting"},
		{ID: "paragraph.format", Capability: "paragraph-formatting"}, {ID: "paragraph.style", Capability: "paragraph-formatting"}, {ID: "paragraph.insert", Capability: "text"}, {ID: "paragraph.delete", Capability: "text"}, {ID: "paragraph.split", Capability: "text"}, {ID: "paragraph.join", Capability: "text"},
		{ID: "list.set", Capability: "lists"}, {ID: "list.remove", Capability: "lists"},
		{ID: "table.insert.suggest", Capability: "tracked-changes"}, {ID: "table.row.insert.suggest", Capability: "tracked-changes"}, {ID: "table.row.delete.suggest", Capability: "tracked-changes"}, {ID: "table.style", Capability: "table-styles"}, {ID: "table.insert", Capability: "tables"}, {ID: "table.delete", Capability: "tables"}, {ID: "table.row.insert", Capability: "tables"}, {ID: "table.row.delete", Capability: "tables"}, {ID: "table.column.insert", Capability: "tables"}, {ID: "table.column.delete", Capability: "tables"},
		{ID: "image.insert", Capability: "images"}, {ID: "shape.insert", Capability: "shapes"}, {ID: "drawing.layout", Capability: "drawing-anchors"}, {ID: "drawing.delete", Capability: "drawing-anchors"},
		{ID: "page.layout", Capability: "page-layout"}, {ID: "page.columns", Capability: "page-columns"}, {ID: "section.insert", Capability: "sections"}, {ID: "section.delete", Capability: "sections"}, {ID: "section.break", Capability: "sections"},
		{ID: "watermark.set", Capability: "text-watermarks"}, {ID: "watermark.remove", Capability: "text-watermarks"}, {ID: "header.options", Capability: "headers"}, {ID: "header.set", Capability: "headers"}, {ID: "header.remove", Capability: "headers"}, {ID: "footer.set", Capability: "footers"}, {ID: "footer.remove", Capability: "footers"},
	}
	for index := range commands {
		commands[index].InputSchema = DOCXCommandSchema(commands[index].ID)
	}
	return commands
}

func ApplyDOCXCommands(document *ir.Document, commands []DocumentCommand) (*ir.Document, error) {
	if document == nil || len(commands) == 0 || len(commands) > 1000 {
		return nil, ErrInvalidCommand
	}
	data, err := json.Marshal(document)
	if err != nil {
		return nil, err
	}
	var edited ir.Document
	if err := json.Unmarshal(data, &edited); err != nil {
		return nil, err
	}
	encodedCommands, err := json.Marshal(commands)
	if err != nil {
		return nil, err
	}
	var operations []DocumentCommand
	if err := json.Unmarshal(encodedCommands, &operations); err != nil {
		return nil, err
	}
	for index, command := range operations {
		if err := applyDOCXCommand(&edited, command); err != nil {
			return nil, fmt.Errorf("%w at command %d (%s): %s", ErrInvalidCommand, index, command.ID, err)
		}
	}
	report := inspectDocument(&edited)
	if !report.Editable {
		return nil, fmt.Errorf("%w: %s", ErrUnsupportedDocument, strings.Join(report.Unsupported, ", "))
	}
	return &edited, nil
}

func commandBlocks(document *ir.Document, command DocumentCommand) (*[]ir.Block, error) {
	if command.Section < 0 || command.Section >= len(document.Sections) {
		return nil, errors.New("section does not exist")
	}
	section := &document.Sections[command.Section]
	switch command.Part {
	case "", "body":
		return &section.Blocks, nil
	case "header", "footer":
		values := append([]*ir.HeaderFooter{section.Header}, section.Headers...)
		if command.Part == "footer" {
			values = append([]*ir.HeaderFooter{section.Footer}, section.Footers...)
		}
		for _, value := range values {
			if value != nil && value.Type == command.Variant {
				return &value.Blocks, nil
			}
		}
		return nil, errors.New("header or footer does not exist")
	default:
		return nil, errors.New("unknown document part")
	}
}

func findCommandBlock(blocks *[]ir.Block, id string) (*[]ir.Block, int, error) {
	for index, block := range *blocks {
		if block.BlockID() == id {
			return blocks, index, nil
		}
		if shape, ok := block.(*ir.Shape); ok {
			if parent, index, err := findCommandBlock(&shape.Blocks, id); err == nil {
				return parent, index, nil
			}
		}
		if table, ok := block.(*ir.Table); ok {
			for row := range table.Rows {
				for cell := range table.Rows[row].Cells {
					parent, index, err := findCommandBlock(&table.Rows[row].Cells[cell].Blocks, id)
					if err == nil {
						return parent, index, nil
					}
				}
			}
		}
	}
	return nil, 0, errors.New("block does not exist")
}

func applyDOCXCommand(document *ir.Document, command DocumentCommand) error {
	if command.Section < 0 || command.Section >= len(document.Sections) {
		return errors.New("section does not exist")
	}
	if command.ID == "change.accept" || command.ID == "change.reject" {
		return resolveTextChange(document, command)
	}
	if command.ID == "table.insert.suggest" {
		if command.Change == nil || command.Change.Author == "" {
			return errors.New("tracked table insertion requires author and date")
		}
		direct := command
		direct.ID = "table.insert"
		if err := applyDOCXCommand(document, direct); err != nil {
			return err
		}
		blocks, err := commandBlocks(document, direct)
		if err != nil {
			return err
		}
		table, ok := (*blocks)[command.At].(*ir.Table)
		if !ok {
			return errors.New("inserted table is unavailable")
		}
		id := nextTextChangeID(document)
		for index := range table.Rows {
			change := *command.Change
			change.ID = id + index
			change.Type = "insert"
			table.Rows[index].Change = &change
		}
		return nil
	}
	section := &document.Sections[command.Section]
	if command.ID == "section.delete" {
		if hasPendingChanges(section.Blocks) {
			return errors.New("resolve tracked changes before deleting the section")
		}
		parts := append([]*ir.HeaderFooter{section.Header, section.Footer}, section.Headers...)
		parts = append(parts, section.Footers...)
		for _, part := range parts {
			if part != nil && hasPendingChanges(part.Blocks) {
				return errors.New("resolve tracked changes before deleting the section")
			}
		}
	}
	if command.ID == "header.set" || command.ID == "header.remove" || command.ID == "footer.set" || command.ID == "footer.remove" {
		parts := append([]*ir.HeaderFooter{section.Header}, section.Headers...)
		if strings.HasPrefix(command.ID, "footer.") {
			parts = append([]*ir.HeaderFooter{section.Footer}, section.Footers...)
		}
		targetType := command.Variant
		if command.HeaderFooter != nil {
			targetType = command.HeaderFooter.Type
		}
		for _, part := range parts {
			if part != nil && part.Type == targetType && hasPendingChanges(part.Blocks) {
				return errors.New("resolve tracked changes before replacing this header or footer")
			}
		}
	}
	switch command.ID {
	case "header.options":
		if command.EvenAndOddHeaders == nil && command.DifferentFirstPage == nil {
			return errors.New("header options are required")
		}
		if command.EvenAndOddHeaders != nil {
			document.EvenAndOddHeaders = command.EvenAndOddHeaders
		}
		if command.DifferentFirstPage != nil {
			section.Properties.DifferentFirstPage = command.DifferentFirstPage
		}
		return nil
	case "style.set":
		if command.StyleDefinition == nil || command.StyleDefinition.ID == "" {
			return errors.New("a named style definition is required")
		}
		if document.Styles.Named == nil {
			document.Styles.Named = make(map[string]ir.Style)
		}
		document.Styles.Named[command.StyleDefinition.ID] = *command.StyleDefinition
		return nil
	case "style.remove":
		used := false
		for _, section := range document.Sections {
			parts := [][]ir.Block{section.Blocks}
			for _, part := range append(append([]*ir.HeaderFooter{section.Header, section.Footer}, section.Headers...), section.Footers...) {
				if part != nil {
					parts = append(parts, part.Blocks)
				}
			}
			for _, blocks := range parts {
				walkStructuralBlocks(blocks, func(block ir.Block) {
					if p, ok := block.(*ir.Paragraph); ok {
						if p.Style == command.Style {
							used = true
						}
						for _, run := range p.Runs {
							if run.Style == command.Style || run.Change != nil && run.Change.PreviousStyle == command.Style {
								used = true
							}
						}
					}
				})
			}
		}
		if used {
			return errors.New("the style is still used by document content")
		}
		for _, style := range document.Styles.Named {
			if style.BasedOn == command.Style {
				return errors.New("another style inherits this definition")
			}
		}
		delete(document.Styles.Named, command.Style)
		return nil
	case "section.break":
		section.Break = command.SectionBreak
		return nil
	case "page.layout":
		if command.Page == nil || !validEditingPage(*command.Page) {
			return errors.New("page dimensions and margins must leave room for content")
		}
		section.Properties = *command.Page
		return nil
	case "page.columns":
		if len(command.ColumnsLayout) > 100 {
			return errors.New("too many page columns")
		}
		available := section.Properties.Width - section.Properties.MarginLeft - section.Properties.MarginRight - section.Properties.Gutter
		total := 0.0
		for index, column := range command.ColumnsLayout {
			if !finite(column.Width) || !finite(column.Spacing) || column.Width <= 0 || column.Spacing < 0 {
				return errors.New("invalid column measurements")
			}
			total += column.Width
			if index < len(command.ColumnsLayout)-1 {
				total += column.Spacing
			}
		}
		if total > available+0.01 {
			return errors.New("columns exceed the page content width")
		}
		section.Columns = append([]ir.Column(nil), command.ColumnsLayout...)
		return nil
	case "section.insert":
		if command.At < 0 || command.At > len(section.Blocks) {
			return errors.New("invalid section split position")
		}
		data, err := json.Marshal(section)
		if err != nil {
			return err
		}
		var next ir.Section
		if err := json.Unmarshal(data, &next); err != nil {
			return err
		}
		next.Blocks = slices.Clone(section.Blocks[command.At:])
		section.Blocks = section.Blocks[:command.At]
		if len(section.Blocks) == 0 {
			section.Blocks = []ir.Block{newEditingParagraph("")}
		}
		if len(next.Blocks) == 0 {
			next.Blocks = []ir.Block{newEditingParagraph("")}
		}
		document.Sections = slices.Insert(document.Sections, command.Section+1, next)
		return nil
	case "section.delete":
		if len(document.Sections) == 1 {
			return errors.New("the document needs at least one section")
		}
		document.Sections = slices.Delete(document.Sections, command.Section, command.Section+1)
		return nil
	case "watermark.set", "watermark.remove":
		if command.ID == "watermark.set" && command.Watermark == nil {
			return errors.New("watermark content is required")
		}
		parts := append([]*ir.HeaderFooter{section.Header}, section.Headers...)
		if section.Header == nil && len(section.Headers) == 0 && command.ID == "watermark.set" {
			section.Header = &ir.HeaderFooter{Type: ir.HeaderFooterDefault, Blocks: []ir.Block{newEditingParagraph("")}}
			parts = []*ir.HeaderFooter{section.Header}
		}
		for _, part := range parts {
			if part == nil {
				continue
			}
			var first *ir.Paragraph
			walkBlocksForEditing(part.Blocks, func(paragraph *ir.Paragraph) {
				if first == nil {
					first = paragraph
				}
				runs := paragraph.Runs[:0]
				for _, run := range paragraph.Runs {
					hadWatermark := run.Watermark != nil
					run.Watermark = nil
					if hadWatermark && run.Text == "" && run.Break == ir.BreakNone && run.Field == nil && run.Hyperlink == nil && run.Style == "" && run.Attrs == (ir.RunAttrs{}) {
						continue
					}
					runs = append(runs, run)
				}
				paragraph.Runs = runs
			})
			if command.ID == "watermark.set" {
				if first == nil {
					first = newEditingParagraph("")
					part.Blocks = append([]ir.Block{first}, part.Blocks...)
				}
				first.Runs = append([]ir.Run{{Watermark: command.Watermark}}, first.Runs...)
			}
		}
		return nil
	case "header.set", "footer.set", "header.remove", "footer.remove":
		header := strings.HasPrefix(command.ID, "header")
		primary, others := &section.Header, &section.Headers
		if !header {
			primary, others = &section.Footer, &section.Footers
		}
		partType := command.Variant
		value := command.HeaderFooter
		if strings.HasSuffix(command.ID, ".set") {
			if value == nil {
				return errors.New("header or footer content is required")
			}
			partType = value.Type
		}
		if partType < ir.HeaderFooterDefault || partType > ir.HeaderFooterEven {
			return errors.New("invalid header or footer variant")
		}
		if *primary != nil && (*primary).Type == partType {
			*primary = value
			return nil
		}
		for index, current := range *others {
			if current != nil && current.Type == partType {
				if value == nil {
					*others = slices.Delete(*others, index, index+1)
				} else {
					(*others)[index] = value
				}
				return nil
			}
		}
		if value != nil {
			if *primary == nil {
				*primary = value
			} else {
				*others = append(*others, value)
			}
		}
		return nil
	}
	blocks, err := commandBlocks(document, command)
	if err != nil {
		return err
	}
	if strings.HasPrefix(command.ID, "image.") || strings.HasPrefix(command.ID, "shape.") || strings.HasPrefix(command.ID, "drawing.") {
		return applyDrawingCommand(blocks, command)
	}
	if command.ID == "paragraph.insert" || command.ID == "table.insert" {
		if command.At < 0 || command.At > len(*blocks) {
			return errors.New("invalid insertion position")
		}
		var block ir.Block = newEditingParagraph(command.Text)
		if command.ID == "table.insert" {
			if command.Rows < 1 || command.Columns < 1 || command.Rows > 1000 || command.Columns > 1000 || command.Rows*command.Columns > 10000 {
				return errors.New("invalid table size")
			}
			table := &ir.Table{ID: rand.Text(), ColumnWidths: make([]float64, command.Columns)}
			if command.Style != "" {
				style, exists := document.Styles.Named[command.Style]
				if !exists || style.Table == nil {
					return errors.New("a defined table style is required")
				}
				table.Style = command.Style
			}
			width := (section.Properties.Width - section.Properties.MarginLeft - section.Properties.MarginRight - section.Properties.Gutter) / float64(command.Columns)
			if width <= 0 {
				width = 72
			}
			for index := range table.ColumnWidths {
				table.ColumnWidths[index] = width
			}
			for row := 0; row < command.Rows; row++ {
				value := ir.TableRow{}
				for column := 0; column < command.Columns; column++ {
					value.Cells = append(value.Cells, newEditingCell())
				}
				table.Rows = append(table.Rows, value)
			}
			block = table
		}
		*blocks = slices.Insert(*blocks, command.At, block)
		return nil
	}
	if command.ID == "bookmark.set" {
		if command.Bookmark == nil || command.Bookmark.Name == "" {
			return errors.New("a named bookmark range is required")
		}
		if command.Bookmark.ID == "" {
			command.Bookmark.ID = rand.Text()
		}
		if parent, index, err := findCommandBlock(blocks, command.Bookmark.ID); err == nil {
			if _, ok := (*parent)[index].(*ir.Bookmark); !ok {
				return errors.New("bookmark ID belongs to another block")
			}
			(*parent)[index] = command.Bookmark
		} else {
			*blocks = append(*blocks, command.Bookmark)
		}
		return nil
	}
	if command.Block == "" {
		return errors.New("a block ID is required")
	}
	parent, index, err := findCommandBlock(blocks, command.Block)
	if err != nil {
		return err
	}
	block := (*parent)[index]
	if command.ID == "bookmark.remove" {
		if block.BlockType() != ir.BlockBookmark {
			return errors.New("a bookmark is required")
		}
		*parent = slices.Delete(*parent, index, index+1)
		return nil
	}
	if command.ID == "paragraph.delete" || command.ID == "table.delete" {
		if hasPendingChanges([]ir.Block{block}) {
			return errors.New("resolve tracked changes before deleting this block")
		}
		if command.ID == "paragraph.delete" && block.BlockType() != ir.BlockParagraph || command.ID == "table.delete" && block.BlockType() != ir.BlockTable {
			return errors.New("block type does not match the command")
		}
		*parent = slices.Delete(*parent, index, index+1)
		if len(*parent) == 0 {
			*parent = append(*parent, newEditingParagraph(""))
		}
		return nil
	}
	if strings.HasPrefix(command.ID, "table.") {
		table, ok := block.(*ir.Table)
		if !ok {
			return errors.New("a table is required")
		}
		if command.ID == "table.style" {
			if command.Style != "" && command.Style != "TableNormal" {
				style, exists := document.Styles.Named[command.Style]
				if !exists || style.Table == nil {
					return errors.New("a defined table style is required")
				}
			}
			table.Style = command.Style
			return nil
		}
		if command.ID != "table.row.insert.suggest" && command.ID != "table.row.delete.suggest" && hasPendingChanges([]ir.Block{table}) {
			return errors.New("resolve tracked changes before changing the table structure")
		}
		for _, row := range table.Rows {
			if row.Change != nil && command.ID != "table.row.insert.suggest" && command.ID != "table.row.delete.suggest" {
				return errors.New("resolve tracked row changes before changing the table structure")
			}
			for _, cell := range row.Cells {
				if cell.RowSpan > 1 || cell.ColSpan > 1 {
					return errors.New("structural edits to merged cells need an explicit merge operation")
				}
			}
		}
		switch command.ID {
		case "table.row.insert", "table.row.insert.suggest":
			if len(table.Rows) == 0 {
				return errors.New("table has no rows")
			}
			if command.At < 0 || command.At > len(table.Rows) || len(table.Rows) >= 10000 {
				return errors.New("invalid table row")
			}
			row := ir.TableRow{}
			for range table.Rows[0].Cells {
				row.Cells = append(row.Cells, newEditingCell())
			}
			if command.ID == "table.row.insert.suggest" {
				if command.Change == nil || command.Change.Author == "" {
					return errors.New("tracked row insertion requires author and date")
				}
				change := *command.Change
				change.ID = nextTextChangeID(document)
				change.Type = "insert"
				row.Change = &change
			}
			table.Rows = slices.Insert(table.Rows, command.At, row)
		case "table.row.delete", "table.row.delete.suggest":
			if command.At < 0 || command.At >= len(table.Rows) || len(table.Rows) == 1 {
				return errors.New("invalid table row deletion")
			}
			if command.ID == "table.row.delete.suggest" {
				if command.Change == nil || command.Change.Author == "" || table.Rows[command.At].Change != nil {
					return errors.New("tracked row deletion requires author, date, and an unchanged row")
				}
				change := *command.Change
				change.ID = nextTextChangeID(document)
				change.Type = "delete"
				table.Rows[command.At].Change = &change
			} else {
				table.Rows = slices.Delete(table.Rows, command.At, command.At+1)
			}
		case "table.column.insert", "table.column.delete":
			if len(table.Rows) == 0 {
				return errors.New("table has no rows")
			}
			count := len(table.Rows[0].Cells)
			if command.At < 0 || command.At > count || command.ID == "table.column.delete" && (command.At == count || count == 1) {
				return errors.New("invalid table column")
			}
			for row := range table.Rows {
				if len(table.Rows[row].Cells) != count {
					return errors.New("table rows have different column counts")
				}
			}
			for row := range table.Rows {
				if command.ID == "table.column.insert" {
					table.Rows[row].Cells = slices.Insert(table.Rows[row].Cells, command.At, newEditingCell())
				} else {
					table.Rows[row].Cells = slices.Delete(table.Rows[row].Cells, command.At, command.At+1)
				}
			}
			if len(table.ColumnWidths) == count {
				if command.ID == "table.column.insert" {
					table.ColumnWidths = slices.Insert(table.ColumnWidths, command.At, table.ColumnWidths[max(0, command.At-1)])
				} else {
					table.ColumnWidths = slices.Delete(table.ColumnWidths, command.At, command.At+1)
				}
			}
		default:
			return errors.New("unknown table command")
		}
		return nil
	}
	paragraph, ok := block.(*ir.Paragraph)
	if !ok {
		return errors.New("a paragraph is required")
	}
	switch command.ID {
	case "footnote.set", "footnote.remove":
		if command.Footnote == nil {
			return errors.New("a footnote definition is required")
		}
		note := *command.Footnote
		if note.ID == "" {
			if command.ID == "footnote.remove" {
				return errors.New("a footnote ID is required")
			}
			note.ID = rand.Text()
		}
		for index, current := range paragraph.Footnotes {
			if current.ID == note.ID {
				if command.ID == "footnote.remove" {
					paragraph.Footnotes = slices.Delete(paragraph.Footnotes, index, index+1)
				} else {
					paragraph.Footnotes[index] = note
				}
				return nil
			}
		}
		if command.ID == "footnote.remove" {
			return errors.New("footnote does not exist")
		}
		paragraph.Footnotes = append(paragraph.Footnotes, note)
	case "paragraph.format":
		if command.Paragraph == nil {
			return errors.New("paragraph formatting is required")
		}
		paragraph.Para = paragraph.Para.Overlay(*command.Paragraph)
	case "paragraph.style":
		if command.Style != "" {
			style, exists := document.Styles.Named[command.Style]
			if !exists {
				var known bool
				style, known = ir.BuiltinStyle(command.Style)
				if !known {
					return errors.New("paragraph style does not exist")
				}
				if document.Styles.Named == nil {
					document.Styles.Named = make(map[string]ir.Style)
				}
				document.Styles.Named[command.Style] = style
			}
			if style.IsChar {
				return errors.New("the selected style is a character style")
			}
		}
		paragraph.Style = command.Style
	case "paragraph.split":
		before, after, err := splitEditingRuns(paragraph.Runs, command.At)
		if err != nil {
			return err
		}
		next := *paragraph
		next.ID = rand.Text()
		updateEditingBookmarks(*blocks, func(position *ir.BookmarkPosition, _ bool) {
			if position.ParagraphID == paragraph.ID && position.Offset >= command.At {
				position.ParagraphID = next.ID
				position.Offset -= command.At
			}
		})
		next.Runs = after
		next.Footnotes = nil
		var retained []ir.Footnote
		for _, note := range paragraph.Footnotes {
			if note.Offset == nil || *note.Offset >= command.At {
				if note.Offset != nil {
					position := *note.Offset - command.At
					note.Offset = &position
				}
				next.Footnotes = append(next.Footnotes, note)
			} else {
				retained = append(retained, note)
			}
		}
		paragraph.Footnotes = retained
		paragraph.Runs = before
		*parent = slices.Insert(*parent, index+1, ir.Block(&next))
	case "paragraph.join":
		if index+1 >= len(*parent) {
			return errors.New("there is no next paragraph")
		}
		next, ok := (*parent)[index+1].(*ir.Paragraph)
		if !ok {
			return errors.New("the next block is not a paragraph")
		}
		size := 0
		for _, run := range paragraph.Runs {
			size += len(utf16.Encode([]rune(run.Text)))
			if run.Break != ir.BreakNone {
				size++
			}
		}
		updateEditingBookmarks(*blocks, func(position *ir.BookmarkPosition, _ bool) {
			if position.ParagraphID == next.ID {
				position.ParagraphID = paragraph.ID
				position.Offset += size
			}
		})
		for index := range paragraph.Footnotes {
			if paragraph.Footnotes[index].Offset == nil {
				position := size
				paragraph.Footnotes[index].Offset = &position
			}
		}
		for index := range next.Footnotes {
			if next.Footnotes[index].Offset != nil {
				position := *next.Footnotes[index].Offset + size
				next.Footnotes[index].Offset = &position
			}
		}
		paragraph.Runs = append(paragraph.Runs, next.Runs...)
		paragraph.Footnotes = append(paragraph.Footnotes, next.Footnotes...)
		*parent = slices.Delete(*parent, index+1, index+2)
	case "list.remove":
		paragraph.Numbering = nil
	case "list.set":
		if command.Numbering == nil || command.Level < 0 {
			return errors.New("a numbering definition and level are required")
		}
		definition := *command.Numbering
		if definition.ID == "" {
			definition.ID = rand.Text()
		}
		found := false
		for _, level := range definition.Levels {
			if level.Level == command.Level {
				found = true
			}
		}
		if !found {
			return errors.New("numbering level does not exist")
		}
		replaced := false
		for index := range document.Styles.Numbering {
			if document.Styles.Numbering[index].ID == definition.ID {
				document.Styles.Numbering[index] = definition
				replaced = true
			}
		}
		if !replaced {
			document.Styles.Numbering = append(document.Styles.Numbering, definition)
		}
		paragraph.Numbering = &ir.NumberingRef{ID: definition.ID, Level: command.Level, Restart: command.Restart}
	case "text.format.suggest", "text.suggest", "text.replace", "text.format", "text.link":
		if command.Start < 0 || command.End < command.Start {
			return errors.New("invalid text selection")
		}
		before, remaining, err := splitEditingRuns(paragraph.Runs, command.Start)
		if err != nil {
			return err
		}
		selected, after, err := splitEditingRuns(remaining, command.End-command.Start)
		if err != nil {
			return err
		}
		for _, run := range selected {
			if run.Change != nil {
				return errors.New("resolve the existing tracked change before editing this selection")
			}
		}
		switch command.ID {
		case "text.format.suggest":
			if command.Change == nil || command.Change.Author == "" || command.Run == nil {
				return errors.New("tracked formatting requires an author, date, and formatting")
			}
			id := nextTextChangeID(document)
			for index := range selected {
				previous := selected[index].Attrs
				revision := *command.Change
				revision.ID = id + index
				revision.Type = "format"
				revision.Previous = &previous
				revision.PreviousStyle = selected[index].Style
				selected[index].Attrs = previous.Overlay(*command.Run)
				selected[index].Change = &revision
			}
		case "text.suggest":
			if command.Change == nil || command.Change.Author == "" || !utf8.ValidString(command.Text) || strings.ContainsAny(command.Text, "\r\n") {
				return errors.New("tracked text requires an author, date, and valid paragraph text")
			}
			id := nextTextChangeID(document)
			for i := range selected {
				if selected[i].Field != nil || selected[i].Hyperlink != nil || selected[i].Break != ir.BreakNone {
					return errors.New("tracked changes to fields, links, and breaks are not supported")
				}
				revision := *command.Change
				revision.ID = id
				revision.Type = "delete"
				selected[i].Change = &revision
			}
			var attrs ir.RunAttrs
			if len(selected) > 0 {
				attrs = selected[0].Attrs
			} else if len(before) > 0 {
				attrs = before[len(before)-1].Attrs
			}
			if command.Text != "" {
				revision := *command.Change
				revision.ID = id + 1
				revision.Type = "insert"
				selected = append(selected, ir.Run{Text: command.Text, Attrs: attrs, Change: &revision})
			}
		case "text.replace":
			size := len(utf16.Encode([]rune(command.Text)))
			replaceEditingBookmarkText(*blocks, paragraph.ID, command.Start, command.End, size)
			for index := range paragraph.Footnotes {
				position := paragraph.Footnotes[index].Offset
				if position == nil {
					continue
				}
				if *position > command.End {
					*position += size - (command.End - command.Start)
				} else if *position >= command.Start {
					*position = command.Start + size
				}
			}
			if !utf8.ValidString(command.Text) || strings.ContainsAny(command.Text, "\r\n") {
				return errors.New("replacement text must be valid paragraph text")
			}
			var attrs ir.RunAttrs
			if len(selected) > 0 {
				attrs = selected[0].Attrs
			} else if len(before) > 0 {
				attrs = before[len(before)-1].Attrs
			}
			if command.Run != nil {
				attrs = attrs.Overlay(*command.Run)
			}
			selected = nil
			if command.Text != "" {
				selected = []ir.Run{{Text: command.Text, Attrs: attrs}}
			}
		case "text.format":
			if command.Run == nil {
				return errors.New("character formatting is required")
			}
			for index := range selected {
				selected[index].Attrs = selected[index].Attrs.Overlay(*command.Run)
			}
		case "text.link":
			for index := range selected {
				selected[index].Hyperlink = command.Hyperlink
			}
		}
		paragraph.Runs = append(append(before, selected...), after...)
	default:
		return errors.New("unknown document command")
	}
	return nil
}

func splitEditingRuns(runs []ir.Run, offset int) ([]ir.Run, []ir.Run, error) {
	if offset < 0 {
		return nil, nil, errors.New("invalid text position")
	}
	var before, after []ir.Run
	for index := 0; index < len(runs); index++ {
		run := runs[index]
		if run.Field != nil && run.Field.Type == "begin" {
			depth, end, length := 1, index+1, 0
			for ; end < len(runs); end++ {
				item := runs[end]
				if item.Field != nil {
					if item.Field.Type == "begin" {
						depth++
					}
					if item.Field.Type == "end" {
						depth--
					}
				}
				length += len(utf16.Encode([]rune(item.Text)))
				if item.Break != ir.BreakNone {
					length++
				}
				if depth == 0 {
					break
				}
			}
			if depth != 0 {
				return nil, nil, errors.New("text editing cannot split a field across paragraphs")
			}
			if offset == 0 {
				return before, append(after, runs[index:]...), nil
			}
			if offset < length {
				return nil, nil, errors.New("text editing cannot split a document field")
			}
			before = append(before, runs[index:end+1]...)
			offset -= length
			index = end
			continue
		}
		size := len(utf16.Encode([]rune(run.Text)))
		if run.Break != ir.BreakNone {
			size++
		}
		if offset >= size {
			before = append(before, run)
			offset -= size
			continue
		}
		if offset == 0 {
			after = append(after, run)
			continue
		}
		encoded := utf16.Encode([]rune(run.Text))
		if offset == len(encoded) {
			left, right := run, run
			left.Break = ir.BreakNone
			right.Text = ""
			before = append(before, left)
			after = append(after, right)
			offset = 0
			continue
		}
		if offset < len(encoded) && utf16.IsSurrogate(rune(encoded[offset-1])) && encoded[offset-1] >= 0xD800 && encoded[offset-1] <= 0xDBFF {
			return nil, nil, errors.New("text position splits a Unicode character")
		}
		left, right := run, run
		left.Text = string(utf16.Decode(encoded[:offset]))
		left.Break = ir.BreakNone
		right.Text = string(utf16.Decode(encoded[offset:]))
		before = append(before, left)
		after = append(after, right)
		offset = 0
	}
	if offset != 0 {
		return nil, nil, errors.New("text position exceeds paragraph length")
	}
	return before, after, nil
}
func newEditingParagraph(text string) *ir.Paragraph {
	return &ir.Paragraph{ID: rand.Text(), Runs: []ir.Run{{Text: text}}}
}
func newEditingCell() ir.TableCell {
	return ir.TableCell{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{newEditingParagraph("")}, Padding: ir.Padding{Top: 3, Bottom: 3, Left: 5, Right: 5}}
}
func finite(value float64) bool { return !math.IsNaN(value) && !math.IsInf(value, 0) }
func validEditingPage(page ir.PageProperties) bool {
	for _, value := range []float64{page.Width, page.Height, page.MarginTop, page.MarginBottom, page.MarginLeft, page.MarginRight, page.Gutter} {
		if !finite(value) || value < 0 {
			return false
		}
	}
	return page.Width > page.MarginLeft+page.MarginRight+page.Gutter && page.Height > page.MarginTop+page.MarginBottom && (page.Orientation == ir.Portrait || page.Orientation == ir.Landscape)
}

func walkBlocksForEditing(blocks []ir.Block, visit func(*ir.Paragraph)) {
	for _, block := range blocks {
		switch value := block.(type) {
		case *ir.Paragraph:
			visit(value)
		case *ir.Table:
			for _, row := range value.Rows {
				for _, cell := range row.Cells {
					walkBlocksForEditing(cell.Blocks, visit)
				}
			}
		case *ir.Shape:
			walkBlocksForEditing(value.Blocks, visit)
		}
	}
}
