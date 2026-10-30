package ooxml

import (
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"

	"github.com/lean-docs/lean/pkg/ir"
)

type bookmarkEvent struct {
	image    *ir.Image
	shape    *ir.Shape
	noteID   int
	offset   int
	id, name string
	end      bool
}

func paragraphTextLength(paragraph *ir.Paragraph) int {
	size := 0
	for _, run := range paragraph.Runs {
		size += len(utf16.Encode([]rune(run.Text)))
		if run.Break != ir.BreakNone {
			size++
		}
	}
	return size
}
func (state *packageState) prepareBookmarks() {
	state.bookmarkEvents = make(map[*ir.Paragraph][]bookmarkEvent)
	state.anchoredBookmarks = make(map[*ir.Bookmark]bool)
	state.anchoredImages = make(map[*ir.Image]bool)
	state.anchoredShapes = make(map[*ir.Shape]bool)
	next := 0
	collect := func(blocks []ir.Block) {
		paragraphs := map[string]*ir.Paragraph{}
		var first, previous *ir.Paragraph
		walkBlocks(blocks, func(block ir.Block) {
			if p, ok := block.(*ir.Paragraph); ok {
				paragraphs[p.ID] = p
				if first == nil {
					first = p
				}
			}
		})
		walkBlocks(blocks, func(block ir.Block) {
			if shape, ok := block.(*ir.Shape); ok && shape.Anchor != nil && shape.Anchor.ParagraphID != "" {
				if paragraph := paragraphs[shape.Anchor.ParagraphID]; paragraph != nil {
					state.anchoredShapes[shape] = true
					state.bookmarkEvents[paragraph] = append(state.bookmarkEvents[paragraph], bookmarkEvent{offset: shape.Anchor.Offset, shape: shape})
				}
			}
			if image, ok := block.(*ir.Image); ok && image.Anchor != nil && image.Anchor.ParagraphID != "" {
				if paragraph := paragraphs[image.Anchor.ParagraphID]; paragraph != nil {
					state.anchoredImages[image] = true
					state.bookmarkEvents[paragraph] = append(state.bookmarkEvents[paragraph], bookmarkEvent{offset: image.Anchor.Offset, image: image})
				}
			}
		})
		walkBlocks(blocks, func(block ir.Block) {
			if p, ok := block.(*ir.Paragraph); ok {
				previous = p
				return
			}
			bookmark, ok := block.(*ir.Bookmark)
			if !ok {
				return
			}
			start, end := bookmark.Start, bookmark.End
			if start == nil {
				target := previous
				offset := 0
				if target == nil {
					target = first
				} else {
					offset = paragraphTextLength(target)
				}
				if target == nil {
					return
				}
				start = &ir.BookmarkPosition{ParagraphID: target.ID, Offset: offset}
				end = start
			}
			if end == nil {
				return
			}
			a, b := paragraphs[start.ParagraphID], paragraphs[end.ParagraphID]
			if a == nil || b == nil {
				return
			}
			next++
			id := strconv.Itoa(next)
			state.bookmarkEvents[a] = append(state.bookmarkEvents[a], bookmarkEvent{offset: start.Offset, id: id, name: bookmark.Name})
			state.bookmarkEvents[b] = append(state.bookmarkEvents[b], bookmarkEvent{offset: end.Offset, id: id, end: true})
			state.anchoredBookmarks[bookmark] = true
		})
	}
	var body []ir.Block
	for _, section := range state.doc.Sections {
		body = append(body, section.Blocks...)
	}
	collect(body)
	for _, part := range state.auxiliary {
		collect(part.value.Blocks)
	}
	for _, note := range state.notes {
		collect(note.Blocks)
	}
	for paragraph, ids := range state.footnoteIDs {
		for index, id := range ids {
			offset := paragraphTextLength(paragraph)
			if paragraph.Footnotes[index].Offset != nil {
				offset = *paragraph.Footnotes[index].Offset
			}
			state.bookmarkEvents[paragraph] = append(state.bookmarkEvents[paragraph], bookmarkEvent{offset: offset, noteID: id})
		}
	}
	for paragraph, events := range state.bookmarkEvents {
		sort.SliceStable(events, func(a, b int) bool { return events[a].offset < events[b].offset })
		state.bookmarkEvents[paragraph] = events
	}
}
func (state *packageState) writeBookmarkedRuns(output *strings.Builder, paragraph *ir.Paragraph) {
	events := state.bookmarkEvents[paragraph]
	next, offset := 0, 0
	emit := func() {
		for next < len(events) && events[next].offset == offset {
			event := events[next]
			if event.shape != nil {
				writeShapeRun(output, event.shape, state)
			} else if event.image != nil {
				writeDrawingRun(output, event.image, state.imageIDs[event.image], state)
			} else if event.noteID > 0 {
				output.WriteString(`<w:r><w:footnoteReference w:id="` + strconv.Itoa(event.noteID) + `"/></w:r>`)
			} else if event.end {
				output.WriteString(`<w:bookmarkEnd w:id="` + event.id + `"/>`)
			} else {
				output.WriteString(`<w:bookmarkStart w:id="` + event.id + `" w:name="` + escape(event.name) + `"/>`)
			}
			next++
		}
	}
	emit()
	for _, run := range paragraph.Runs {
		text := utf16.Encode([]rune(run.Text))
		start := 0
		for next < len(events) && events[next].offset <= offset+len(text)-start {
			if events[next].offset < offset {
				next++
				continue
			}
			end := start + events[next].offset - offset
			if end > start {
				part := run
				part.Text = string(utf16.Decode(text[start:end]))
				part.Break = ir.BreakNone
				writeRun(output, part, state)
				offset += end - start
				start = end
			}
			emit()
		}
		if start < len(text) {
			part := run
			part.Text = string(utf16.Decode(text[start:]))
			part.Break = ir.BreakNone
			writeRun(output, part, state)
			offset += len(text) - start
		}
		emit()
		if run.Break != ir.BreakNone {
			part := run
			part.Text = ""
			writeRun(output, part, state)
			offset++
			emit()
		}
	}
	emit()
}
