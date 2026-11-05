# Document authoring commands

`table.format` applies optional `table_format` properties to an existing table. Column widths require one point measurement per logical column. Alignment, padding, vertical alignment, shading and borders retain the canonical IR types. Cell properties apply to every cell. `repeat_header` marks only the first row. Omitted properties remain unchanged. Pending tracked changes must be resolved first. Command batches are atomic and use normal DOCX validation and export.

`drawing.layout` accepts optional `drawing.alt` to update a picture's alternative text without changing image bytes or its anchor. Existing drawing dimensions and anchor fields remain optional.

`table.merge` merges an unmerged rectangle selected by zero-based `cell_range.row`, `column`, `rows` and `columns`. All selected cell contents move to the origin in row order. Split existing merged cells before merging another range in that table. `table.split` selects a merged origin by row and logical column; its contents stay in the first resulting cell. Both commands reject pending tracked changes. Vertical spans retain continuation cells in canonical IR and export as DOCX vertical merges. Other export formats may flatten the table geometry. Table style selection remains a separate operation.

Set `table_format.reset_style` to true to clear direct cell borders, shading, padding and alignment and direct table borders/shading. The named style and column widths remain unchanged. Explicit properties in the same command apply after the reset.

`drawing.layout` can replace picture bytes through `drawing.data` (base64 in JSON) and `drawing.format`. Supply both together; the object identity, dimensions, alternative text and anchor remain unchanged unless explicitly updated in the same command. Replacement also supports images nested inside tables or text boxes.
