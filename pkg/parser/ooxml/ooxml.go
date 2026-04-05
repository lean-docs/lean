// Package ooxml provides a .docx (OOXML) parser that produces an IR document.
package ooxml

import (
	"errors"

	"github.com/lean-docs/lean/pkg/ir"
)

// ErrNotImplemented is returned when the parser is not yet implemented.
var ErrNotImplemented = errors.New("ooxml parser: not implemented")

// Parse parses .docx bytes into an IR Document.
func Parse(input []byte) (*ir.Document, error) {
	return nil, ErrNotImplemented
}
