# Document authoring commands

`table.format` applies optional `table_format` properties to an existing table. Column widths require one point measurement per logical column. Alignment, padding, vertical alignment, shading and borders retain the canonical IR types. Cell properties apply to every cell. `repeat_header` marks only the first row. Omitted properties remain unchanged. Pending tracked changes must be resolved first. Command batches are atomic and use normal DOCX validation and export.

`drawing.layout` accepts optional `drawing.alt` to update a picture's alternative text without changing image bytes or its anchor. Existing drawing dimensions and anchor fields remain optional.

Vertical merging and cell splitting are not added by these commands. Table style selection remains a separate operation.
