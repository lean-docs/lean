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
	"errors"
	"fmt"
	"io"

	"github.com/lean-docs/lean/pkg/ir"
)

var (
	ErrEmptyInput        = errors.New("ooxml: empty input")
	ErrInvalidZip        = errors.New("ooxml: not a valid .docx (zip) archive")
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
	doc.Sections = doc.Sections[:0]
	if err := parseDocument(docXML, doc); err != nil {
		return nil, err
	}
	if len(doc.Sections) == 0 {
		doc.Sections = append(doc.Sections, ir.Section{})
	}
	return doc, nil
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
		defer rc.Close()
		return io.ReadAll(rc)
	}
	if name == "word/document.xml" {
		return nil, ErrMissingDocumentXML
	}
	return nil, fmt.Errorf("ooxml: %s missing", name)
}
