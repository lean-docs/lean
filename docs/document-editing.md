# Document editing

`ApplyDOCXCommands` applies a batch to a cloned document. A rejected command or unsupported result leaves the input unchanged. `DOCXCommands` lists the operation IDs and their fidelity capabilities.

```go
document := lean.NewDOCX("Meeting notes")
paragraphID := document.Sections[0].Blocks[0].BlockID()
edited, err := lean.ApplyDOCXCommands(document, []lean.DocumentCommand{
    {ID: "paragraph.style", Block: paragraphID, Style: "Title"},
    {ID: "paragraph.insert", At: 1, Text: "First action"},
    {ID: "table.insert", At: 2, Rows: 2, Columns: 3},
})
```

Check the error before exporting the result. `SaveDOCX` refuses canonical saves when the result contains unsupported features.

## Command line

```sh
lean docx commands
lean docx inspect input.docx
lean docx apply input.docx commands.json output.docx
```

The command file contains a JSON array:

```json
[
  {"id": "paragraph.style", "block": "p1", "style": "Heading1"},
  {"id": "paragraph.insert", "at": 1, "text": "New paragraph"}
]
```

The output must not exist. Inspect the input first to obtain its paragraph and table IDs. Inputs with unsupported content remain available for reading but cannot be edited through this command.

## Operations

| Area | Operations |
| --- | --- |
| Text | `text.replace`, `text.format`, `text.link` |
| Paragraphs | `paragraph.insert`, `paragraph.delete`, `paragraph.split`, `paragraph.join`, `paragraph.style`, `paragraph.format` |
| Styles | `style.set`, `style.remove` |
| Numbering | `list.set`, `list.remove` |
| Tables | `table.insert`, `table.delete`, `table.row.insert`, `table.row.delete`, `table.column.insert`, `table.column.delete` |
| Layout | `page.layout`, `page.columns`, `section.insert`, `section.delete`, `section.break` |
| Auxiliary parts | `header.set`, `header.remove`, `footer.set`, `footer.remove` |
| References | `bookmark.set`, `bookmark.remove`, `footnote.set`, `footnote.remove` |

A block target is selected within its section and part. The default section is zero and the default part is the document body. Header/footer commands can select default, first or even variants. Nested table cell blocks retain their IDs and can be targeted directly.

Text ranges use UTF-16 offsets. A supplementary character counts as two units and an inline break counts as one. Ranges cannot split a supplementary character. Paragraph insertion uses a block index; table row and column commands use a row or column index.

All layout measurements use points. Page margins must leave a content area. Column widths and gaps must fit that area. Sections retain continuous, next-page, odd-page, even-page and next-column breaks, plus header/footer distances and gutter margins. Structural table commands refuse merged-cell changes.

Formatting setters distinguish inherited values from explicit off, zero or black values. Named paragraph styles resolve their parent chain; cyclic definitions and removing a style still in use are rejected.

Bookmark and footnote anchors follow text replacements and paragraph splits and joins. Footnote offsets are optional; an omitted offset places the reference at the paragraph end. Existing bookmark ranges must have valid paragraph anchors and cannot split a supplementary character.

Fields, tracked changes, character-style references, arbitrary table styles and row spans still require additional preservation support. Applications must inspect the fidelity report rather than infer editability from a successful parse.
