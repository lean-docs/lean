package lean

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"slices"
	"strings"

	ooxmlexport "github.com/lean-docs/lean/pkg/export/ooxml"
	"github.com/lean-docs/lean/pkg/ir"
	ooxmlparser "github.com/lean-docs/lean/pkg/parser/ooxml"
)

var ErrUnsupportedDocument = errors.New("lean: document contains unsupported features")

type FidelityReport struct {
	Format      string   `json:"format"`
	Editable    bool     `json:"editable"`
	Preserved   []string `json:"preserved"`
	Unsupported []string `json:"unsupported"`
}

func OpenDOCX(input []byte) (*ir.Document, FidelityReport, error) {
	report, err := inspectDOCX(input)
	if err != nil {
		return nil, report, err
	}
	document, err := ooxmlparser.Parse(input)
	if err != nil {
		return nil, report, err
	}
	return document, report, nil
}

func SaveDOCX(document *ir.Document) ([]byte, FidelityReport, error) {
	report := inspectDocument(document)
	if !report.Editable {
		return nil, report, ErrUnsupportedDocument
	}
	content, err := ooxmlexport.Export(document)
	if err != nil {
		return nil, report, err
	}
	if _, _, err := OpenDOCX(content); err != nil {
		return nil, report, err
	}
	return content, report, nil
}

func inspectDOCX(input []byte) (FidelityReport, error) {
	report := FidelityReport{Format: "docx", Preserved: []string{"text", "run-formatting", "paragraph-formatting", "tables"}}
	reader, err := zip.NewReader(bytes.NewReader(input), int64(len(input)))
	if err != nil {
		return report, err
	}
	var documentXML string
	for _, file := range reader.File {
		switch {
		case file.Name == "word/document.xml":
			stream, openErr := file.Open()
			if openErr != nil {
				return report, openErr
			}
			content, readErr := io.ReadAll(stream)
			_ = stream.Close()
			if readErr != nil {
				return report, readErr
			}
			documentXML = string(content)
		case strings.HasPrefix(file.Name, "word/header"):
			report.addUnsupported("headers")
		case strings.HasPrefix(file.Name, "word/footer"):
			report.addUnsupported("footers")
		case file.Name == "word/footnotes.xml":
			report.addUnsupported("footnotes")
		case file.Name == "word/comments.xml":
			report.addUnsupported("comments")
		}
	}
	markers := map[string]string{
		"<w:drawing":    "images",
		"<w:hyperlink":  "hyperlinks",
		"<w:numPr":      "numbering",
		"<w:pStyle":     "styles",
		"<w:ins":        "tracked-changes",
		"<w:del":        "tracked-changes",
		"<w:sdt":        "content-controls",
		"<w:fldChar":    "fields",
		"<w:highlight":  "highlighting",
		"<w:lang":       "run-language",
		"<w:pBdr":       "paragraph-borders",
		"<w:shd":        "shading",
		"<w:bidi":       "bidirectional-text",
		"<w:outlineLvl": "outline-levels",
		"<w:tblBorders": "table-borders",
		"<w:tcBorders":  "table-borders",
		"<w:tcMar":      "cell-padding",
		"<w:vMerge":     "row-spans",
		"<w:vAlign":     "cell-alignment",
		"<w:trHeight":   "row-heights",
		"<w:tblHeader":  "header-rows",
		"w:orient=":     "page-orientation",
	}
	for marker, feature := range markers {
		if strings.Contains(documentXML, marker) {
			report.addUnsupported(feature)
		}
	}
	if hasElementWithin(documentXML, "rPr", "spacing") {
		report.addUnsupported("character-spacing")
	}
	if hasNonDefaultTableStyle(documentXML) {
		report.addUnsupported("table-styles")
	}
	if hasNonDefaultColumns(documentXML) {
		report.addUnsupported("page-columns")
	}
	report.Editable = len(report.Unsupported) == 0
	slices.Sort(report.Unsupported)
	return report, nil
}

func hasElementWithin(documentXML, parent, child string) bool {
	decoder := xml.NewDecoder(strings.NewReader(documentXML))
	depth := 0
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == parent {
				depth++
			} else if depth > 0 && element.Name.Local == child {
				return true
			}
		case xml.EndElement:
			if element.Name.Local == parent {
				depth--
			}
		}
	}
}

func hasNonDefaultTableStyle(documentXML string) bool {
	decoder := xml.NewDecoder(strings.NewReader(documentXML))
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		element, ok := token.(xml.StartElement)
		if !ok || element.Name.Local != "tblStyle" {
			continue
		}
		for _, attribute := range element.Attr {
			if attribute.Name.Local == "val" && attribute.Value != "TableNormal" {
				return true
			}
		}
	}
}

func hasNonDefaultColumns(documentXML string) bool {
	decoder := xml.NewDecoder(strings.NewReader(documentXML))
	insideColumns := false
	for {
		token, err := decoder.Token()
		if err != nil {
			return false
		}
		switch element := token.(type) {
		case xml.StartElement:
			if element.Name.Local == "cols" {
				insideColumns = true
				for _, attribute := range element.Attr {
					if attribute.Name.Local == "num" && attribute.Value != "" && attribute.Value != "1" {
						return true
					}
				}
			} else if insideColumns && element.Name.Local == "col" {
				return true
			}
		case xml.EndElement:
			if element.Name.Local == "cols" {
				insideColumns = false
			}
		}
	}
}

func inspectDocument(document *ir.Document) FidelityReport {
	report := FidelityReport{Format: "docx", Editable: document != nil, Preserved: []string{"text", "run-formatting", "paragraph-formatting", "tables"}}
	if document == nil {
		report.addUnsupported("document")
		return report
	}
	if len(document.Styles.Named) > 0 {
		report.addUnsupported("styles")
	}
	if len(document.Styles.Numbering) > 0 {
		report.addUnsupported("numbering")
	}
	for _, section := range document.Sections {
		if section.Properties.Orientation != ir.Portrait {
			report.addUnsupported("page-orientation")
		}
		if len(section.Columns) > 0 {
			report.addUnsupported("page-columns")
		}
		if section.Header != nil {
			report.addUnsupported("headers")
		}
		if section.Footer != nil {
			report.addUnsupported("footers")
		}
		walkDocumentBlocks(section.Blocks, &report)
	}
	report.Editable = len(report.Unsupported) == 0
	slices.Sort(report.Unsupported)
	return report
}

func walkDocumentBlocks(blocks []ir.Block, report *FidelityReport) {
	for _, block := range blocks {
		switch value := block.(type) {
		case *ir.Image:
			report.addUnsupported("images")
		case *ir.Bookmark:
			report.addUnsupported("bookmarks")
		case *ir.Paragraph:
			if value.Style != "" {
				report.addUnsupported("styles")
			}
			if value.Numbering != nil {
				report.addUnsupported("numbering")
			}
			if len(value.Footnotes) > 0 {
				report.addUnsupported("footnotes")
			}
			if value.Para.OutlineLevel != 0 {
				report.addUnsupported("outline-levels")
			}
			if value.Para.Borders != (ir.ParaBorders{}) {
				report.addUnsupported("paragraph-borders")
			}
			if value.Para.Shading != (ir.Color{}) {
				report.addUnsupported("shading")
			}
			if value.Para.BiDi {
				report.addUnsupported("bidirectional-text")
			}
			for _, run := range value.Runs {
				if run.Hyperlink != nil {
					report.addUnsupported("hyperlinks")
				}
				if run.Attrs.Highlight != (ir.Color{}) {
					report.addUnsupported("highlighting")
				}
				if run.Attrs.Tracking != 0 {
					report.addUnsupported("character-spacing")
				}
				if run.Attrs.Language != "" {
					report.addUnsupported("run-language")
				}
			}
		case *ir.Table:
			if value.Style != "" && value.Style != "TableNormal" {
				report.addUnsupported("table-styles")
			}
			if value.Align != ir.AlignLeft || value.Borders != (ir.TableBorders{}) || value.Shading != (ir.Color{}) {
				report.addUnsupported("table-formatting")
			}
			for _, row := range value.Rows {
				if row.IsHeader {
					report.addUnsupported("header-rows")
				}
				if row.Height != nil {
					report.addUnsupported("row-heights")
				}
				for _, cell := range row.Cells {
					if cell.RowSpan > 1 {
						report.addUnsupported("row-spans")
					}
					if cell.Borders != (ir.CellBorders{}) || cell.Shading != (ir.Color{}) || cell.Padding != (ir.Padding{}) {
						report.addUnsupported("cell-formatting")
					}
					if cell.VAlign != ir.VAlignTop {
						report.addUnsupported("cell-alignment")
					}
					walkDocumentBlocks(cell.Blocks, report)
				}
			}
		}
	}
}

func (report *FidelityReport) addUnsupported(feature string) {
	if !slices.Contains(report.Unsupported, feature) {
		report.Unsupported = append(report.Unsupported, feature)
	}
}
