// Package md exports an IR document to CommonMark Markdown.
package md

import (
	"errors"

	"github.com/lean-docs/lean/pkg/ir"
)

// ErrNotImplemented is returned when the exporter is not yet implemented.
var ErrNotImplemented = errors.New("md exporter: not implemented")

// Export exports an IR Document to Markdown bytes.
func Export(doc *ir.Document) ([]byte, error) {
	return nil, ErrNotImplemented
}
