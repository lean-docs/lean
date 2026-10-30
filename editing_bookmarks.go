package lean

import "github.com/lean-docs/lean/pkg/ir"

func updateEditingBookmarks(blocks []ir.Block, change func(*ir.BookmarkPosition, bool)) {
	walkStructuralBlocks(blocks, func(block ir.Block) {
		if bookmark, ok := block.(*ir.Bookmark); ok {
			if bookmark.Start != nil {
				change(bookmark.Start, false)
			}
			if bookmark.End != nil {
				change(bookmark.End, true)
			}
		}
	})
}
func replaceEditingBookmarkText(blocks []ir.Block, id string, start, end, size int) {
	updateEditingBookmarks(blocks, func(position *ir.BookmarkPosition, right bool) {
		if position.ParagraphID != id {
			return
		}
		if position.Offset > end {
			position.Offset += size - (end - start)
		} else if position.Offset >= start {
			position.Offset = start
			if right {
				position.Offset += size
			}
		}
	})
}
