// Package html provides an HTML parser that produces an IR document.
package html

import (
	"errors"

	"github.com/lean-docs/lean/pkg/ir"
)

// ErrNotImplemented is returned when the parser is not yet implemented.
var ErrNotImplemented = errors.New("html parser: not implemented")

// Parse parses HTML bytes into an IR Document.
func Parse(input []byte) (*ir.Document, error) {
	return nil, ErrNotImplemented
}
