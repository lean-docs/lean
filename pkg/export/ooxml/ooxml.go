// Package ooxml exports an IR document to .docx (OOXML) format.
package ooxml

import (
	"errors"

	"github.com/lean-docs/lean/pkg/ir"
)

// ErrNotImplemented is returned when the exporter is not yet implemented.
var ErrNotImplemented = errors.New("ooxml exporter: not implemented")

// Export exports an IR Document to .docx bytes.
func Export(doc *ir.Document) ([]byte, error) {
	return nil, ErrNotImplemented
}
