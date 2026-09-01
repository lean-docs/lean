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
		"<w:numPr":   "numbering",
		"<w:pStyle":  "styles",
		"<w:sdt":     "content-controls",
		"<w:fldChar": "fields",
		"<w:vMerge":  "row-spans",
	}
	for marker, feature := range markers {
		if strings.Contains(documentXML, marker) {
			report.addUnsupported(feature)
		}
	}
	if hasElementWithin(documentXML, "rPr", "shd") {
		report.addUnsupported("run-shading")
	}
	if hasElementWithin(documentXML, "body", "ins") || hasElementWithin(documentXML, "body", "del") {
		report.addUnsupported("tracked-changes")
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
			if len(value.Data) == 0 || value.Width <= 0 || value.Height <= 0 {
				report.addUnsupported("images")
			}
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
			for _, run := range value.Runs {
				if run.Attrs.Highlight != (ir.Color{}) && !supportedHighlight(run.Attrs.Highlight) {
					report.addUnsupported("highlighting")
				}
			}
		case *ir.Table:
			if value.Style != "" && value.Style != "TableNormal" {
				report.addUnsupported("table-styles")
			}
			for _, row := range value.Rows {
				for _, cell := range row.Cells {
					if cell.RowSpan > 1 {
						report.addUnsupported("row-spans")
					}
					walkDocumentBlocks(cell.Blocks, report)
				}
			}
		}
	}
}

func supportedHighlight(color ir.Color) bool {
	for _, supported := range []ir.Color{
		{R: 255, G: 255, B: 0}, {R: 0, G: 255, B: 0}, {R: 0, G: 255, B: 255},
		{R: 255, G: 0, B: 255}, {R: 0, G: 0, B: 255}, {R: 255, G: 0, B: 0},
		{R: 0, G: 0, B: 0}, {R: 255, G: 255, B: 255}, {R: 0, G: 0, B: 139},
		{R: 0, G: 139, B: 139}, {R: 0, G: 100, B: 0}, {R: 139, G: 0, B: 139},
		{R: 139, G: 0, B: 0}, {R: 128, G: 128, B: 0}, {R: 169, G: 169, B: 169},
		{R: 211, G: 211, B: 211},
	} {
		if color == supported {
			return true
		}
	}
	return false
}

func (report *FidelityReport) addUnsupported(feature string) {
	if !slices.Contains(report.Unsupported, feature) {
		report.Unsupported = append(report.Unsupported, feature)
	}
}
