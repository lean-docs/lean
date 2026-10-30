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
	SectionBreak    string              `json:"section_break,omitempty"`
	Footnote        *ir.Footnote        `json:"footnote,omitempty"`
	StyleDefinition *ir.Style           `json:"style_definition,omitempty"`
	Bookmark        *ir.Bookmark        `json:"bookmark,omitempty"`
	ID              string              `json:"id"`
	Section         int                 `json:"section,omitempty"`
	Part            string              `json:"part,omitempty"`
	Variant         ir.HeaderFooterKind `json:"variant,omitempty"`
	Block           string              `json:"block,omitempty"`
	At              int                 `json:"at,omitempty"`
	Start           int                 `json:"start,omitempty"`
	End             int                 `json:"end,omitempty"`
	Text            string              `json:"text,omitempty"`
	Style           string              `json:"style,omitempty"`
	Run             *ir.RunAttrs        `json:"run,omitempty"`
	Paragraph       *ir.ParaAttrs       `json:"paragraph,omitempty"`
	Numbering       *ir.NumberingDef    `json:"numbering,omitempty"`
	Level           int                 `json:"level,omitempty"`
	Restart         *int                `json:"restart,omitempty"`
	Rows            int                 `json:"rows,omitempty"`
	Columns         int                 `json:"columns,omitempty"`
	Page            *ir.PageProperties  `json:"page,omitempty"`
	ColumnsLayout   []ir.Column         `json:"columns_layout,omitempty"`
	HeaderFooter    *ir.HeaderFooter    `json:"header_footer,omitempty"`
	Hyperlink       *ir.Hyperlink       `json:"hyperlink,omitempty"`
}

type DocumentCommandInfo struct {
	ID         string `json:"id"`
	Capability string `json:"capability"`
}

func DOCXCommands() []DocumentCommandInfo {
	return []DocumentCommandInfo{
		{"text.replace", "text"}, {"text.format", "run-formatting"}, {"text.link", "hyperlinks"},
		{"bookmark.set", "bookmarks"}, {"bookmark.remove", "bookmarks"},
		{"footnote.set", "footnotes"}, {"footnote.remove", "footnotes"},
		{"style.set", "paragraph-formatting"}, {"style.remove", "paragraph-formatting"},
		{"paragraph.format", "paragraph-formatting"}, {"paragraph.style", "paragraph-formatting"}, {"paragraph.insert", "text"}, {"paragraph.delete", "text"}, {"paragraph.split", "text"}, {"paragraph.join", "text"},
		{"list.set", "lists"}, {"list.remove", "lists"},
		{"table.insert", "tables"}, {"table.delete", "tables"}, {"table.row.insert", "tables"}, {"table.row.delete", "tables"}, {"table.column.insert", "tables"}, {"table.column.delete", "tables"},
		{"page.layout", "page-layout"}, {"page.columns", "page-columns"}, {"section.insert", "sections"}, {"section.delete", "sections"}, {"section.break", "sections"},
		{"header.set", "headers"}, {"header.remove", "headers"}, {"footer.set", "footers"}, {"footer.remove", "footers"},
	}
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
			if value != nil && value.Kind == command.Variant {
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
	section := &document.Sections[command.Section]
	switch command.ID {
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
					if p, ok := block.(*ir.Paragraph); ok && p.Style == command.Style {
						used = true
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
		next.Blocks = section.Blocks[command.At:]
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
	case "header.set", "footer.set", "header.remove", "footer.remove":
		header := strings.HasPrefix(command.ID, "header")
		primary, others := &section.Header, &section.Headers
		if !header {
			primary, others = &section.Footer, &section.Footers
		}
		kind := command.Variant
		value := command.HeaderFooter
		if strings.HasSuffix(command.ID, ".set") {
			if value == nil {
				return errors.New("header or footer content is required")
			}
			kind = value.Kind
		}
		if kind < ir.HeaderFooterDefault || kind > ir.HeaderFooterEven {
			return errors.New("invalid header or footer variant")
		}
		if *primary != nil && (*primary).Kind == kind {
			*primary = value
			return nil
		}
		for index, current := range *others {
			if current != nil && current.Kind == kind {
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
		for _, row := range table.Rows {
			for _, cell := range row.Cells {
				if cell.RowSpan > 1 || cell.ColSpan > 1 {
					return errors.New("structural edits to merged cells need an explicit merge operation")
				}
			}
		}
		switch command.ID {
		case "table.row.insert":
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
			table.Rows = slices.Insert(table.Rows, command.At, row)
		case "table.row.delete":
			if command.At < 0 || command.At >= len(table.Rows) || len(table.Rows) == 1 {
				return errors.New("invalid table row deletion")
			}
			table.Rows = slices.Delete(table.Rows, command.At, command.At+1)
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
	case "text.replace", "text.format", "text.link":
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
		switch command.ID {
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
	for _, run := range runs {
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
