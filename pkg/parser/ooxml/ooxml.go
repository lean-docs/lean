// Package ooxml parses .docx files into Lean's intermediate representation.
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
	styleNumbering := make(map[string]*ir.NumberingRef)
	if coreXML, coreErr := readOptionalEntry(r, "docProps/core.xml"); coreErr != nil {
		return nil, coreErr
	} else if len(coreXML) > 0 {
		parseCoreProperties(coreXML, doc)
	}
	if numberingXML, numberingErr := readOptionalEntry(r, "word/numbering.xml"); numberingErr != nil {
		return nil, numberingErr
	} else if len(numberingXML) > 0 {
		if err := parseNumbering(numberingXML, doc); err != nil {
			return nil, fmt.Errorf("ooxml: parse numbering: %w", err)
		}
	}
	if stylesXML, stylesErr := readOptionalEntry(r, "word/styles.xml"); stylesErr != nil {
		return nil, stylesErr
	} else if len(stylesXML) > 0 {
		if err := parseStyles(stylesXML, doc); err != nil {
			return nil, fmt.Errorf("ooxml: parse styles: %w", err)
		}
		styleNumbering, err = parseStyleNumbering(stylesXML)
		if err != nil {
			return nil, fmt.Errorf("ooxml: parse style numbering: %w", err)
		}
	}
	doc.Sections = doc.Sections[:0]
	footnotes := make(map[string]ir.Footnote)
	if footnotesXML, footnotesErr := readOptionalEntry(r, "word/footnotes.xml"); footnotesErr != nil {
		return nil, footnotesErr
	} else if len(footnotesXML) > 0 {
		footnotes, err = parseFootnotes(footnotesXML)
		if err != nil {
			return nil, fmt.Errorf("ooxml: parse footnotes: %w", err)
		}
	}
	images, hyperlinks, headers, footers, err := readRelationships(r)
	if err != nil {
		return nil, err
	}
	if err := parseDocument(docXML, doc, images, hyperlinks, headers, footers, styleNumbering, footnotes); err != nil {
		return nil, err
	}
	if len(doc.Sections) == 0 {
		doc.Sections = append(doc.Sections, ir.Section{})
	}
	return doc, nil
}

type xmlFootnotes struct {
	Notes []xmlFootnote `xml:"footnote"`
}

type xmlFootnote struct {
	ID         string         `xml:"id,attr"`
	Type       string         `xml:"type,attr"`
	Paragraphs []xmlParagraph `xml:"p"`
}

func parseFootnotes(content []byte) (map[string]ir.Footnote, error) {
	var source xmlFootnotes
	if err := xml.Unmarshal(content, &source); err != nil {
		return nil, err
	}
	result := make(map[string]ir.Footnote)
	context := &parseCtx{}
	for _, note := range source.Notes {
		if note.Type != "" || note.ID == "" {
			continue
		}
		footnote := ir.Footnote{ID: note.ID}
		for _, paragraph := range note.Paragraphs {
			footnote.Blocks = append(footnote.Blocks, convertParagraphBlocks(paragraph, context)...)
		}
		result[note.ID] = footnote
	}
	return result, nil
}

type xmlRelationships struct {
	Relationships []xmlRelationship `xml:"Relationship"`
}

type xmlRelationship struct {
	ID         string `xml:"Id,attr"`
	Type       string `xml:"Type,attr"`
	Target     string `xml:"Target,attr"`
	TargetMode string `xml:"TargetMode,attr"`
}

func readRelationships(reader *zip.Reader) (map[string]imageResource, map[string]string, map[string]*ir.HeaderFooter, map[string]*ir.HeaderFooter, error) {
	content, err := readOptionalEntry(reader, "word/_rels/document.xml.rels")
	if err != nil || len(content) == 0 {
		return nil, nil, nil, nil, err
	}
	var relationships xmlRelationships
	if err := xml.Unmarshal(content, &relationships); err != nil {
		return nil, nil, nil, nil, fmt.Errorf("ooxml: parse document relationships: %w", err)
	}
	images := make(map[string]imageResource)
	hyperlinks := make(map[string]string)
	headers := make(map[string]*ir.HeaderFooter)
	footers := make(map[string]*ir.HeaderFooter)
	for _, relationship := range relationships.Relationships {
		if strings.HasSuffix(relationship.Type, "/hyperlink") && relationship.TargetMode == "External" {
			hyperlinks[relationship.ID] = relationship.Target
			continue
		}
		if strings.HasSuffix(relationship.Type, "/header") || strings.HasSuffix(relationship.Type, "/footer") {
			target := path.Clean(path.Join("word", relationship.Target))
			data, readErr := readOptionalEntry(reader, target)
			if readErr != nil {
				return nil, nil, nil, nil, readErr
			}
			root := "hdr"
			destination := headers
			if strings.HasSuffix(relationship.Type, "/footer") {
				root = "ftr"
				destination = footers
			}
			value, parseErr := parseHeaderFooter(data, root, ir.HeaderFooterDefault)
			if parseErr != nil {
				return nil, nil, nil, nil, parseErr
			}
			destination[relationship.ID] = value
			continue
		}
		if !strings.HasSuffix(relationship.Type, "/image") {
			continue
		}
		target := path.Clean(path.Join("word", relationship.Target))
		data, readErr := readOptionalEntry(reader, target)
		if readErr != nil {
			return nil, nil, nil, nil, readErr
		}
		if len(data) == 0 {
			continue
		}
		images[relationship.ID] = imageResource{Data: data, Format: imageFormat(target)}
	}
	return images, hyperlinks, headers, footers, nil
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
