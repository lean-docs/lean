// Package typst exports an IR document to Typst markup.
package typst

import (
	"errors"

	"github.com/lean-docs/lean/pkg/ir"
)

// ErrNotImplemented is returned when the exporter is not yet implemented.
var ErrNotImplemented = errors.New("typst exporter: not implemented")

// Export exports an IR Document to Typst markup bytes.
func Export(doc *ir.Document) ([]byte, error) {
	return nil, ErrNotImplemented
}
