// Package md provides a CommonMark Markdown parser that produces an IR document.
package md

import "github.com/lean-docs/lean/pkg/ir"

import "errors"

// ErrNotImplemented is returned when the parser is not yet implemented.
var ErrNotImplemented = errors.New("md parser: not implemented")

// Parse parses CommonMark Markdown bytes into an IR Document.
func Parse(input []byte) (*ir.Document, error) {
	return nil, ErrNotImplemented
}
