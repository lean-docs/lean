// Package ooxml parses .docx (OOXML) files into Lean's IR.
//
// Scope of this initial implementation: ZIP container, document.xml body
// walk, paragraphs, runs, and run-level formatting (bold, italic, underline,
// strike, small/all caps, baseline shift, font size/name, run color,
// highlight). Paragraph-level properties, tables, images, lists, styles,
// sections, headers/footers, and error-recovery for malformed XML are
// follow-on work.
package ooxml

import (
	"archive/zip"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/lean-docs/lean/pkg/ir"
)

var (
	ErrEmptyInput         = errors.New("ooxml: empty input")
	ErrInvalidZip         = errors.New("ooxml: not a valid .docx (zip) archive")
	ErrMissingDocumentXML = errors.New("ooxml: word/document.xml missing")
)

// Parse parses .docx bytes into an IR Document.
func Parse(input []byte) (*ir.Document, error) {
	if len(input) == 0 {
		return nil, ErrEmptyInput
	}

	r, err := zip.NewReader(bytes.NewReader(input), int64(len(input)))
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidZip, err)
	}

	docXML, err := readEntry(r, "word/document.xml")
	if err != nil {
		return nil, err
	}

	doc := ir.NewDocument()
	if coreXML, coreErr := readOptionalEntry(r, "docProps/core.xml"); coreErr != nil {
		return nil, coreErr
	} else if len(coreXML) > 0 {
		parseCoreProperties(coreXML, doc)
	}
	doc.Sections = doc.Sections[:0]
	images, err := readImages(r)
	if err != nil {
		return nil, err
	}
	if err := parseDocument(docXML, doc, images); err != nil {
		return nil, err
	}
	if len(doc.Sections) == 0 {
		doc.Sections = append(doc.Sections, ir.Section{})
	}
	return doc, nil
}

type xmlRelationships struct {
	Relationships []xmlRelationship `xml:"Relationship"`
}

type xmlRelationship struct {
	ID     string `xml:"Id,attr"`
	Type   string `xml:"Type,attr"`
	Target string `xml:"Target,attr"`
}

func readImages(reader *zip.Reader) (map[string]imageResource, error) {
	content, err := readOptionalEntry(reader, "word/_rels/document.xml.rels")
	if err != nil || len(content) == 0 {
		return nil, err
	}
	var relationships xmlRelationships
	if err := xml.Unmarshal(content, &relationships); err != nil {
		return nil, fmt.Errorf("ooxml: parse document relationships: %w", err)
	}
	images := make(map[string]imageResource)
	for _, relationship := range relationships.Relationships {
		if !strings.HasSuffix(relationship.Type, "/image") {
			continue
		}
		target := path.Clean(path.Join("word", relationship.Target))
		data, readErr := readOptionalEntry(reader, target)
		if readErr != nil {
			return nil, readErr
		}
		if len(data) == 0 {
			continue
		}
		images[relationship.ID] = imageResource{Data: data, Format: imageFormat(target)}
	}
	return images, nil
}

func imageFormat(name string) ir.ImageFormat {
	switch strings.ToLower(path.Ext(name)) {
	case ".jpg", ".jpeg":
		return ir.ImageJPEG
	case ".gif":
		return ir.ImageGIF
	case ".svg":
		return ir.ImageSVG
	default:
		return ir.ImagePNG
	}
}

func readOptionalEntry(r *zip.Reader, name string) ([]byte, error) {
	for _, file := range r.File {
		if file.Name != name {
			continue
		}
		stream, err := file.Open()
		if err != nil {
			return nil, fmt.Errorf("ooxml: open %s: %w", name, err)
		}
		content, err := io.ReadAll(stream)
		_ = stream.Close()
		return content, err
	}
	return nil, nil
}

func parseCoreProperties(content []byte, document *ir.Document) {
	decoder := xml.NewDecoder(bytes.NewReader(content))
	for {
		token, err := decoder.Token()
		if err != nil {
			return
		}
		element, ok := token.(xml.StartElement)
		if !ok || (element.Name.Local != "title" && element.Name.Local != "creator") {
			continue
		}
		var value string
		if decoder.DecodeElement(&value, &element) != nil {
			return
		}
		if element.Name.Local == "title" {
			document.Meta.Title = value
		} else {
			document.Meta.Author = value
		}
	}
}

func readEntry(r *zip.Reader, name string) ([]byte, error) {
	for _, f := range r.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		if err != nil {
			return nil, fmt.Errorf("ooxml: open %s: %w", name, err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		return data, err
	}
	if name == "word/document.xml" {
		return nil, ErrMissingDocumentXML
	}
	return nil, fmt.Errorf("ooxml: %s missing", name)
}
