# Tracked document changes

The canonical DOCX model preserves pending text insertions, text deletions, character-formatting revisions, and table-row insertions and deletions. Revision metadata includes an integer ID, type, author, and date. Unsupported revision structures appear in the fidelity report and prevent a lossy save.

## Structured proposals

`text.suggest` proposes a replacement within one paragraph using UTF-16 start and end offsets. The previous text remains a deletion and the replacement becomes an insertion. Fields, links, and manual breaks are excluded from tracked replacements.

`text.format.suggest` preserves the previous direct character properties and character style while proposing new formatting. Accepting keeps the proposed formatting; rejecting restores the previous properties.

`table.insert.suggest` creates a table with inserted-row revisions. `table.row.insert.suggest` and `table.row.delete.suggest` propose row changes in an existing table without merged cells. Each row revision is resolved independently. Rejecting every row of a proposed table removes the empty table.

Hosts supply author identity and an RFC3339 date from authenticated context. The engine assigns revision IDs. `change.accept` and `change.reject` resolve the supplied `change_id` across the document. Normal structural commands must not discard unresolved row revisions.

`DOCXCommands` returns capability requirements and parameter schemas. Hosts must also check document fidelity and resource permissions before offering an operation.

## DOCX representation

Text uses WordprocessingML insertion/deletion wrappers, with deleted text stored as `delText`. Character formatting uses `rPrChange` with the prior run properties. Row revisions use insertion/deletion markers within `trPr`.

Pending revisions survive DOCX export and reopening. Move revisions, column/cell structural revisions, paragraph-property history, and nested revision histories are outside the supported subset. These structures must remain fidelity failures rather than being dropped.

## Table design and media

Table creation can select a defined table style; `table.style` changes an existing table's style. Named table styles preserve conditional regions and formatting. Image watermarks retain embedded PNG, JPEG, or GIF data, opacity, dimensions, alignment, reference area, and offsets. Text watermarks retain their text, font, color, rotation, opacity, and placement.

## Standard references

- [Word inserted runs](https://learn.microsoft.com/en-us/dotnet/api/documentformat.openxml.wordprocessing.insertedrun)
- [Word row revisions](https://learn.microsoft.com/en-us/office/open-xml/word/how-to-accept-all-revisions-in-a-word-processing-document)
