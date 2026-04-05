package ir_test

import (
	"testing"

	// "github.com/lean-docs/lean/pkg/ir"
	// "github.com/stretchr/testify/assert"
	// "github.com/stretchr/testify/require"
)

// --- Cluster 12: Mutable IR ---

// C12.1 InsertText inserts text into a run at the given offset.
func TestInsertText(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{
	// 		ID:   "p1",
	// 		Runs: []ir.Run{{Text: "Hello world"}},
	// 	},
	// }
	// err := ir.InsertText(doc, "p1", 5, ", brave")
	// require.NoError(t, err)
	// p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "Hello, brave world", p.Runs[0].Text)
}

// C12.2 InsertText at offset 0 prepends text.
func TestInsertTextAtStart(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{
	// 		ID:   "p1",
	// 		Runs: []ir.Run{{Text: "world"}},
	// 	},
	// }
	// err := ir.InsertText(doc, "p1", 0, "Hello ")
	// require.NoError(t, err)
	// p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "Hello world", p.Runs[0].Text)
}

// C12.3 DeleteText removes text from a run.
func TestDeleteText(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{
	// 		ID:   "p1",
	// 		Runs: []ir.Run{{Text: "Hello, world"}},
	// 	},
	// }
	// err := ir.DeleteText(doc, "p1", 5, 1) // delete the comma
	// require.NoError(t, err)
	// p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "Hello world", p.Runs[0].Text)
}

// C12.4 DeleteText with out-of-range offset returns error.
func TestDeleteTextOutOfRange(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{
	// 		ID:   "p1",
	// 		Runs: []ir.Run{{Text: "short"}},
	// 	},
	// }
	// err := ir.DeleteText(doc, "p1", 100, 5)
	// assert.Error(t, err)
}

// C12.5 FormatRun applies formatting to a character range within a paragraph.
func TestFormatRun(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{
	// 		ID:   "p1",
	// 		Runs: []ir.Run{{Text: "Hello, world"}},
	// 	},
	// }
	// err := ir.FormatRun(doc, "p1", 0, 5, ir.RunAttrs{Bold: true})
	// require.NoError(t, err)
	// p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// // After formatting, the paragraph should have multiple runs:
	// // "Hello" (bold) + ", world" (not bold)
	// require.GreaterOrEqual(t, len(p.Runs), 2)
	// assert.True(t, p.Runs[0].Attrs.Bold)
	// assert.Equal(t, "Hello", p.Runs[0].Text)
	// assert.False(t, p.Runs[1].Attrs.Bold)
}

// C12.6 InsertBlock inserts a new paragraph at a given position in a section.
func TestInsertParagraph(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "first"}}},
	// 	&ir.Paragraph{ID: "p3", Runs: []ir.Run{{Text: "third"}}},
	// }
	// newBlock := &ir.Paragraph{ID: "p2", Runs: []ir.Run{{Text: "second"}}}
	// err := ir.InsertBlock(doc, 0, 1, newBlock)
	// require.NoError(t, err)
	// require.Len(t, doc.Sections[0].Blocks, 3)
	// p := doc.Sections[0].Blocks[1].(*ir.Paragraph)
	// assert.Equal(t, "p2", p.ID)
	// assert.Equal(t, "second", p.Runs[0].Text)
}

// C12.7 DeleteBlock removes a block by ID.
func TestDeleteParagraph(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "first"}}},
	// 	&ir.Paragraph{ID: "p2", Runs: []ir.Run{{Text: "second"}}},
	// 	&ir.Paragraph{ID: "p3", Runs: []ir.Run{{Text: "third"}}},
	// }
	// err := ir.DeleteBlock(doc, "p2")
	// require.NoError(t, err)
	// require.Len(t, doc.Sections[0].Blocks, 2)
	// assert.Equal(t, "p1", doc.Sections[0].Blocks[0].BlockID())
	// assert.Equal(t, "p3", doc.Sections[0].Blocks[1].BlockID())
}

// C12.8 SplitParagraph splits one paragraph into two at the given text offset.
func TestSplitParagraph(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{
	// 		ID:   "p1",
	// 		Runs: []ir.Run{{Text: "Hello world"}},
	// 	},
	// }
	// err := ir.SplitParagraph(doc, "p1", 5) // split after "Hello"
	// require.NoError(t, err)
	// require.Len(t, doc.Sections[0].Blocks, 2)
	// p1 := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// p2 := doc.Sections[0].Blocks[1].(*ir.Paragraph)
	// assert.Equal(t, "Hello", p1.Runs[0].Text)
	// assert.Equal(t, " world", p2.Runs[0].Text)
}

// C12.9 MergeParagraph merges two adjacent paragraphs into one.
func TestMergeParagraph(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello"}}},
	// 	&ir.Paragraph{ID: "p2", Runs: []ir.Run{{Text: " world"}}},
	// }
	// err := ir.MergeParagraph(doc, "p1", "p2")
	// require.NoError(t, err)
	// require.Len(t, doc.Sections[0].Blocks, 1)
	// p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// // Merged paragraph should contain all runs from both originals
	// require.Len(t, p.Runs, 2)
	// assert.Equal(t, "Hello", p.Runs[0].Text)
	// assert.Equal(t, " world", p.Runs[1].Text)
}

// C12.10 InsertBlock can insert a table.
func TestInsertTable(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "before"}}},
	// }
	// tbl := &ir.Table{
	// 	ID: "t1",
	// 	Rows: []ir.TableRow{
	// 		{Cells: []ir.TableCell{
	// 			{ColSpan: 1, RowSpan: 1, Blocks: []ir.Block{
	// 				&ir.Paragraph{ID: "tc1", Runs: []ir.Run{{Text: "cell"}}},
	// 			}},
	// 		}},
	// 	},
	// }
	// err := ir.InsertBlock(doc, 0, 1, tbl)
	// require.NoError(t, err)
	// require.Len(t, doc.Sections[0].Blocks, 2)
	// assert.Equal(t, ir.BlockTable, doc.Sections[0].Blocks[1].BlockType())
}

// C12.11 ChangeStyle changes the style reference on a paragraph.
func TestChangeStyle(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Styles.Named["Heading1"] = ir.Style{
	// 	ID: "heading1", Name: "Heading1",
	// 	RunAttrs: ir.RunAttrs{Bold: true, FontSize: 24},
	// }
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{
	// 		ID:    "p1",
	// 		Style: "Normal",
	// 		Runs:  []ir.Run{{Text: "Title"}},
	// 	},
	// }
	// err := ir.ChangeStyle(doc, "p1", "Heading1")
	// require.NoError(t, err)
	// p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "Heading1", p.Style)
}

// C12.12 Mutation operations mark the document as dirty (ModifiedAt updated).
func TestDirtyTracking(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello"}}},
	// }
	// originalModified := doc.Meta.ModifiedAt
	// time.Sleep(1 * time.Millisecond)
	// err := ir.InsertText(doc, "p1", 5, " world")
	// require.NoError(t, err)
	// assert.True(t, doc.Meta.ModifiedAt.After(originalModified),
	// 	"ModifiedAt should be updated after mutation")
}

// C12.13 EventEmit: mutation operations emit an event describing the change.
func TestEventEmit(t *testing.T) {
	t.Skip("mutation API not implemented")

	// This test would verify that the mutation API emits events
	// that can be observed for collaborative editing or undo/redo.
	//
	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello"}}},
	// }
	// var events []ir.MutationEvent
	// doc.OnMutation(func(e ir.MutationEvent) {
	// 	events = append(events, e)
	// })
	// _ = ir.InsertText(doc, "p1", 5, " world")
	// require.Len(t, events, 1)
	// assert.Equal(t, "insert_text", events[0].Type)
}

// C12.14 UndoSingleOp: undo reverts the last mutation.
func TestUndoSingleOp(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello"}}},
	// }
	// history := ir.NewHistory(doc)
	// _ = history.Do(ir.InsertTextOp("p1", 5, " world"))
	// p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "Hello world", p.Runs[0].Text)
	// _ = history.Undo()
	// p = doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "Hello", p.Runs[0].Text)
}

// C12.15 RedoSingleOp: redo reapplies an undone operation.
func TestRedoSingleOp(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "Hello"}}},
	// }
	// history := ir.NewHistory(doc)
	// _ = history.Do(ir.InsertTextOp("p1", 5, " world"))
	// _ = history.Undo()
	// _ = history.Redo()
	// p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "Hello world", p.Runs[0].Text)
}

// C12.16 UndoStack: multiple undo/redo operations maintain correct stack order.
func TestUndoStack(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: ""}}},
	// }
	// history := ir.NewHistory(doc)
	// _ = history.Do(ir.InsertTextOp("p1", 0, "A"))
	// _ = history.Do(ir.InsertTextOp("p1", 1, "B"))
	// _ = history.Do(ir.InsertTextOp("p1", 2, "C"))
	// p := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "ABC", p.Runs[0].Text)
	//
	// _ = history.Undo() // remove C
	// p = doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "AB", p.Runs[0].Text)
	//
	// _ = history.Undo() // remove B
	// p = doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "A", p.Runs[0].Text)
	//
	// _ = history.Redo() // re-add B
	// p = doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// assert.Equal(t, "AB", p.Runs[0].Text)
}

// C12.17 OpSerialization: mutation operations can be serialized to/from JSON.
func TestOpSerialization(t *testing.T) {
	t.Skip("mutation API not implemented")

	// op := ir.InsertTextOp("p1", 5, " world")
	// data, err := json.Marshal(op)
	// require.NoError(t, err)
	//
	// var restored ir.Operation
	// err = json.Unmarshal(data, &restored)
	// require.NoError(t, err)
	// assert.Equal(t, op.Type(), restored.Type())
	// assert.Equal(t, op.BlockID(), restored.BlockID())
}

// C12.18 ConcurrentOps: concurrent mutations on different blocks do not conflict.
func TestConcurrentOps(t *testing.T) {
	t.Skip("mutation API not implemented")

	// doc := ir.NewDocument()
	// doc.Sections[0].Blocks = []ir.Block{
	// 	&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: "AAA"}}},
	// 	&ir.Paragraph{ID: "p2", Runs: []ir.Run{{Text: "BBB"}}},
	// }
	// var wg sync.WaitGroup
	// wg.Add(2)
	// go func() {
	// 	defer wg.Done()
	// 	_ = ir.InsertText(doc, "p1", 3, "aaa")
	// }()
	// go func() {
	// 	defer wg.Done()
	// 	_ = ir.InsertText(doc, "p2", 3, "bbb")
	// }()
	// wg.Wait()
	// p1 := doc.Sections[0].Blocks[0].(*ir.Paragraph)
	// p2 := doc.Sections[0].Blocks[1].(*ir.Paragraph)
	// assert.Equal(t, "AAAaaa", p1.Runs[0].Text)
	// assert.Equal(t, "BBBbbb", p2.Runs[0].Text)
}
