package lean

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"math"
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

func NewDOCX(title string) *ir.Document {
	document := ir.NewDocument()
	document.Meta.Title = title
	for _, id := range []string{"Normal", "Title", "Subtitle", "Heading1", "Heading2", "Heading3", "Heading4", "Heading5", "Heading6"} {
		style, _ := ir.BuiltinStyle(id)
		document.Styles.Named[id] = style
	}
	document.Sections[0].Blocks = []ir.Block{&ir.Paragraph{ID: "p1", Runs: []ir.Run{{Text: title}}}}
	return document
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
	validation := inspectDocument(document)
	for _, feature := range validation.Unsupported {
		report.addUnsupported(feature)
	}
	report.Editable = len(report.Unsupported) == 0
	slices.Sort(report.Unsupported)
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
	_, reopened, err := OpenDOCX(content)
	if err != nil {
		return nil, report, err
	}
	if !reopened.Editable {
		return nil, reopened, ErrUnsupportedDocument
	}
	return content, report, nil
}

func inspectDOCX(input []byte) (FidelityReport, error) {
	report := FidelityReport{Format: "docx", Preserved: []string{"text", "run-formatting", "explicit-run-toggles", "explicit-run-values", "underline-variants", "paragraph-formatting", "explicit-paragraph-values", "explicit-run-colour", "lists", "page-layout", "sections", "headers", "footers", "footnotes", "page-columns", "list-restarts", "bookmarks", "hyperlinks", "tables", "images", "drawing-anchors", "shapes", "text-boxes"}}
	reader, err := zip.NewReader(bytes.NewReader(input), int64(len(input)))
	if err != nil {
		return report, err
	}
	var documentXML string
	var auxiliaryXML strings.Builder
	for _, file := range reader.File {
		switch {
		case file.Name == "word/document.xml" || file.Name == "word/styles.xml" || strings.HasPrefix(file.Name, "word/header") || strings.HasPrefix(file.Name, "word/footer") || file.Name == "word/footnotes.xml":
			stream, openErr := file.Open()
			if openErr != nil {
				return report, openErr
			}
			content, readErr := io.ReadAll(stream)
			_ = stream.Close()
			if readErr != nil {
				return report, readErr
			}
			inspectUnderlineProperties(content, &report)
			inspectDrawingFeatures(content, &report)
			auxiliaryXML.Write(content)
			if file.Name == "word/document.xml" {
				documentXML = string(content)
			}
		case file.Name == "word/comments.xml":
			report.addUnsupported("comments")
		}
	}
	markers := map[string]string{
		"<w:sdt":     "content-controls",
		"<w:fldChar": "fields",
		"<w:vMerge":  "row-spans",
	}
	for marker, feature := range markers {
		if strings.Contains(auxiliaryXML.String(), marker) {
			report.addUnsupported(feature)
		}
	}
	if hasElementWithin(auxiliaryXML.String(), "rPr", "shd") {
		report.addUnsupported("run-shading")
	}
	if hasElementWithin(auxiliaryXML.String(), "rPr", "rStyle") {
		report.addUnsupported("character-style-references")
	}
	if hasElementWithin(documentXML, "body", "ins") || hasElementWithin(documentXML, "body", "del") {
		report.addUnsupported("tracked-changes")
	}
	if hasNonDefaultTableStyle(auxiliaryXML.String()) {
		report.addUnsupported("table-styles")
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

func inspectDocument(document *ir.Document) FidelityReport {
	report := FidelityReport{Format: "docx", Editable: document != nil, Preserved: []string{"text", "run-formatting", "explicit-run-toggles", "explicit-run-values", "underline-variants", "paragraph-formatting", "explicit-paragraph-values", "explicit-run-colour", "lists", "page-layout", "sections", "headers", "footers", "footnotes", "page-columns", "list-restarts", "bookmarks", "hyperlinks", "tables", "images", "drawing-anchors", "shapes", "text-boxes"}}
	if document == nil {
		report.addUnsupported("document")
		return report
	}
	inspectDocumentStructure(document, &report)
	inspectRunValues(document.Styles.Defaults, &report)
	for _, style := range document.Styles.Named {
		inspectRunValues(style.RunAttrs, &report)
	}
	for _, section := range document.Sections {
		for _, column := range section.Columns {
			if column.Width <= 0 || column.Spacing < 0 || !finite(column.Width) || !finite(column.Spacing) {
				report.addUnsupported("page-columns")
			}
		}
		for _, group := range [][]*ir.HeaderFooter{append([]*ir.HeaderFooter{section.Header}, section.Headers...), append([]*ir.HeaderFooter{section.Footer}, section.Footers...)} {
			seen := map[ir.HeaderFooterKind]bool{}
			for _, part := range group {
				if part == nil {
					continue
				}
				if part.Kind < ir.HeaderFooterDefault || part.Kind > ir.HeaderFooterEven || seen[part.Kind] {
					report.addUnsupported("header-footer-variants")
					continue
				}
				seen[part.Kind] = true
				walkDocumentBlocks(part.Blocks, &report)
			}
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
		case *ir.Shape:
			if !finite(value.Width) || !finite(value.Height) || value.Width <= 0 || value.Height <= 0 || !slices.Contains([]string{"rect", "roundRect", "ellipse", "line"}, value.Geometry) {
				report.addUnsupported("shape-geometry")
			}
			walkDocumentBlocks(value.Blocks, report)
		case *ir.Image:
			if len(value.Data) == 0 || !finite(value.Width) || !finite(value.Height) || value.Width <= 0 || value.Height <= 0 || value.Format < ir.ImagePNG || value.Format > ir.ImageSVG {
				report.addUnsupported("images")
			}
		case *ir.Bookmark:
			if value.Start != nil && (value.End == nil || value.Start.Offset < 0 || value.End.Offset < 0) {
				report.addUnsupported("bookmark-ranges")
			}
		case *ir.Paragraph:
			for _, note := range value.Footnotes {
				walkDocumentBlocks(note.Blocks, report)
			}
			for _, run := range value.Runs {
				inspectRunValues(run.Attrs, report)
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

func inspectRunValues(attrs ir.RunAttrs, report *FidelityReport) {
	if attrs.Underline < ir.UnderlineNone || attrs.Underline > ir.UnderlineWavyDouble {
		report.addUnsupported("underline-styles")
	}
	if math.IsNaN(attrs.Tracking) || math.IsInf(attrs.Tracking, 0) {
		report.addUnsupported("character-spacing")
	}
}

func inspectUnderlineProperties(content []byte, report *FidelityReport) {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	for {
		token, err := decoder.Token()
		if err != nil {
			return
		}
		element, ok := token.(xml.StartElement)
		if !ok || element.Name.Local != "u" {
			continue
		}
		for _, attribute := range element.Attr {
			if attribute.Name.Space == "xmlns" || attribute.Name.Local == "xmlns" {
				continue
			}
			if attribute.Name.Local != "val" {
				report.addUnsupported("underline-properties")
				continue
			}
			switch attribute.Value {
			case "", "none", "single", "double", "wave", "wavy", "dash", "words", "thick", "dotted", "dottedHeavy", "dashedHeavy", "dashLong", "dashLongHeavy", "dotDash", "dashDotHeavy", "dotDotDash", "dashDotDotHeavy", "wavyHeavy", "wavyDouble":
			default:
				report.addUnsupported("underline-styles")
			}
		}
	}
}
